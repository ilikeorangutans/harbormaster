package azkaban

type ProjectRepository interface {
	ListProjects() ([]Project, error)
}

func NewProjectRepository(client *Client) ProjectRepository {
	return &projectRepoImpl{
		client: client,
	}
}

type projectRepoImpl struct {
	client *Client
}

func (r *projectRepoImpl) ListProjects() ([]Project, error) {
	return r.client.ListProjects()
}
