package client

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"go.uber.org/zap"

	"path/filepath"

	"net/url"

	"encoding/xml"

	"html"

	"github.com/pkg/errors"
)

// Backup backups whole of the pentaho.
func (c *Client) Backup(output string) error {
	c.Logger.Debug("Backup", zap.String("output", output))
	resp, err := c.client.R().
		SetOutput(output).
		Get(fmt.Sprintf("api/repo/files/backup"))
	switch resp.StatusCode() {
	case 200:
		return nil
	case 403:
		return errors.New("User does not have administrative permissions")
	case 500:
		return errors.New("Failure to complete the export")
	default:
		if err != nil {
			return err
		}
		return fmt.Errorf("Unknown error. statusCode=%d", resp.StatusCode())
	}
}

// Restore restores whole of the pentaho from a file.
func (c *Client) Restore(input string, overwrite bool) error {
	c.Logger.Debug("Restore", zap.String("input", input), zap.Bool("overwrite", overwrite))
	resp, err := c.client.R().
		SetFile("fileUpload", input).
		SetFormData(map[string]string{
			"overwriteFile": strconv.FormatBool(overwrite),
		}).
		Post(fmt.Sprintf("api/repo/files/systemRestore"))
	switch resp.StatusCode() {
	case 200:
		return nil
	case 403:
		return errors.New("User does not have administrative permissions")
	case 500:
		return errors.New("Failure to complete the export")
	default:
		if err != nil {
			return err
		}
		return fmt.Errorf("Unknown error. statusCode=%d", resp.StatusCode())
	}
}

// Tree list the children of the specified path.
func (c *Client) Tree(path string, depth int, showHidden bool) (*FileEntry, error) {
	c.Logger.Debug("Tree", zap.String("path", path), zap.Int("depth", depth), zap.Bool("showHidden", showHidden))
	var root FileEntry
	resp, err := c.client.R().
		SetQueryParam("showHidden", strconv.FormatBool(showHidden)).
		SetQueryParam("depth", fmt.Sprintf("%d", depth)).
		SetHeader("Accept", "application/json").
		SetResult(&root).
		Get(fmt.Sprintf("api/repo/files/%s/tree", strings.Replace(path, "/", ":", -1)))

	switch resp.StatusCode() {
	case 200:
		return &root, nil
	case 404:
		return nil, errors.New("invalid parameters")
	case 500:
		return nil, errors.New("server error")
	default:
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("Unknown error. statusCode=%d", resp.StatusCode())
	}
}

// GetACL gets the access control list of file.
func (c *Client) GetACL(path string) (*ACL, error) {
	c.Logger.Debug("GetACL", zap.String("path", path))
	var acl ACL
	resp, err := c.client.R().
		SetHeader("Accept", "application/json").
		SetResult(&acl).
		Get(fmt.Sprintf("api/repo/files/%s/acl", strings.Replace(path, "/", ":", -1)))

	switch resp.StatusCode() {
	case 200:
		return &acl, nil
	case 403:
		return nil, errors.New("Failed to save acls due to missing or incorrect properties")
	case 400:
		return nil, errors.New("Failed to save acls due to malformed xml")
	case 500:
		return nil, errors.New("Failed to save acls due to another error")
	default:
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("Unknown error. statusCode=%d", resp.StatusCode())
	}
}

// SetACL set the access control list of file.
func (c *Client) SetACL(path string, acl *ACL) error {
	if acl == nil {
		return errors.New("acl == nil")
	}
	dto := struct {
		ACL
		XMLName struct{} `xml:"repositoryFileAclDto"`
	}{ACL: *acl}
	xmlBytes, _ := xml.Marshal(dto)
	c.Logger.Debug("SetACL", zap.String("path", path), zap.String("acl", string(xmlBytes)))
	resp, err := c.client.R().
		SetHeader("Content-Type", "application/xml").
		SetBody(dto).
		Put(fmt.Sprintf("api/repo/files/%s/acl", strings.Replace(path, "/", ":", -1)))

	switch resp.StatusCode() {
	case 200:
		return nil
	case 403:
		return errors.New("Failed to save acls due to missing or incorrect properties")
	case 400:
		return errors.New("Failed to save acls due to malformed xml")
	case 500:
		return errors.New("Failed to save acls due to another error")
	default:
		if err != nil {
			return err
		}
		return fmt.Errorf("Unknown error. statusCode=%d", resp.StatusCode())
	}
}

