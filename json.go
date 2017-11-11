package azkabanlib

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
	Executions []Execution `json:"executions"`
}

type Execution struct {
	StartTime   int64  `json:"startTime"`
	Status      string `json:"status"`
	ExecutionID int64  `json:"execId"`
	EndTime     int64  `json:"endTime"`
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