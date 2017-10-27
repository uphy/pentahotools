package client

import (
	"encoding/xml"
	"errors"
	"fmt"
	"strings"

	resty "gopkg.in/resty.v0"

	"go.uber.org/zap"
)

// JobClient is the carte client for Job.
type JobClient struct {
	client *resty.Client
	logger Logger
}

// GetStatus gets the status of the job.
func (c *JobClient) GetStatus(id string, name string, from int) (Status, error) {
	c.logger.Debug("GetStatusJob", zap.String("id", id), zap.String("name", name), zap.Int("from", from))
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
	if from >= 0 {
		req.SetQueryParam("from", fmt.Sprint(from))
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

// Run runs a job
func (c *JobClient) Run(file string, level LogLevel) (string, error) {
	c.logger.Debug("RunJob", zap.String("file", file))
	if strings.HasSuffix(file, ".kjb") {
		file = file[0 : len(file)-4]
	}
	resp, err := c.client.R().
		SetFormData(map[string]string{
			"job":   file,
			"level": string(level),
		}).
		SetHeader("Accept", "*/*").
		Post("kettle/runJob/")
	switch resp.StatusCode() {
	case 200:
		var result webResult
		xml.Unmarshal(resp.Body(), &result)
		if result.Result != "OK" {
			return "", errors.New(result.Message)
		}
		return result.ID, nil
	case 500:
		return "", errors.New("server error")
	default:
		if err != nil {
			return "", err
		}
		return "", fmt.Errorf("Unknown error. statusCode=%d", resp.StatusCode())
	}
}

// Remove removes job
func (c *JobClient) Remove(id, name string) error {
	c.logger.Debug("RemoveJob", zap.String("id", id), zap.String("name", name))
	req := c.client.R().
		SetQueryParam("xml", "Y")
	if id != "" {
		req.SetQueryParam("id", id)
	}
	if name != "" {
		req.SetQueryParam("name", name)
	}
	resp, err := req.Get("kettle/removeJob/")
	switch resp.StatusCode() {
	case 200:
		return nil
	case 500:
		return errors.New("Internal server error occurs during request processing")
	default:
		if err != nil {
			return err
		}
		return fmt.Errorf("Unknown error. statusCode=%d", resp.StatusCode())
	}
}