const (
	ACPermissionRead                        = "0"
	ACPermissionReadWrite                   = "1"
	ACPermissionReadWriteDelete             = "2"
	ACPermissionReadWriteDeleteAdministrate = "4"
)

// Ac represents an access control
type Ac struct {
	Modifiable    string `xml:"modifiable"`
	Permissions   string `xml:"permissions"`
	Recipient     string `xml:"recipient"`
	RecipientType string `xml:"recipientType"`
}

func (a *Ac) PermissionsString() string {
	switch a.Permissions {
	case ACPermissionRead:
		return "Read"
	case ACPermissionReadWrite:
		return "Read/Write"
	case ACPermissionReadWriteDelete:
		return "Read/Write/Delete"
	case ACPermissionReadWriteDeleteAdministrate:
		return "Read/Write/Delete/Administrate"
	}
	return fmt.Sprintf("<Unknown:%s>", a.Permissions)
}

// ACL represents an access control list
type ACL struct {
	Aces              []Ac   `xml:"aces"`
	EntriesInheriting string `xml:"entriesInheriting"`
	ID                string `xml:"id"`
	Owner             string `xml:"owner"`
	OwnerType         string `xml:"ownerType"`
}

func (a *ACL) GetOrNewAC(recipient string) *Ac {
	for _, ac := range a.Aces {
		if ac.Recipient == recipient {
			return &ac
		}
	}
	ac := Ac{
		Modifiable:    "true",
		Permissions:   ACPermissionRead,
		Recipient:     recipient,
		RecipientType: "0",
	}
	a.Aces = append(a.Aces, ac)
	return &ac
}

func (a *ACL) SetAC(ac *Ac) error {
	if ac.Permissions != ACPermissionRead && ac.Permissions != ACPermissionReadWrite && ac.Permissions != ACPermissionReadWriteDelete && ac.Permissions != ACPermissionReadWriteDeleteAdministrate {
		return errors.New("Unknown permissions:" + ac.Permissions)
	}
	for i, ac2 := range a.Aces {
		if ac2.Recipient == ac.Recipient {
			a.Aces[i] = *ac
			return nil
		}
	}
	a.Aces = append(a.Aces, *ac)
	return nil
}

func (a *ACL) DeleteAC(recipient string) {
	for i, ac := range a.Aces {
		if ac.Recipient == recipient {
			a.Aces = append(a.Aces[:i], a.Aces[i+1:]...)
			break
		}
	}
}

const (
	FileMetadataSchedulable = "_PERM_SCHEDULABLE"
	FileMetadataHidden      = "_PERM_HIDDEN"
)

type FileMetadata struct {
	Entries []FileMetadataEntry `json:"stringKeyStringValueDto"`
}

func (m *FileMetadata) load(metadata map[string]interface{}) {
	for key, value := range metadata {
		m.Entries = append(m.Entries, FileMetadataEntry{key, value})
	}
}

func (m *FileMetadata) toMap() map[string]interface{} {
	result := map[string]interface{}{}
	for _, entry := range m.Entries {
		result[entry.Key] = entry.Value
	}
	return result
}

type FileMetadataEntry struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

