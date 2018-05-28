package main

import (
	"github.com/ilikeorangutans/harbormaster/azkaban"
)

func suggestProjects() []string {
	r := azkaban.NewProjectRepository(nil)
	projects, _ := r.ListProjects()
	var result []string

	for _, p := range projects {
		result = append(result, p.Name)
	}

	return result
}
