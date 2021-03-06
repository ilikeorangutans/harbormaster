package azkaban

import (
	"bytes"
	"fmt"
	"github.com/ilikeorangutans/harbormaster/format"
	"strconv"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/fatih/color"
)

type AzkabanTimestamp time.Time

func (t *AzkabanTimestamp) UnmarshalJSON(b []byte) error {
	i, err := strconv.ParseInt(string(b), 10, 64)
	if err != nil {
		return err
	}

	*(*time.Time)(t) = time.Unix(0, i*1000000)

	return nil
}

func (t AzkabanTimestamp) Time() time.Time {
	return time.Time(t)
}

type LoginResponse struct {
	Status    string `json:"status"`
	SessionID string `json:"session.id"`
}

type AzkabanError interface {
	AzkabanError() string
}

type AzkabanResponse struct {
	Error string `json:"error"`
}

func (r AzkabanResponse) AzkabanError() string {
	return r.Error
}

type ListFlowsResponse struct {
	AzkabanResponse
	Project   string `json:"project"`
	ProjectID int    `json:"projectId"`
	Flows     []Flow `json:"flows"`
}
type Flow struct {
	FlowID string `json:"flowId"`
}

type ExecutionsList struct {
	AzkabanResponse
	Total      int        `json:"total"`
	Executions Executions `json:"executions"`
	ProjectID  int64      `json:"projectId"`
	Project    string     `json:"project"`
}

type Executions []Execution

type ExecutionHistogram struct {
	Failures    int
	Total       int
	Successes   int
	Running     int
	Histogram   string
	EndTime     time.Time
	LastSuccess *time.Time
}

func (e Executions) Health() Health {
	health := Healthy
	currentlyRunning := false
	for _, execution := range e {
		if execution.IsSuccess() {
			health = Healthy
			break
		}
		if execution.IsRunning() {
			currentlyRunning = true
		}

		if execution.IsFailure() {
			if currentlyRunning {
				health = Concerning
			} else {
				health = Critical
			}
			break
		}
	}

	return health
}

func (e Executions) Histogram() ExecutionHistogram {
	result := ExecutionHistogram{}
	for _, execution := range e {
		if execution.IsFailure() {
			result.Failures++

			result.Histogram += color.RedString("⨉")
		} else if execution.IsSuccess() {
			if result.LastSuccess == nil {
				endTime := execution.EndTime.Time()
				result.LastSuccess = &endTime
			}
			result.Successes++
			result.Histogram += color.GreenString("•")
		} else {
			result.Running++
			result.Histogram += color.CyanString("?")
		}
	}

	result.Total = len(e)

	return result
}

func (e Executions) MostRecentExecution() Execution {
	return e[0]
}

func (e Executions) HistogramDetails(n int) []string {
	var lines []string

	upperBound := n
	if n > len(e) {
		upperBound = len(e)
	}
	for i := upperBound - 1; i >= 0; i-- {
		var buffer bytes.Buffer
		for j := 0; j < i; j++ {
			// TODO might be more efficient to only switch color code when status changes
			colorFunc := e[j].Status.ColorFunc()
			buffer.WriteString(colorFunc("│"))
		}

		execution := e[i]
		colorFunc := execution.Status.ColorFunc()

		buffer.WriteString(colorFunc("╰"))
		buffer.WriteString(colorFunc(strings.Repeat("─", upperBound-i)))
		buffer.WriteString(" ")

		buffer.WriteString(
			fmt.Sprintf(
				"%-16s %-16s %s",
				execution.Status.Colored(),
				humanize.Time(execution.StartTime.Time()),
				format.DurationHumanReadable(execution.Duration()),
			),
		)

		lines = append(lines, buffer.String())
	}

	return lines
}

type Execution struct {
	SubmitTime AzkabanTimestamp `json:"submitTime"`
	StartTime  AzkabanTimestamp `json:"startTime"`
	Status     Status           `json:"status"`
	ID         int64            `json:"execId"`
	EndTime    AzkabanTimestamp `json:"endTime"`
	ProjectID  int64            `json:"projectId"`
}

func (e Execution) IsFailure() bool {
	return e.Status == "FAILED"
}

func (e Execution) IsSuccess() bool {
	return e.Status == "SUCCEEDED"
}

func (e Execution) IsRunning() bool {
	return e.Status == "RUNNING"
}

func (e Execution) Duration() time.Duration {
	endTime := e.EndTime.Time()
	if e.IsRunning() {
		endTime = time.Now()
	}
	return endTime.Sub(e.StartTime.Time())
}

type FlowJobList struct {
	AzkabanResponse
	Nodes     []FlowJob `json:"nodes"`
	ProjectID int64     `json:"projectId"`
	FlowID    string    `json:"flow"`
}

type FlowJob struct {
	ID   string   `json:"id"`
	Type string   `json:"type"`
	In   []string `json:"in"`
	Next *FlowJob
	Prev *FlowJob
}

type FlowJobLog struct {
	Data   string `json:"data"`
	Length int64  `json:"length"`
	Offset int64  `json:"offset"`
}

type ScheduleResponse struct {
	Schedule *FlowSchedule `json:"schedule"`
}

func (s ScheduleResponse) Empty() bool {
	return s.Schedule == nil
}

type FlowSchedule struct {
	ID           string            `json:"scheduleId"`
	NextExecTime AzkabanStringTime `json:"nextExecTime"`
	Period       string            `json:"period"` // TODO make this a time.Duration
}

func (f FlowSchedule) IsScheduled() bool {
	return f.NextExecTime.Time().Unix() > 0
}

type AzkabanStringTime time.Time

func (t *AzkabanStringTime) UnmarshalJSON(b []byte) error {
	// Because azkaban for some reason runs in EST
	loc, _ := time.LoadLocation("EST")
	unquoted, err := strconv.Unquote(string(b))
	if err != nil {
		return err
	}
	x, err := time.ParseInLocation("2006-01-02 15:04:05", unquoted, loc)
	if err != nil {
		return err
	}

	*(*time.Time)(t) = x

	return nil
}

func (t AzkabanStringTime) Time() time.Time {
	return time.Time(t)
}

type FlowExecutionStatus struct {
	Attempt   int         `json:"attempt"`
	Status    Status      `json:"status"`
	ProjectID int64       `json:"projectId"`
	FlowID    string      `json:"flow"`
	Nodes     []JobStatus `json:"nodes"`
}

type JobStatus struct {
	ID     string `json:"id"`
	Status Status `json:"status"`
}

type ListAllProjectsResponse struct {
	Projects []Project `json:"projects"`
	Error    string    `json:"error"`
}

func (l *ListAllProjectsResponse) AzkabanError() string {
	return ""
}