func (c *Client) SetMetadata(file string, metadata map[string]interface{}) error {
	c.Logger.Debug("SetMetadata", zap.String("file", file), zap.String("metadata", fmt.Sprint(metadata)))
	m := FileMetadata{[]FileMetadataEntry{}}
	m.load(metadata)
	if metadata == nil {
		return errors.New("metadata == nil")
	}
	resp, err := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(m).
		Put(fmt.Sprintf("api/repo/files/%s/metadata", strings.Replace(file, "/", ":", -1)))

	switch resp.StatusCode() {
	case 200:
		return nil
	case 403:
		return errors.New("Invalid path")
	case 400:
		return errors.New("Invalid payload")
	case 500:
		return errors.New("Server Error")
	default:
		if err != nil {
			return err
		}
		return fmt.Errorf("Unknown error. statusCode=%d", resp.StatusCode())
	}
}

func (c *Client) GetMetadata(file string) (map[string]interface{}, error) {
	c.Logger.Debug("GetMetadata", zap.String("file", file))
	var metadata FileMetadata
	resp, err := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetResult(&metadata).
		Get(fmt.Sprintf("api/repo/files/%s/metadata", strings.Replace(file, "/", ":", -1)))

	switch resp.StatusCode() {
	case 200:
		return metadata.toMap(), nil
	case 403:
		return nil, errors.New("Invalid path")
	case 400:
		return nil, errors.New("Invalid payload")
	case 500:
		return nil, errors.New("Server Error")
	default:
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("Unknown error. statusCode=%d", resp.StatusCode())
	}
}

// PutFile put the file to the repository.
// destination should be a absolute file path.
func (c *Client) PutFile(file string, destination string) error {
	c.Logger.Debug("PutFile", zap.String("file", file), zap.String("destination", destination))
	if strings.HasSuffix(destination, "/") {
		_, filename := filepath.Split(file)
		destination = destination + filename
	}
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	resp, err := c.client.R().
		SetBody(data).
		Put(fmt.Sprintf("api/repo/files/%s", strings.Replace(destination, "/", ":", -1)))

	switch resp.StatusCode() {
	case 200:
		return nil
	case 403:
		return errors.New("Failure to create the file due to permissions, file already exists, or invalid path id")
	case 500:
		return errors.New("server error")
	default:
		if err != nil {
			return err
		}
		return fmt.Errorf("Unknown error. statusCode=%d", resp.StatusCode())
	}
}

// CreateDirectory creates new directory by the specified path.
func (c *Client) CreateDirectory(path string) error {
	resp, err := c.client.R().
		Put(fmt.Sprintf("api/repo/files/%s/createDir", strings.Replace(path, "/", ":", -1)))
	switch resp.StatusCode() {
	case 200:
		return nil
	case 409:
		return errors.New("Path already exists")
	case 500:
		return errors.New("server error")
	default:
		if err != nil {
			return err
		}
		return fmt.Errorf("Unknown error. statusCode=%d", resp.StatusCode())
	}
}

// ClearCache clears the cache
func (c *Client) ClearCache(catalog string) error {
	resp, err := c.client.R().
		SetFormData(map[string]string{
			"catalog": catalog,
		}).
		Post("api/repos/xanalyzer/service/ajax/clearCache")
	switch resp.StatusCode() {
	case 200:
		var result Message
		xml.Unmarshal(resp.Body(), &result)
		detail := html.UnescapeString(result.Details)
		if strings.Contains(detail, "find") {
			return errors.New("unable to find catalog: " + catalog)
		}
		body := string(resp.Body())
		if body == "true" {
			return nil
		}
		return errors.New("failed to clear cache")
	default:
		if err != nil {
			return err
		}
		return fmt.Errorf("Unknown error. statusCode=%d", resp.StatusCode())
	}
}

// Message represents the response message object
type Message struct {
	ID      string `xml:"id,attr"`
	Ticket  string `xml:"ticket,attr"`
	Details string `xml:"details,attr"`
	Type    string `xml:"type,attr"`
}

// GetFile gets file from the repository
func (c *Client) GetFile(repositoryPath string, destination string) error {
	if _, err := c.getFile(repositoryPath, destination); err != nil {
		return err
	}
	return nil
}

