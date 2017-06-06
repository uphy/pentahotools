package pentahoclient

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

func (c *Client) Tree(path string, depth int, showHidden bool) (*FileEntry, error) {
	c.client.Debug = true
	root := FileEntry{}
	resp, err := c.client.R().
		SetQueryParam("showHidden", strconv.FormatBool(showHidden)).
		SetQueryParam("depth", fmt.Sprintf("%d", depth)).
		SetHeader("Content-Type", "application/json").
		SetBody(&root).
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

type FileInfo struct {
	AclNode               string `json:"aclNode"`
	CreatedDate           string `json:"createdDate"`
	FileSize              string `json:"fileSize"`
	Folder                string `json:"folder"`
	Hidden                string `json:"hidden"`
	Id                    string `json:"id"`
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
type FileEntry struct {
	File     FileInfo    `json:"file"`
	Children []FileEntry `json:"children,omitempty"`
}
