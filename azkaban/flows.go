package azkaban

type FlowRepository interface {
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
