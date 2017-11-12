package azkaban

import (
	"strconv"
	"time"
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

type ListFlowsResponse struct {
	Project   string `json:"project"`
	ProjectID int    `json:"projectId"`
	Flows     []Flow `json:"flows"`
}
type Flow struct {
	FlowID string `json:"flowId"`
}

type ExecutionsList struct {
	Total      int         `json:"total"`
	Executions []Execution `json:"executions"`
}

type Execution struct {
	StartTime   AzkabanTimestamp `json:"startTime"`
	Status      Status           `json:"status"`
	ExecutionID int64            `json:"execId"`
	EndTime     AzkabanTimestamp `json:"endTime"`
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
	Nodes []FlowJob `json:"nodes"`
}

type FlowJob struct {
	ID   string   `json:"id"`
	Type string   `json:"type"`
	In   []string `json:"in"`
}

type FlowJobLog struct {
	Data   string `json:"data"`
	Length int64  `json:"length"`
	Offset int64  `json:"offset"`
}
