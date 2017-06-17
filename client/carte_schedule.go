package client

import (
	"errors"
	"fmt"

	"go.uber.org/zap"
)

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

// Job represents a schedule job.
type Job struct {
	JobID     string
	State     string
	JobParams JobParams
	NextRun   string
	UserName  string
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
