package azkaban

type ExecutionRepository interface {
	ListExecutions(Project, Flow, Paginator) (Executions, error)
}

func NewExecutionRepository(client *Client) ExecutionRepository {
	return &ajaxExecRepo{
		azkaban: client,
	}
}

type ajaxExecRepo struct {
	azkaban *Client
}

func (r *ajaxExecRepo) ListExecutions(proj Project, f Flow, p Paginator) (Executions, error) {
	return r.azkaban.FlowExecutions(proj.Name, f.FlowID, p)
}
