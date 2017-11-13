package azkaban

type ProjectRepository interface {
	ListProjects() ([]Project, error)
}

func NewProjectRepository() ProjectRepository {
	return &projectRepoImpl{}
}

type projectRepoImpl struct {
}

func (r *projectRepoImpl) ListProjects() ([]Project, error) {
	// TODO hardcoded for now because azkaban does not have an endpoint for this.
	return []Project{
		{Name: "Longboat"},
		{Name: "LongboatStaging"},
	}, nil
}
