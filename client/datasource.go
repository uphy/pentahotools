package client

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strconv"

	"github.com/pkg/errors"

	resty "gopkg.in/resty.v0"

	"archive/zip"

	"strings"

	"go.uber.org/zap"
)

// ListAnalysisDatasources lists the analysis datasources.
func (c *Client) ListAnalysisDatasources() ([]string, error) {
	c.Logger.Debug("ListAnalysisDatasources")
	var result DatasourceCatalog
	resp, err := c.client.R().
		SetHeader("Accept", "application/json").
		SetResult(&result).
		Get("plugin/data-access/api/datasource/analysis/catalog")
	switch resp.StatusCode() {
	case 200:
		return result.getItemNames()
	default:
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("Unknown error. statusCode=%d", resp.StatusCode())
	}
}

// AnalysisDatasourceInfo represents analysis datasource info
type AnalysisDatasourceInfo struct {
	DataSource string
	Provider   string
	EnableXmla bool
	Overwrite  bool
}

// Print prints the analysis datasource info
func (a *AnalysisDatasourceInfo) Print() {
	fmt.Printf("Datasource : %s\n", a.DataSource)
	fmt.Printf("Provider   : %s\n", a.Provider)
	fmt.Printf("EnableXmla : %v\n", a.EnableXmla)
	fmt.Printf("Overwrite  : %v\n", a.Overwrite)
}

// GetAnalysisDatasourceInfo retreives the information of analysis datasource.
func (c *Client) GetAnalysisDatasourceInfo(name string) (*AnalysisDatasourceInfo, error) {
	c.Logger.Debug("ListAnalysisDatasources", zap.String("name", name))
	resp, err := c.client.R().
		Get(fmt.Sprintf("plugin/data-access/api/datasource/%s/getAnalysisDatasourceInfo", name))
	switch resp.StatusCode() {
	case 200:
		result := string(resp.Body())
		tokens := strings.Split(result, ";")
		dataSource := tokens[0][12 : len(tokens[0])-1]
		provider := tokens[1][10 : len(tokens[1])-1]
		enableXmla := tokens[2][12:len(tokens[2])-1] == "true"
		overwrite := tokens[3][11:len(tokens[3])-1] == "true"
		return &AnalysisDatasourceInfo{dataSource, provider, enableXmla, overwrite}, nil
	default:
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("Unknown error. statusCode=%d", resp.StatusCode())
	}
}

// ExportAnalysisDatasource exports an analysis datasource.
func (c *Client) ExportAnalysisDatasource(name string, file string, overwrite bool) (string, error) {
	c.Logger.Debug("ExportAnalysisDatasource", zap.String("name", name), zap.String("file", file), zap.Bool("overwrite", overwrite))
	helper := NewDownloadHelper(file, overwrite)
	err := helper.PrepareTemporaryFile()
	if err != nil {
		return "", err
	}
	defer helper.Clean()
	resp, err := c.client.R().
		SetOutput(helper.GetTemporaryFilePath()).
		Get(fmt.Sprintf("plugin/data-access/api/datasource/analysis/catalog/%s", name))
	switch resp.StatusCode() {
	case 200:
		dest, err := helper.MoveTemporaryFileToDestination(resp)
		return dest, err
	case 401:
		return "", errors.New("Unauthorized")
	default:
		if err != nil {
			return "", err
		}
		return "", fmt.Errorf("Unknown error. statusCode=%d", resp.StatusCode())
	}
}

// DeleteAnalysisDatasource deletes an analysis datasource.
func (c *Client) DeleteAnalysisDatasource(name string) error {
	c.Logger.Debug("DeleteAnalysisDatasource", zap.String("name", name))
	resp, err := c.client.R().
		Delete(fmt.Sprintf("plugin/data-access/api/datasource/analysis/catalog/%s", name))
	switch resp.StatusCode() {
	case 200:
		return nil
	case 401:
		return errors.New("User is not authorized to delete the analysis datasource")
	case 500:
		return errors.New("Unable to remove the analysis data")
	default:
		if err != nil {
			return err
		}
		return fmt.Errorf("Unknown error. statusCode=%d", resp.StatusCode())
	}
}

func detectAnalysisSchemaName(data *[]byte) (string, error) {
	type analysisSchema struct {
		Name string `xml:"name,attr"`
	}
	var a analysisSchema
	xml.Unmarshal(*data, &a)
	schemaName := a.Name
	if len(schemaName) == 0 {
		return "", errors.New("unsupported file")
	}
	return schemaName, nil
}

