package azkaban

type ExecutionRepository interface {
	ListExecutions(Project, Flow, Paginator) ([]Execution, error)
}