// GetFileContent gets the content of the file in the repository
func (c *Client) GetFileContent(repositoryPath string) ([]byte, error) {
	data, err := c.getFile(repositoryPath, "")
	if err != nil {
		return nil, err
	}
	return data, nil
}

// getFile gets file from the repository
func (c *Client) getFile(repositoryPath string, destination string) ([]byte, error) {
	c.Logger.Debug("getFile", zap.String("repositoryPath", repositoryPath), zap.String("destination", destination))
	isCatMode := destination == ""
	req := c.client.R()
	if !isCatMode {
		req.SetOutput(destination)
	}
	resp, err := req.Get(fmt.Sprintf("api/repo/files/%s", strings.Replace(repositoryPath, "/", ":", -1)))
	defer func() {
		if resp.StatusCode() == 200 {
			return
		}
		if _, err := os.Stat(destination); err == nil {
			c.Logger.Debug("deleting temp file due to the previous error", zap.String("destination", destination))
			if err := os.Remove(destination); err != nil {
				c.Logger.Error("failed to delete the temp file.", zap.String("target", destination))
			}
		}
	}()

	switch resp.StatusCode() {
	case 200:
		if isCatMode {
			return resp.Body(), nil
		}
		return nil, nil
	case 403:
		return nil, errors.New("Failure to create the file due to permissions, file already exists, or invalid path id")
	case 404:
		return nil, errors.New("file not found")
	case 500:
		return nil, errors.New("server error")
	default:
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("Unknown error. statusCode=%d", resp.StatusCode())
	}
}

// DownloadFile gets file from the repository
func (c *Client) DownloadFile(repositoryPath string, destination string, withManifest bool, overwrite bool) (string, error) {
	c.Logger.Debug("DownloadFile", zap.String("repositoryPath", repositoryPath), zap.String("destination", destination), zap.Bool("withManifest", withManifest), zap.Bool("overwrite", overwrite))
	helper := NewDownloadHelper(destination, overwrite)
	err := helper.PrepareTemporaryFile()
	if err != nil {
		return "", err
	}
	defer helper.Clean()
	resp, err := c.client.R().
		SetOutput(helper.GetTemporaryFilePath()).
		SetHeader("User-Agent", "Firefox").
		SetQueryParam("withManifest", strconv.FormatBool(withManifest)).
		Get(fmt.Sprintf("api/repo/files/%s/download", strings.Replace(repositoryPath, "/", ":", -1)))
	switch resp.StatusCode() {
	case 200:
		return helper.MoveTemporaryFileToDestination(resp)
	case 403:
		return "", errors.New("Failure to create the file due to permissions, file already exists, or invalid path id")
	case 404:
		return "", errors.New("file not found")
	case 500:
		return "", errors.New("server error")
	default:
		if err != nil {
			return "", err
		}
		return "", fmt.Errorf("Unknown error. statusCode=%d", resp.StatusCode())
	}
}

// ImportFile imports a file to the directory in the repository.
func (c *Client) ImportFile(file string, importDir string, params *ImportParameters) error {
	c.Logger.Debug("ImportFile", zap.String("file", file), zap.String("importDir", importDir), zap.String("params", fmt.Sprint(params)))
	resp, err := c.client.R().
		SetFiles(map[string]string{
			"fileUpload": file,
		}).
		SetMultiValueFormData(url.Values{
			"overwriteFile":           []string{strconv.FormatBool(params.OverwriteFile)},
			"logLevel":                []string{params.LogLevel},
			"retainOwnership":         []string{strconv.FormatBool(params.RetainOwnership)},
			"fileNameOverride":        []string{params.FileNameOverride},
			"importDir":               []string{importDir},
			"charSet":                 []string{params.Charset},
			"applyAclPermissions":     []string{strconv.FormatBool(params.ApplyACLPermissions)},
			"overwriteAclPermissions": []string{strconv.FormatBool(params.OverwriteACLPermissions)},
		}).
		Post("api/repo/files/import")
	if err != nil {
		return err
	}
	switch resp.StatusCode() {
	case 200:
		return nil
	case 500:
		return errors.New("server error")
	default:
		if err != nil {
			return err
		}
		return fmt.Errorf("Unknown error. statusCode=%d", resp.StatusCode())
	}
}

