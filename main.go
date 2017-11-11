package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	humanize "github.com/dustin/go-humanize"
	"github.com/ilikeorangutans/harbormaster/azkaban"
	"github.com/ilikeorangutans/harbormaster/cli/cmd/check"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

const (
	AzkabanSessionIDEnv = "AZKABAN_SESSION_ID"
	AzkabanHostEnv      = "AZKABAN_HOST"
)

var (
	app = kingpin.New("harbormaster", "a thingy that makes working with the boats easier")

	login         = app.Command("login", "authenticate against azkaban and sets session id as environment variable")
	loginHost     = login.Arg("host", "username").Required().URL()
	loginUsername = login.Arg("username", "username").Required().String()
	loginPassword = login.Arg("password", "password").Required().String()

	projects     = app.Command("project", "")
	projectsList = projects.Command("list", "")

	flows        = app.Command("flow", "")
	flowsList    = flows.Command("list", "")
	flowsProject = flowsList.Arg("project", "").Required().HintAction(suggestProjects).String()
	flowsFilter  = flowsList.Arg("filter", "").String()

	executions        = app.Command("executions", "")
	executionsProject = executions.Arg("project", "").Required().HintAction(suggestProjects).String()
	executionsFlow    = executions.Arg("flow", "").Required().String()

	logs       = app.Command("logs", "")
	logsJobID  = logs.Arg("jobID", "").Required().HintAction(suggestProjects).String()
	logsExecID = logs.Arg("execID", "").Required().HintAction(suggestExecID).Int()
)

func suggestFlow() []string {
	client := getClient()
	flows, err := client.ProjectFlows(*executionsProject)
	if err != nil {
		log.Printf("Error retrieving list of flows: %s", err.Error())
		return nil
	}
	result := []string{}
	for _, f := range flows {
		result = append(result, f.FlowID)
	}
	return result
}

func suggestExecID() []string {
	return []string{"1", "2", "3"}
}

func main() {

	check.ConfigureCommand(app, getClient)

	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case login.FullCommand():
		client, err := azkaban.ConnectWithUsernameAndPassword((*loginHost).String(), *loginUsername, *loginPassword)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		fmt.Printf("export %s=%s\n", AzkabanSessionIDEnv, client.SessionID)
		fmt.Printf("export %s=%s\n", AzkabanHostEnv, (*loginHost).String())
	case projectsList.FullCommand():

	case flowsList.FullCommand():
		client := getClient()

		flows, err := client.ProjectFlows(*flowsProject)
		if err != nil {
			panic(err)
		}

		for _, f := range flows {
			if len(*flowsFilter) > 0 {
				if strings.HasPrefix(f.FlowID, *flowsFilter) {
					fmt.Printf("%s\n", f.FlowID)
				}
			} else {
				fmt.Printf("%s\n", f.FlowID)
			}
		}

	case executions.FullCommand():
		client := getClient()

		executions, err := client.FlowExecutions(*executionsProject, *executionsFlow)
		if err != nil {
			panic(err)
		}

		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 8, 4, 1, '\t', 0)
		for _, e := range executions {
			start := time.Unix(0, e.StartTime*1000000)
			fmt.Fprintf(w, "%d \t %s \t %s \t %s\n", e.ExecutionID, e.Status, start.Format(time.RFC1123), humanize.Time(start))
		}
		w.Flush()

	case logs.FullCommand():
		client := getClient()

		log, err := client.ExecutionJobLog(*logsExecID, *logsJobID)
		if err != nil {
			panic(err)
		}

		os.Stdout.WriteString(log)

	}
}

var client *azkaban.Client

func getClient() *azkaban.Client {
	if client != nil {
		return client
	}

	sessionID := os.Getenv(AzkabanSessionIDEnv)
	host := os.Getenv(AzkabanHostEnv)
	var err error
	client, err = azkaban.ConnectWithSessionID(host, sessionID)
	if err != nil {
		panic(err)
	}

	return client
}

func suggestProjects() []string {
	// TODO: there's no way of retrieving a list of projects from azkaban.
	return []string{"Longboat", "LongboatStaging"}
}
