package azkaban

import (
	"errors"
	"fmt"
)

type ProjectRepository interface {
	ListProjects() ([]Project, error)
	ByName(name string) (Project, error)
}

func NewProjectRepository(client *Client) ProjectRepository {
	return &projectRepoImpl{
		client: client,
	}
}

type projectRepoImpl struct {
	client *Client
}

func (r *projectRepoImpl) ByName(name string) (Project, error) {
	projects, err := r.ListProjects()
	if err != nil {
		return Project{}, err
	}

	for _, p := range projects {
		if p.Name == name {
			return p, nil
		}
	}
	return Project{}, errors.New(fmt.Sprintf("no project with name %s", name))
}

func (r *projectRepoImpl) ListProjects() ([]Project, error) {
	return r.client.ListProjects()
}
