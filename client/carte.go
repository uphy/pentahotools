package client

import (
	"compress/gzip"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
)

// GetCarteServerStatus gets the status of the carte server.
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

func (c *Client) GetStatus(id string, name string) (Status, error) {
	carteStatus, err := c.GetStatusCarteServer()
	if err != nil {
		return nil, err
	}
	for _, s := range carteStatus.JobStatusList.List {
		if s.ID == id || s.Name == name {
			return c.GetStatusJob(id, name)
		}
	}
	for _, s := range carteStatus.TransformationStatusList.List {
		if s.ID == id || s.Name == name {
			return c.GetStatusTransformation(id, name)
		}
	}
	return nil, fmt.Errorf("No such jobs/transformations. (id=%s, name=%s)", id, name)
}

// GetStatusTransformation gets the status of the transformation.
func (c *Client) GetStatusTransformation(id string, name string) (*TransformationStatus, error) {
	if id == "" && name == "" {
		return nil, errors.New("specify either id or name")
	}
	var status TransformationStatus
	req := c.client.R().
		SetResult(&status).
		SetQueryParam("xml", "Y")
	if id != "" {
		req.SetQueryParam("id", id)
	}
	if name != "" {
		req.SetQueryParam("name", name)
	}
	resp, err := req.Get(fmt.Sprintf("kettle/transStatus/"))
	switch resp.StatusCode() {
	case 200:
		status.LoggingString = decodeLoggingString(status.LoggingString)
		return &status, nil
	case 403:
		return nil, errors.New("User does not have administrative permissions")
	case 500:
		return nil, errors.New("server error")
	default:
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("Unknown error. statusCode=%d", resp.StatusCode())
	}
}

// GetStatusJob gets the status of the job.
func (c *Client) GetStatusJob(id string, name string) (*JobStatus, error) {
	if id == "" && name == "" {
		return nil, errors.New("specify either id or name")
	}
	var status JobStatus
	req := c.client.R().
		SetResult(&status).
		SetQueryParam("xml", "Y")
	if id != "" {
		req.SetQueryParam("id", id)
	}
	if name != "" {
		req.SetQueryParam("name", name)
	}
	resp, err := req.Get(fmt.Sprintf("kettle/jobStatus/"))
	switch resp.StatusCode() {
	case 200:
		status.LoggingString = decodeLoggingString(status.LoggingString)
		return &status, nil
	case 403:
		return nil, errors.New("User does not have administrative permissions")
	case 500:
		return nil, errors.New("server error")
	default:
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("Unknown error. statusCode=%d", resp.StatusCode())
	}
}

func decodeLoggingString(loggingString string) string {
	loggingString = loggingString[len("<![CDATA[") : len(loggingString)-len("]]>")]
	if len(loggingString) == 0 {
		return ""
	}
	reader := strings.NewReader(loggingString)
	base64Reader := base64.NewDecoder(base64.StdEncoding, reader)
	gzipReader, _ := gzip.NewReader(base64Reader)
	decoded, _ := ioutil.ReadAll(gzipReader)
	return string(decoded)
}
