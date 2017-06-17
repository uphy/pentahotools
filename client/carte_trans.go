package client

import (
	"encoding/xml"
	"errors"
	"fmt"
	"strings"

	resty "gopkg.in/resty.v0"

	"go.uber.org/zap"
)

// TransformationClient is a carte client for transformations.
type TransformationClient struct {
	client *resty.Client
}

// GetStatus gets the status of the transformation.
func (c *TransformationClient) GetStatus(id string, name string, from int) (Status, error) {
	Logger.Debug("GetStatusTransformation", zap.String("id", id), zap.String("name", name), zap.Int("from", from))
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
	if from >= 0 {
		req.SetQueryParam("from", fmt.Sprint(from))
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

// Run runs the transformation.
func (c *TransformationClient) Run(file string, level LogLevel) (string, error) {
	Logger.Debug("RunTrans", zap.String("file", file))
	if strings.HasSuffix(file, ".ktr") {
		file = file[0 : len(file)-4]
	}
	resp, err := c.client.R().
		SetFormData(map[string]string{
			"trans": file,
			"level": string(level),
		}).
		SetHeader("Accept", "*/*").
		Post("kettle/runTrans/")
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

// Remove removes the transformation.
func (c *TransformationClient) Remove(id, name string) error {
	Logger.Debug("RemoveTransformation", zap.String("id", id), zap.String("name", name))
	req := c.client.R().
		SetQueryParam("xml", "Y")
	if id != "" {
		req.SetQueryParam("id", id)
	}
	if name != "" {
		req.SetQueryParam("name", name)
	}
	resp, err := req.Get("kettle/removeTrans/")
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
