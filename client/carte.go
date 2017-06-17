package client

import (
	"compress/gzip"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"

	"go.uber.org/zap"
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

func (c *Client) GetStatus(id string, name string) (Status, error) {
	Logger.Debug("GetStatus", zap.String("id", id), zap.String("id", name))
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

// RemoveJobOrTransformation removes job or transformation
func (c *Client) RemoveJobOrTransformation(id, name string) error {
	err1 := c.RemoveJob(id, name)
	err2 := c.RemoveTransformation(id, name)
	if err1 != nil && err2 != nil {
		return errors.New("failed to remove")
	}
	return nil
}

// RemoveJob removes job
func (c *Client) RemoveJob(id, name string) error {
	Logger.Debug("RemoveJob", zap.String("id", id), zap.String("name", name))
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

// RemoveTransformation removes a transformation
func (c *Client) RemoveTransformation(id, name string) error {
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

// GetStatusTransformation gets the status of the transformation.
func (c *Client) GetStatusTransformation(id string, name string) (*TransformationStatus, error) {
	Logger.Debug("GetStatusTransformation", zap.String("id", id), zap.String("id", name))
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
	Logger.Debug("GetStatusJob", zap.String("id", id), zap.String("id", name))
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

// RunJob run the specified job now.
func (c *Client) RunJob(file string) (string, error) {
	Logger.Debug("RunJob", zap.String("file", file))
	return c.ScheduleJob(&JobScheduleRequest{
		InputFile:       file,
		RunInBackground: true,
	})
}

// GetJobInfo gets the job info.
func (c *Client) GetJobInfo(jobID string) (*Job, error) {
	Logger.Debug("GetJobInfo", zap.String("jobID", jobID))
	var job Job
	resp, err := c.client.R().
		SetQueryParam("jobId", jobID).
		SetHeader("Accept", "application/json").
		SetResult(&job).
		Get("api/scheduler/jobinfo")
	switch resp.StatusCode() {
	case 200:
		return &job, nil
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

// ScheduleJob schedules job.
func (c *Client) ScheduleJob(req *JobScheduleRequest) (string, error) {
	Logger.Debug("ScheduleJob")
	resp, err := c.client.R().
		SetBody(req).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "text/plain").
		Post("api/scheduler/job")
	switch resp.StatusCode() {
	case 200:
		return string(resp.Body()), nil
	case 403:
		return "", errors.New("User does not have administrative permissions")
	case 500:
		return "", errors.New("server error")
	default:
		if err != nil {
			return "", err
		}
		return "", fmt.Errorf("Unknown error. statusCode=%d", resp.StatusCode())
	}
}

// JobParam is a job parameters
type JobParam struct {
	Name  string
	Value string
}

// JobParams is a list of JobParams
type JobParams struct {
	JobParams []JobParam
}

// Get gets the parameter value.
func (p *JobParams) Get(name string) *string {
	for _, param := range p.JobParams {
		if param.Name == name {
			return &param.Value
		}
	}
	return nil
}

// Job represents a schedule job.
type Job struct {
	JobID     string
	State     string
	JobParams JobParams
	NextRun   string
	UserName  string
}

// JobScheduleRequest is the specification for a schedule.
type JobScheduleRequest struct {
	JobName          string            `json:"jobName,omitempty"`
	SimpleJobTrigger *SimpleJobTrigger `json:"simpleJobTrigger,omitempty"`
	InputFile        string            `json:"inputFile"`
	OutputFile       string            `json:"outputFile,omitempty"`
	RunInBackground  bool              `json:"runInBackground"`
	JobParameters    *JobParameters    `json:"jobParameters,omitempty"`
}

// SimpleJobTrigger represents a trigger of the schedule.
type SimpleJobTrigger struct {
	UIPassParam    string `json:"uiPassParam,omitempty"`
	RepeatInterval int    `json:"repeatInterval"`
	RepeatCount    int    `json:"repeatCount"`
	StartTime      string `json:"startTime,omitempty"`
	EndTime        string `json:"endTime,omitempty"`
}

// JobParameters is the parameter of the job.
type JobParameters struct {
	Name        string `json:"name,omitempty"`
	Type        string `json:"type,omitempty"`
	StringValue string `json:"stringValue,omitempty"`
}
