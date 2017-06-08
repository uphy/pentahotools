package pentahoclient

import (
	"errors"
	"fmt"
	"sort"
	"time"
)

func (c *Client) GetCarteServerStatus() (*CarteServerStatus, error) {
	var status CarteServerStatus
	resp, err := c.client.R().
		SetQueryParam("xml", "Y").
		SetResult(&status).
		Get(fmt.Sprintf("kettle/status"))
	switch resp.StatusCode() {
	case 200:
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

type StepStatus struct {
}

type StepStatusList struct {
	List []StepStatus `stepstatus`
}

type BaseStatus struct {
	ID                string `xml:"id"`
	StatusDescription string `xml:"status_desc"`
	ErrorDescription  string `xml:"error_desc"`
	LogDate           string `xml:"log_date"`
	LoggingString     string `xml:"logging_string"`
	FirstLogLineNr    int    `xml:"first_log_line_nr"`
	LastLogLineNr     int    `xml:"last_log_line_nr"`
}

func (s *BaseStatus) ParseLogDate() time.Time {
	time, _ := time.Parse("2006/01/02 15:04:05.000", s.LogDate)
	return time
}

type TransformationStatus struct {
	BaseStatus
	Name string `xml:"transname"`

	Paused         string         `xml:"paused"`
	StepStatusList StepStatusList `xml:"stepstatuslist"`
}

type TransformationStatusList struct {
	List []TransformationStatus `xml:"transstatus"`
}

type JobStatus struct {
	BaseStatus
	Name              string `xml:"jobname"`
	ID                string `xml:"id"`
	StatusDescription string `xml:"status_desc"`
	ErrorDescription  string `xml:"error_desc"`
	LogDate           string `xml:"log_date"`
	LoggingString     string `xml:"logging_string"`
	FirstLogLineNr    int    `xml:"first_log_line_nr"`
	LastLogLineNr     int    `xml:"last_log_line_nr"`
}

type JobStatusList struct {
	List []JobStatus `xml:"jobstatus"`
}

type CarteServerStatus struct {
	StatusDescription        string                   `xml:"statusdesc"`
	MemoryFree               int64                    `xml:"memory_free"`
	MemoryTotal              int64                    `xml:"memory_total"`
	CPUCores                 int16                    `xml:"cpu_cores"`
	CPUProcessTime           int64                    `xml:"cpu_process_time"`
	UpTime                   int64                    `xml:"uptime"`
	ThreadCount              int32                    `xml:"thread_count"`
	LoadAverage              float64                  `xml:"load_avg"`
	OSName                   string                   `xml:"os_name"`
	OSVersion                string                   `xml:"os_version"`
	OSArch                   string                   `xml:"os_arch"`
	TransformationStatusList TransformationStatusList `xml:"transstatuslist"`
	JobStatusList            JobStatusList            `xml:"jobstatuslist"`
}

func (s *CarteServerStatus) SortStatusByLogDate() {
	sort.Slice(s.JobStatusList.List, func(i, j int) bool {
		return s.JobStatusList.List[i].ParseLogDate().UnixNano() > s.JobStatusList.List[j].ParseLogDate().UnixNano()
	})
	sort.Slice(s.TransformationStatusList.List, func(i, j int) bool {
		return s.TransformationStatusList.List[i].ParseLogDate().UnixNano() > s.TransformationStatusList.List[j].ParseLogDate().UnixNano()
	})
}
