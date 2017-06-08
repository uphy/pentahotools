package pentahoclient

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

func (c *Client) Backup(output string) error {
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

func (c *Client) Restore(input string, overwrite bool) error {
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
func (c *Client) GetACL(path string) (*Acl, error) {
	var acl Acl
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

// PutFile put the file to the repository.
func (c *Client) PutFile(file string, destination string) error {
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

// DeleteFile delets file from the repository
func (c *Client) DeleteFile(repositoryPath ...string) error {
	return c.deleteFile(false, repositoryPath...)
}

// DeleteFilePermanent delets file from the repository
func (c *Client) DeleteFilePermanent(repositoryPath ...string) error {
	return c.deleteFile(true, repositoryPath...)
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

type Ac struct {
	Modifiable    string
	Permissions   string
	Recipient     string
	RecipientType string
}

type Acl struct {
	Aces              []Ac
	EntriesInheriting string
	ID                string
	Owner             string
	OwnerType         string
}