// ImportAnalysisDatasource imports an analysis datasource.
func (c *Client) ImportAnalysisDatasource(file string, options *ImportAnalysisDatasourceOptions) error {
	c.Logger.Debug("ImportAnalysisDatasource", zap.String("file", file), zap.String("options", fmt.Sprint(options)))
	// detect datasource name from schema file
	if len(options.DatasourceName) == 0 {
		data, _ := ioutil.ReadFile(file)
		name, err := detectAnalysisSchemaName(&data)
		if err != nil {
			return errors.Wrap(err, "failed to read:"+file)
		}
		options.DatasourceName = name
	}
	// set parameters with the datasource name
	if len(options.Parameters) == 0 {
		options.Parameters = fmt.Sprintf("Datasource=%s", options.DatasourceName)
	}

	// construct the request
	values := url.Values{}
	if len(options.SchemaFileInfo) > 0 {
		values["schemaFileInfo"] = []string{options.SchemaFileInfo}
	}
	if len(options.OrigCatalogName) > 0 {
		values["origCatalogName"] = []string{options.OrigCatalogName}
	}
	if len(options.DatasourceName) > 0 {
		values["datasourceName"] = []string{options.DatasourceName}
	}
	values["overwrite"] = []string{strconv.FormatBool(options.Overwrite)}
	values["xmlaEnabledFlag"] = []string{strconv.FormatBool(options.XmlaEnabledFlag)}
	if len(options.Parameters) > 0 {
		values["parameters"] = []string{options.Parameters}
	}
	resp, err := c.client.R().
		SetFile("uploadInput", file).
		SetMultiValueFormData(values).
		Put(fmt.Sprintf("plugin/data-access/api/datasource/analysis/catalog/%s", options.DatasourceName))
	switch resp.StatusCode() {
	case 201:
		return nil
	case 401:
		return errors.New("Import failed because publish is prohibited")
	case 403:
		return errors.New("Access Control Forbidden")
	case 409:
		return errors.New("Content already exists (use overwrite flag to force)")
	case 412:
		return errors.New("Analysis datasource import failed. Error code or message included in response entity")
	case 500:
		return errors.New("Unspecified general error has occurred")
	default:
		if err != nil {
			return err
		}
		return fmt.Errorf("Unknown error. statusCode=%d, msg=%s", resp.StatusCode(), string(resp.Body()))
	}
}

type ImportAnalysisDatasourceOptions struct {
	// A Mondrian schema XML file.
	UploadInput string
	// User selected name for the file. (optional)
	SchemaFileInfo string
	// The original catalog name. (optional)
	OrigCatalogName string
	// The datasource name. (not used)
	DatasourceName string
	// Flag for overwriting existing version of the file. The values are true and false.
	Overwrite bool
	// Is XMLA enabled or not. The values are true and false.
	XmlaEnabledFlag bool
	// Import parameters.
	Parameters string
}

type DatasourceCatalog struct {
	RawItems json.RawMessage `json:"Item"` // export field for unmarshal of entire request
}

func (c *DatasourceCatalog) getItems() ([]DatasourceCatalogItem, error) {
	if len(c.RawItems) == 0 {
		return []DatasourceCatalogItem{}, nil
	}
	if bytes.HasPrefix(bytes.TrimSpace(c.RawItems), []byte{'['}) {
		var items []DatasourceCatalogItem
		err := json.Unmarshal(c.RawItems, &items)
		if err != nil {
			return nil, err
		}
		return items, nil
	}
	var item DatasourceCatalogItem
	err := json.Unmarshal(c.RawItems, &item)
	if err != nil {
		return nil, err
	}
	return []DatasourceCatalogItem{item}, nil
}

func (c *DatasourceCatalog) getItemNames() ([]string, error) {
	items, err := c.getItems()
	if err != nil {
		return nil, err
	}
	itemNames := make([]string, len(items))
	for i, item := range items {
		itemNames[i] = item.Name
	}
	return itemNames, nil
}

type DatasourceCatalogItem struct {
	Type string `json:"@type"`
	Name string `json:"$"`
}

// ListJdbcDatasources lists the jdbc datasources.
func (c *Client) ListJdbcDatasources() ([]string, error) {
	c.Logger.Debug("ListJdbcDatasources")
	var result DatasourceCatalog
	resp, err := c.client.R().
		SetHeader("Accept", "application/json").
		SetResult(&result).
		Get("plugin/data-access/api/datasource/jdbc/connection")
	switch resp.StatusCode() {
	case 200:
		return result.getItemNames()
	default:
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("Unknown error. statusCode=%d", resp.StatusCode())
	}
}

