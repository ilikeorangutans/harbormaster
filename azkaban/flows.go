package azkaban

import "sort"

type Flows []Flow

func (f Flows) Len() int           { return len(f) }
func (f Flows) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }
func (f Flows) Less(i, j int) bool { return f[i].FlowID < f[j].FlowID }

type FlowRepository interface {
	// Flow returns a flow instance for the given name and project
	Flow(Project, string) (Flow, Project, error)
	ListFlows(Project, FlowPredicate) (Flows, error)
}

func NewFlowRepository(client *Client) FlowRepository {
	return &ajaxFlowRepoImpl{
		azkaban: client,
	}
}

type ajaxFlowRepoImpl struct {
	azkaban *Client
}

func (r *ajaxFlowRepoImpl) ListFlows(proj Project, predicate FlowPredicate) (Flows, error) {
	flows, err := r.azkaban.ListFlows(proj.Name)
	if err != nil {
		return nil, err
	}

	var result Flows
	for _, flow := range flows {
		if predicate(flow) {
			result = append(result, flow)
		}
	}

	sort.Sort(result)

	return result, nil
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
