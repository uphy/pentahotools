package client

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

// GetStatusCarteServer gets the status of the carte server.
func (c *Client) GetStatusCarteServer() (*CarteServerStatus, error) {
	Logger.Debug("GetCarteServerStatus")
	var status CarteServerStatus
	resp, err := c.client.R().
		SetQueryParam("xml", "Y").
		SetResult(&status).
		Get(fmt.Sprintf("kettle/status"))
	switch resp.StatusCode() {
	case 200:
		for i := range status.TransformationStatusList.List {
			status.TransformationStatusList.List[i].LoggingString = decodeLoggingString(status.TransformationStatusList.List[i].LoggingString)
		}
		for i := range status.JobStatusList.List {
			status.JobStatusList.List[i].LoggingString = decodeLoggingString(status.JobStatusList.List[i].LoggingString)
		}
		return &status, nil
	case 403:
		return nil, errors.New("User does not have administrative permissions")
	case 500:
		return nil, errors.New("Failure to complete the export")
	default:
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("Unknown error. statusCode=%d", resp.StatusCode())
	}
}

// GetStatus gets the status of job or transformation.
func (c *Client) GetStatus(id string, name string, from int) (Status, error) {
	client, err := c.GetCarteClient(id, name)
	if err != nil {
		return nil, err
	}
	return client.GetStatus(id, name, from)
}

// RemoveJobOrTransformation removes job or transformation.
func (c *Client) RemoveJobOrTransformation(id, name string) error {
	client, err := c.GetCarteClient(id, name)
	if err != nil {
		return err
	}
	return client.Remove(id, name)
}

// Run runs a job.
func (c *Client) Run(file string, level LogLevel) (string, error) {
	if strings.HasSuffix(file, ".kjb") {
		return c.JobClient.Run(file, level)
	} else if strings.HasSuffix(file, ".ktr") {
		return c.TransformationClient.Run(file, level)
	} else {
		// complements the kjb/ktr extension.
		if !strings.HasPrefix(file, "/") {
			file = "/" + file
		}
		dir, filename := filepath.Split(file)
		dirEntry, err := c.Tree(dir, 1, false)
		if err != nil {
			return "", err
		}
		for _, f := range dirEntry.Children {
			if strings.HasPrefix(f.File.Name, filename) {
				switch f.File.Name {
				case filename + ".kjb":
					return c.JobClient.Run(file, level)
				case filename + ".ktr":
					return c.TransformationClient.Run(file, level)
				default:
					continue
				}
			}
		}
		return "", errors.New("unknown file:" + file)
	}
}

// webResult represents the result of the job or transformation.
type webResult struct {
	Result  string `xml:"result"`
	Message string `xml:"message"`
	ID      string `xml:"id"`
}

// GetCarteClient gets the carte client by the specified id and name.
func (c *Client) GetCarteClient(id string, name string) (CarteClient, error) {
	carteStatus, err := c.GetStatusCarteServer()
	if err != nil {
		return nil, errors.Wrap(err, "getting carte status failed")
	}
	for _, s := range carteStatus.JobStatusList.List {
		if s.ID == id || s.Name == name {
			return c.JobClient, nil
		}
	}
	for _, s := range carteStatus.TransformationStatusList.List {
		if s.ID == id || s.Name == name {
			return c.TransformationClient, nil
		}
	}
	return nil, fmt.Errorf("no such job or transformation. (id=%s, name=%s)", id, name)
}

// CarteClient represents the carte job or transformation client.
type CarteClient interface {
	GetStatus(id string, name string, from int) (Status, error)
	Run(file string, level LogLevel) (string, error)
	Remove(id, name string) error
}
