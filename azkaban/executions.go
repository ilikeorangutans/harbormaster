package azkaban

type ExecutionRepository interface {
	ListExecutions(Project, Flow, Paginator) ([]Execution, error)
}

func NewExecutionRepository(client *Client) ExecutionRepository {
	return &ajaxExecRepo{
		azkaban: client,
	}
}

type ajaxExecRepo struct {
	azkaban *Client
}

func (r *ajaxExecRepo) ListExecutions(proj Project, f Flow, p Paginator) ([]Execution, error) {
	return r.azkaban.FlowExecutions(proj.Name, f.FlowID)
}