// ExportJdbcDatasource exports an analysis datasource.
func (c *Client) ExportJdbcDatasource(name string, file string, overwrite bool) (string, error) {
	c.Logger.Debug("ExportJdbcDatasource", zap.String("name", name), zap.String("file", file), zap.Bool("overwrite", overwrite))
	helper := NewDownloadHelper(file, overwrite)
	helper.FilenameFunc = func(resp *resty.Response) string {
		return fmt.Sprintf("%s.jdbc.json", name)
	}
	err := helper.PrepareTemporaryFile()
	if err != nil {
		return "", err
	}
	defer helper.Clean()
	resp, err := c.client.R().
		SetOutput(helper.GetTemporaryFilePath()).
		Get(fmt.Sprintf("plugin/data-access/api/datasource/jdbc/connection/%s", name))
	switch resp.StatusCode() {
	case 200:
		dest, err := helper.MoveTemporaryFileToDestination(resp)
		return dest, err
	case 500:
		return "", errors.New("An error occurred retrieving the JDBC datasource")
	default:
		if err != nil {
			return "", err
		}
		return "", fmt.Errorf("Unknown error. statusCode=%d", resp.StatusCode())
	}
}

// ImportJdbcDatasource imports a jdbc datasource.
func (c *Client) ImportJdbcDatasource(file string) error {
	c.Logger.Debug("ImportJdbcDatasource", zap.String("file", file))
	// read json input file
	data, err := ioutil.ReadFile(file)
	if err != nil {
		errors.Wrap(err, "failed to read:"+file)
	}
	// detect the datasource name
	type datasource struct {
		Name string `json:"name"`
	}
	var d datasource
	json.Unmarshal(data, &d)
	datasourceName := d.Name
	if len(datasourceName) == 0 {
		return errors.New("unsupported file:" + file)
	}
	resp, err := c.client.R().
		SetBody(data).
		SetHeader("Content-Type", "application/json").
		Put(fmt.Sprintf("plugin/data-access/api/datasource/jdbc/connection/%s", datasourceName))
	switch resp.StatusCode() {
	case 200:
		return nil
	case 304:
		// Datasource was not modified
		return nil
	case 403:
		return errors.New("User is not authorized to add JDBC datasources")
	case 500:
		return errors.New("An unexected error occurred while adding the JDBC datasource")
	default:
		if err != nil {
			return err
		}
		return fmt.Errorf("Unknown error. statusCode=%d, msg=%s", resp.StatusCode(), string(resp.Body()))
	}
}

// DeleteJdbcDatasource deletes a jdbc datasource.
func (c *Client) DeleteJdbcDatasource(name string) error {
	c.Logger.Debug("DeleteJdbcDatasource", zap.String("name", name))
	resp, err := c.client.R().
		Delete(fmt.Sprintf("plugin/data-access/api/datasource/jdbc/connection/%s", name))
	switch resp.StatusCode() {
	case 200:
		return nil
	case 304:
		return errors.New("User is not authorized to remove the JDBC datasource or the connection does not exist")
	case 500:
		return errors.New("An unexected error occurred while deleting the JDBC datasource")
	default:
		if err != nil {
			return err
		}
		return fmt.Errorf("Unknown error. statusCode=%d", resp.StatusCode())
	}
}

// ListDswDatasources lists the dsw datasources.
func (c *Client) ListDswDatasources() ([]string, error) {
	c.Logger.Debug("ListDswDatasources")
	var result DatasourceCatalog
	resp, err := c.client.R().
		SetHeader("Accept", "application/json").
		SetResult(&result).
		Get("plugin/data-access/api/datasource/dsw/domain")
	switch resp.StatusCode() {
	case 200:
		return result.getItemNames()
	default:
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("Unknown error. statusCode=%d", resp.StatusCode())
	}
}

// ExportDswDatasource exports a DSW datasource.
func (c *Client) ExportDswDatasource(name string, file string, overwrite bool) (string, error) {
	c.Logger.Debug("ExportDswDatasource", zap.String("name", name), zap.String("file", file), zap.Bool("overwrite", overwrite))
	helper := NewDownloadHelper(file, overwrite)
	err := helper.PrepareTemporaryFile()
	if err != nil {
		return "", err
	}
	defer helper.Clean()
	resp, err := c.client.R().
		SetOutput(helper.GetTemporaryFilePath()).
		Get(fmt.Sprintf("plugin/data-access/api/datasource/dsw/domain/%s", name))
	switch resp.StatusCode() {
	case 200:
		dest, err := helper.MoveTemporaryFileToDestination(resp)
		return dest, err
	case 401:
		return "", errors.New("User is not authorized to export DSW datasource")
	case 500:
		return "", errors.New("Failure to export DSW datasource")
	default:
		if err != nil {
			return "", err
		}
		return "", fmt.Errorf("Unknown error. statusCode=%d", resp.StatusCode())
	}
}

