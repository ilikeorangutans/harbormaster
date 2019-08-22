package main

import (
	"github.com/ilikeorangutans/harbormaster/azkaban"
	"os"
)

type ActionWithContext struct {
	DumpResponses bool
	client        *azkaban.Client
	context       *azkaban.Context
}

func (a *ActionWithContext) Context() *azkaban.Context {
	if a.context != nil {
		return a.context
	}
	a.context = azkaban.NewContext(a.Client())
	return a.context
}

func (a *ActionWithContext) Client() *azkaban.Client {
	if a.client != nil {
		return a.client
	}

	sessionID := os.Getenv(AzkabanSessionIDEnv)
	host := os.Getenv(AzkabanHostEnv)
	var err error
	a.client, err = azkaban.ConnectWithSessionID(host, sessionID)
	if err != nil {
		panic(err)
	}

	a.client.DumpResponses = a.DumpResponses

	return a.client
}
