package client

import (
	"sort"
	"time"
)

// Status represents a status of carte job or transformations.
type Status interface {
	Print(w *IndentWriter)
	IsFinished() bool
}

// StepStatus represents the status of steps.
type StepStatus struct {
	Name              string  `xml:"stepname"`
	Copy              int     `xml:"copy"`
	LinesRead         int     `xml:"linesRead"`
	LinesWritten      int     `xml:"linesWritten"`
	LinesInput        int     `xml:"linesInput"`
	LinesOutput       int     `xml:"linesOutput"`
	LinesUpdated      int     `xml:"linesUpdated"`
	LinesRejected     int     `xml:"linesRejected"`
	Errors            int     `xml:"errors"`
	StatusDescription string  `xml:"statusDescription"`
	Seconds           float32 `xml:"seconds"`
	Speed             string  `xml:"speed"`
	Priority          string  `xml:"priority"`
	Stopped           string  `xml:"stopped"`
	Paused            string  `xml:"paused"`
}

func (s *StepStatus) print(writer *IndentWriter) {
	writer.Printf("Name              : %s\n", s.Name)
	writer.Printf("Status Description: %s\n", s.StatusDescription)
	writer.Printf("Seconds           : %f\n", s.Seconds)
	writer.Printf("Speed             : %s\n", s.Speed)
	writer.Printf("Errors            : %d\n", s.Errors)
	writer.Printf("Copy              : %d\n", s.Copy)
	writer.Printf("Lines(I/O/R/W)    : %d/%d/%d/%d\n", s.LinesInput, s.LinesOutput, s.LinesRead, s.LinesWritten)
	writer.Printf("Lines(Upd/Rej): %d/%d\n", s.LinesUpdated, s.LinesRejected)
	writer.Printf("Priority          : %s\n", s.Priority)
	writer.Printf("Stopped           : %s\n", s.Stopped)
	writer.Printf("Paused            : %s\n", s.Paused)
}

// StepStatusList represents the status of steps.
type StepStatusList struct {
	List []StepStatus `xml:"stepstatus"`
}

func (r *StepStatusList) print(writer *IndentWriter) {
	for i, s := range r.List {
		writer.Printf("Step %d:\n", i+1)
		writer.IncrementLevel()
		s.print(writer)
		writer.DecrementLevel()
	}
}

// Result represents the result of the job or transformation
type Result struct {
	LinesInput     int    `xml:"lines_input"`
	LinesOutput    int    `xml:"lines_output"`
	LinesRead      int    `xml:"lines_read"`
	LinesWritten   int    `xml:"lines_written"`
	LinesUpdated   int    `xml:"lines_updated"`
	LinesRejected  int    `xml:"lines_rejected"`
	LinesDeleted   int    `xml:"lines_deleted"`
	Errors         int    `xml:"nr_errors"`
	FilesRetrieved int    `xml:"nr_files_retrieved"`
	Entry          int    `xml:"entry"`
	Result         string `xml:"result"`
	ExitStatus     int    `xml:"exit_status"`
	IsStopped      string `xml:"is_stopped"`
	LogChannelID   string `xml:"log_channel_id"`
	LogText        string `xml:"log_text"`
}

func (r *Result) print(writer *IndentWriter) {
	writer.Printf("Result            : %s\n", r.Result)
	writer.Printf("Exit Status       : %d\n", r.ExitStatus)
	writer.Printf("Stopped           : %s\n", r.IsStopped)
	writer.Printf("Log Channel ID    : %s\n", r.LogChannelID)

	writer.Println("Log Text          :")
	writer.IncrementLevel()
	writer.PrintMultiline(r.LogText)
	writer.DecrementLevel()

	writer.Printf("Lines(I/O/R/W)    : %d/%d/%d/%d\n", r.LinesInput, r.LinesOutput, r.LinesRead, r.LinesWritten)
	writer.Printf("Lines(Upd/Rej/Del): %d/%d/%d\n", r.LinesUpdated, r.LinesRejected, r.LinesDeleted)
	writer.Printf("Errors            : %d\n", r.Errors)
	writer.Printf("Retrieved Files   : %d\n", r.FilesRetrieved)
	writer.Printf("Entry             : %d\n", r.FilesRetrieved)
}

// BaseStatus is a base struct commonize the job and transformation.
type BaseStatus struct {
	ID                string `xml:"id"`
	StatusDescription string `xml:"status_desc"`
	ErrorDescription  string `xml:"error_desc"`
	LogDate           string `xml:"log_date"`
	FirstLogLineNr    int    `xml:"first_log_line_nr"`
	LastLogLineNr     int    `xml:"last_log_line_nr"`
	Result            Result `xml:"result"`
	LoggingString     string `xml:"logging_string"`
}

// IsFinished check if the job has finished.
func (s *BaseStatus) IsFinished() bool {
	return s.StatusDescription == "Finished"
}

// ParseLogDate parses log
func (s *BaseStatus) ParseLogDate() time.Time {
	time, _ := time.Parse("2006/01/02 15:04:05.000", s.LogDate)
	return time
}

func (s *BaseStatus) print(writer *IndentWriter, name string) {
	writer.Printf("ID  : %s\n", s.ID)
	writer.Printf("Name: %s\n", name)
	writer.Printf("Status: %s\n", s.StatusDescription)
	writer.Printf("Error : %s\n", s.ErrorDescription)
	writer.Printf("Date : %s\n", s.LogDate)

	writer.Println("Result:")
	writer.IncrementLevel()
	s.Result.print(writer)
	writer.DecrementLevel()

	writer.Println("Log:")
	writer.IncrementLevel()
	writer.Printf("First Line : %d\n", s.FirstLogLineNr)
	writer.Printf("Last Line  : %d\n", s.LastLogLineNr)
	writer.PrintMultiline(s.LoggingString)
	writer.DecrementLevel()
}

// TransformationStatus represents the status of the transformations.
type TransformationStatus struct {
	BaseStatus
	Name           string         `xml:"transname"`
	Paused         string         `xml:"paused"`
	StepStatusList StepStatusList `xml:"stepstatuslist"`
}

// Print the status of the transformation.
func (t *TransformationStatus) Print(writer *IndentWriter) {
	t.BaseStatus.print(writer, t.Name)
	writer.Printf("Paused: %s\n", t.Paused)
	writer.Println("Steps:")

	writer.IncrementLevel()
	t.StepStatusList.print(writer)
	writer.DecrementLevel()
}

// TransformationStatusList represents the status list of transformations.
type TransformationStatusList struct {
	List []TransformationStatus `xml:"transstatus"`
}

// JobStatus represents the status of jobs.
type JobStatus struct {
	BaseStatus
	Name string `xml:"jobname"`
}

// Print the status of the transformation.
func (t *JobStatus) Print(writer *IndentWriter) {
	t.BaseStatus.print(writer, t.Name)
}

// JobStatusList represents the status list of the jobs.
type JobStatusList struct {
	List []JobStatus `xml:"jobstatus"`
}

// CarteServerStatus represents the status of carte server.
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

// SortStatusByLogDate sorts the status by log date.
func (s *CarteServerStatus) SortStatusByLogDate() {
	sort.Slice(s.JobStatusList.List, func(i, j int) bool {
		return s.JobStatusList.List[i].ParseLogDate().UnixNano() > s.JobStatusList.List[j].ParseLogDate().UnixNano()
	})
	sort.Slice(s.TransformationStatusList.List, func(i, j int) bool {
		return s.TransformationStatusList.List[i].ParseLogDate().UnixNano() > s.TransformationStatusList.List[j].ParseLogDate().UnixNano()
	})
}