// ImportDswDatasource imports the DSW datasource.
func (c *Client) ImportDswDatasource(file string, overwrite bool, checkConnection bool) error {
	c.Logger.Debug("ImportDswDatasource", zap.String("file", file), zap.Bool("overwrite", overwrite), zap.Bool("checkConnection", checkConnection))

	// detect the domain ID and extract the XMI file
	zipReader, err := zip.OpenReader(file)
	if err != nil {
		return errors.Wrap(err, "Unsupported format: "+file)
	}
	defer zipReader.Close()
	var domainID string
	var xmiFileName string
	var xmiFileReader io.ReadCloser
	for _, f := range zipReader.File {
		if domainID == "" && strings.HasSuffix(f.Name, "mondrian.xml") {
			reader, err := f.Open()
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("failed to open zip entry. (zip=%s, entry=%s)", file, f.Name))
			}
			defer reader.Close()
			data, err := ioutil.ReadAll(reader)
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("failed to read zip entry. (zip=%s, entry=%s)", file, f.Name))
			}
			name, err := detectAnalysisSchemaName(&data)
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("failed to detect schema name in zip entry. (zip=%s, entry=%s)", file, f.Name))
			}
			domainID = name
			continue
		}
		if strings.HasSuffix(f.Name, ".xmi") {
			xmiFileName = f.Name
			xmiFileReader, err = f.Open()
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("failed to open XMI file in zip. (zip=%s,entry=%s)", file, f.Name))
			}
			defer xmiFileReader.Close()
			continue
		}
	}
	if domainID == "" {
		return errors.New("domain ID cannot be detected")
	}
	domainID = domainID + ".xmi"

	// call import API
	resp, err := c.client.R().
		SetFileReader("metadataFile", xmiFileName, xmiFileReader).
		SetFormData(map[string]string{
			"domainId":        domainID,
			"overwrite":       strconv.FormatBool(overwrite),
			"checkConnection": strconv.FormatBool(checkConnection),
		}).
		Put("plugin/data-access/api/datasource/dsw/import")
	switch resp.StatusCode() {
	case 200:
		return nil
	case 409:
		return fmt.Errorf("invalid dsw. msg=%s", string(resp.Body()))
	default:
		if err != nil {
			return err
		}
		return fmt.Errorf("Unknown error. statusCode=%d, msg=%s", resp.StatusCode(), string(resp.Body()))
	}
}

// DeleteDswDatasource deletes a DSW datasource.
func (c *Client) DeleteDswDatasource(name string) error {
	c.Logger.Debug("DeleteDswDatasource", zap.String("name", name))
	resp, err := c.client.R().
		Delete(fmt.Sprintf("plugin/data-access/api/datasource/dsw/domain/%s", name))
	switch resp.StatusCode() {
	case 200:
		return nil
	case 401:
		return errors.New("User is not authorized to remove DSW datasource")
	default:
		if err != nil {
			return err
		}
		return fmt.Errorf("Unknown error. statusCode=%d", resp.StatusCode())
	}
}

// GetACLOfDswDatasource gets the ACL of a DSW datasource.
func (c *Client) GetACLOfDswDatasource(name string) error {
	c.Logger.Debug("GetACLOfDswDatasource", zap.String("name", name))
	resp, err := c.client.R().
		Get(fmt.Sprintf("plugin/data-access/api/datasource/dsw/%s/acl", name))
	switch resp.StatusCode() {
	case 200:
		fmt.Println(string(resp.Body()))
		return nil
	case 404:
		return errors.New("DSW found but no ACL set: " + name)
	case 409:
		return errors.New("DSW doesn't exist: " + name)
	default:
		if err != nil {
			return err
		}
		return fmt.Errorf("Unknown error. statusCode=%d", resp.StatusCode())
	}
}

// SetACLOfDswDatasource set the ACL of a DSW datasource.
func (c *Client) SetACLOfDswDatasource(name string) error {
	c.Logger.Debug("SetACLOfDswDatasource", zap.String("name", name))
	return errors.New("Not implemented")
}