// DeleteFiles move file to trash folder of the repository.
func (c *Client) DeleteFiles(repositoryPaths ...string) error {
	c.Logger.Debug("DeleteFile", zap.Strings("repositoryPaths", repositoryPaths))
	return c.deleteFile(false, repositoryPaths...)
}

// DeleteFilesPermanently deletes file from the repository.
func (c *Client) DeleteFilesPermanently(repositoryPaths ...string) error {
	c.Logger.Debug("DeleteFilesPermanently", zap.Strings("repositoryPaths", repositoryPaths))
	return c.deleteFile(true, repositoryPaths...)
}

func (c *Client) deleteFile(permanent bool, repositoryPath ...string) error {
	ids := make([]string, len(repositoryPath))
	for i, path := range repositoryPath {
		acl, err := c.GetACL(path)
		if err != nil {
			return errors.Wrap(err, "failed to get the ID of the deleting file:"+path)
		}
		ids[i] = acl.ID
	}
	var apiPath string
	if permanent {
		apiPath = "api/repo/files/deletepermanent"
	} else {
		apiPath = "api/repo/files/delete"
	}
	resp, err := c.client.R().
		SetBody(strings.Join(ids, ",")).
		Put(apiPath)

	switch resp.StatusCode() {
	case 200:
		return nil
	case 500:
		return errors.New("Failure to move the files specified in the comma-separated list to the trash")
	default:
		if err != nil {
			return err
		}
		return fmt.Errorf("Unknown error. statusCode=%d", resp.StatusCode())
	}
}

// FileEntry represents a file or directory.
type FileEntry struct {
	Children []FileEntry `json:"children"`
	File     FileInfo    `json:"file"`
}

// FileInfo represents a file information.
type FileInfo struct {
	ACLNode               string `json:"aclNode"`
	CreatedDate           string `json:"createdDate"`
	FileSize              string `json:"fileSize"`
	Folder                string `json:"folder"`
	Hidden                string `json:"hidden"`
	ID                    string `json:"id"`
	Locale                string `json:"locale"`
	Locked                string `json:"locked"`
	Name                  string `json:"name"`
	NotSchedulable        string `json:"notSchedulable"`
	OwnerType             string `json:"ownerType"`
	Path                  string `json:"path"`
	Title                 string `json:"title"`
	VersionCommentEnabled string `json:"versionCommentEnabled"`
	Versioned             string `json:"versioned"`
	VersioningEnabled     string `json:"versioningEnabled"`
}

type PrintOption struct {
	Indent       bool
	AbsolutePath bool
}

// Print prints the file entry recursively
func (e *FileEntry) Print(option *PrintOption) {
	e.print(0, "", option)
}

func (e *FileEntry) print(level int, parentPath string, option *PrintOption) {
	name := e.File.Name
	var path string
	var pathForPrint string
	if level == 0 {
		path = name
		pathForPrint = "/" + name
	} else {
		path = fmt.Sprintf("%s/%s", parentPath, name)
		pathForPrint = path
	}
	if option.Indent {
		fmt.Print(fmt.Sprintf(fmt.Sprintf("%%%ds", level*3), ""))
	}
	if option.AbsolutePath {
		fmt.Println(pathForPrint)
	} else {
		fmt.Println(name)
	}
	for _, entry := range e.Children {
		entry.print(level+1, path, option)
	}
}

// ImportParameters is the parameters of import API.
type ImportParameters struct {
	OverwriteFile           bool
	OverwriteACLPermissions bool
	ApplyACLPermissions     bool
	RetainOwnership         bool
	Charset                 string
	LogLevel                string
	FileNameOverride        string
}
