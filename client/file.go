package pentahoclient

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"go.uber.org/zap"

	"github.com/pkg/errors"
)

// Backup backups whole of the pentaho.
func (c *Client) Backup(output string) error {
	Logger.Debug("Backup", zap.String("output", output))
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
	Logger.Debug("Restore", zap.String("input", input), zap.Bool("overwrite", overwrite))
	resp, err := c.client.R().
		SetFile("fileUpload", input).
		SetFormData(map[string]string{
			"overwrite": strconv.FormatBool(overwrite),
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
	Logger.Debug("Tree", zap.String("path", path), zap.Int("depth", depth), zap.Bool("showHidden", showHidden))
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
	Logger.Debug("GetACL", zap.String("path", path))
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

// Ac represents an access control
type Ac struct {
	Modifiable    string
	Permissions   string
	Recipient     string
	RecipientType string
}

// ACL represents an access control list
type ACL struct {
	Aces              []Ac
	EntriesInheriting string
	ID                string
	Owner             string
	OwnerType         string
}

// PutFile put the file to the repository.
func (c *Client) PutFile(file string, destination string) error {
	Logger.Debug("PutFile", zap.String("file", file), zap.String("destination", destination))
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

// GetFile gets file from the repository
func (c *Client) GetFile(repositoryPath string, destination string) error {
	Logger.Debug("GetFile", zap.String("repositoryPath", repositoryPath), zap.String("destination", destination))
	resp, err := c.client.R().
		SetOutput(destination).
		Get(fmt.Sprintf("api/repo/files/%s", strings.Replace(repositoryPath, "/", ":", -1)))

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

// DeleteFiles move file to trash folder of the repository.
func (c *Client) DeleteFiles(repositoryPaths ...string) error {
	Logger.Debug("DeleteFile", zap.Strings("repositoryPaths", repositoryPaths))
	return c.deleteFile(false, repositoryPaths...)
}

// DeleteFilesPermanently deletes file from the repository.
func (c *Client) DeleteFilesPermanently(repositoryPaths ...string) error {
	Logger.Debug("DeleteFilesPermanently", zap.Strings("repositoryPaths", repositoryPaths))
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

// Print prints the file entry recursively
func (e *FileEntry) Print() {
	e.print(0)
}

func (e *FileEntry) print(level int) {
	indent := fmt.Sprintf(fmt.Sprintf("%%%ds", level*2), "")
	fmt.Printf("%s%s (%s)\n", indent, e.File.Name, e.File.Path)
	for _, entry := range e.Children {
		entry.print(level + 1)
	}
}