// ListMetadataDatasources lists the metadata datasources.
func (c *Client) ListMetadataDatasources() ([]string, error) {
	c.Logger.Debug("ListMetadataDatasources")
	var result DatasourceCatalog
	resp, err := c.client.R().
		SetHeader("Accept", "application/json").
		SetResult(&result).
		Get("plugin/data-access/api/datasource/metadata/domain")
	switch resp.StatusCode() {
	case 200:
		return result.getItemNames()
	default:
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("Unknown error. statusCode=%d", resp.StatusCode())
	}
}

// ExportMetadataDatasource exports a DSW datasource.
func (c *Client) ExportMetadataDatasource(name string, file string, overwrite bool) (string, error) {
	c.Logger.Debug("ExportMetadataDatasource", zap.String("name", name), zap.String("file", file), zap.Bool("overwrite", overwrite))
	helper := NewDownloadHelper(file, overwrite)
	err := helper.PrepareTemporaryFile()
	if err != nil {
		return "", err
	}
	defer helper.Clean()
	resp, err := c.client.R().
		SetOutput(helper.GetTemporaryFilePath()).
		Get(fmt.Sprintf("plugin/data-access/api/datasource/metadata/domain/%s", name))
	switch resp.StatusCode() {
	case 200:
		dest, err := helper.MoveTemporaryFileToDestination(resp)
		return dest, err
	case 401:
		return "", errors.New("User is not authorized to export Metadata datasource")
	case 500:
		return "", errors.New("Failure to export Metadata datasource")
	default:
		if err != nil {
			return "", err
		}
		return "", fmt.Errorf("Unknown error. statusCode=%d", resp.StatusCode())
	}
}

// ImportMetadataDatasource imports the metadata datasource.
func (c *Client) ImportMetadataDatasource(file string, domainID string, overwrite bool) error {
	c.Logger.Debug("ImportMetadataDatasource", zap.String("file", file), zap.String("domainID", domainID), zap.Bool("overwrite", overwrite))

	if domainID == "" {
		_, name := filepath.Split(file)
		if strings.HasSuffix(name, ".xmi") {
			domainID = name[0 : len(name)-4]
		} else {
			return errors.New("domain ID cannot be detected")
		}
	}

	reader, err := os.Open(file)
	if err != nil {
		return errors.Wrap(err, "failed to open the input file: "+file)
	}
	defer reader.Close()

	// call import API
	resp, err := c.client.R().
		SetFileReader("metadataFile", domainID, reader).
		SetFormData(map[string]string{
			"overwrite": strconv.FormatBool(overwrite),
		}).
		Put(fmt.Sprintf("plugin/data-access/api/datasource/metadata/domain/%s", domainID))
	switch resp.StatusCode() {
	case 201:
		return nil
	case 401:
		return errors.New("Import failed because publish is prohibited")
	case 403:
		return errors.New("Access Control Forbidden")
	case 409:
		return errors.New("Content already exists (use overwrite flag to force")
	case 412:
		return errors.New("Metadata datasource import failed.  Error code or message included in response entity")
	case 500:
		return errors.New("Unspecified general error has occurred")
	default:
		if err != nil {
			return err
		}
		return fmt.Errorf("Unknown error. statusCode=%d, msg=%s", resp.StatusCode(), string(resp.Body()))
	}
}

// DeleteMetadataDatasource deletes a DSW datasource.
func (c *Client) DeleteMetadataDatasource(name string) error {
	c.Logger.Debug("DeleteMetadataDatasource", zap.String("name", name))
	resp, err := c.client.R().
		Delete(fmt.Sprintf("plugin/data-access/api/datasource/metadata/domain/%s", name))
	switch resp.StatusCode() {
	case 200:
		return nil
	case 401:
		return errors.New("User is not authorized to delete the Metadata datasource")
	default:
		if err != nil {
			return err
		}
		return fmt.Errorf("Unknown error. statusCode=%d", resp.StatusCode())
	}
}

/*
func (c *Client) ExportDatasources(file string) error {
	tempDir, err := ioutil.TempDir("", "exporttemp")
	if err != nil {
		return errors.Wrap(err, "failed to create temporary directory")
	}
	defer os.RemoveAll(tempDir)
	f, err := os.Create(file)
	if err != nil {
		return errors.Wrap(err, "failed to create file: "+file)
	}
	writer := zip.NewWriter(f)
	defer writer.Close()

	analysisDatasources, err := c.ListAnalysisDatasources()
	if err != nil {
		return errors.Wrap(err, "failed to list analysis datasources")
	}
	for _, d := range analysisDatasources {
		filename, err := c.ExportAnalysisDatasource(d, tempDir, true)
		entryWriter, err := writer.Create(filename)
		//ioutil.
	}
	return nil
}
*/
