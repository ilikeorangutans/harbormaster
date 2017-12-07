package azkaban

type FlowRepository interface {
	// Flow returns a flow instance for the given name and project
	Flow(Project, string) (Flow, Project, error)
	ListFlows(Project) ([]Flow, error)
}

func NewFlowRepository(client *Client) FlowRepository {
	return &ajaxFlowRepoImpl{
		azkaban: client,
	}
}

type ajaxFlowRepoImpl struct {
	azkaban *Client
}

func (r *ajaxFlowRepoImpl) ListFlows(proj Project) ([]Flow, error) {
	return r.azkaban.ListFlows(proj.Name)
}

func (r *ajaxFlowRepoImpl) Flow(proj Project, flowID string) (Flow, Project, error) {
	jobList, err := r.azkaban.FlowJobList(proj.Name, flowID)
	flow := Flow{}
	if err != nil {
		return flow, proj, err
	}

	proj.ID = jobList.ProjectID

	flow.FlowID = jobList.FlowID
	return flow, proj, err
}
