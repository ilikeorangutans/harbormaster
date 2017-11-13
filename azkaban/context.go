package azkaban

func NewContext(client *Client) *Context {
	return &Context{
		client: client,
	}
}

type Context struct {
	client *Client
}

func (c Context) Client() *Client {
	return c.client
}

func (c Context) Projects() ProjectRepository {
	return NewProjectRepository()
}

func (c Context) Flows() FlowRepository {
	return NewFlowRepository(c.client)
}

func (c Context) Executions() ExecutionRepository {
	return NewExecutionRepository(c.client)
}
