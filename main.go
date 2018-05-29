package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	humanize "github.com/dustin/go-humanize"
	"github.com/ilikeorangutans/harbormaster/azkaban"
	"github.com/ilikeorangutans/harbormaster/cli/cmd/check"
	"github.com/ilikeorangutans/harbormaster/cli/cmd/report"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

const (
	AzkabanSessionIDEnv = "AZKABAN_SESSION_ID"
	AzkabanHostEnv      = "AZKABAN_HOST"
)

var (
	app = kingpin.New("harbormaster", "a thingy that makes working with the boats easier")

	dumpResponses = app.Flag("dump-responses", "Dump all HTTP responses from Azkaban to stdout").Bool()

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

	logs       = app.Command("logs", "Fetchs logs for an execution, either via URL or by job and exec ID")
	logURL     = logs.Arg("url", "execution URL").URL()
	logsJobID  = logs.Arg("jobID", "job ID").HintAction(suggestProjects).String()
	logsExecID = logs.Arg("execID", "exec ID").HintAction(suggestExecID).Int64()
	logsFollow = logs.Flag("follow", "follow log, indefinitely updates every 2 seconds").Short('f').Default("false").Bool()
)

func suggestExecID() []string {
	return []string{"1", "2", "3"}
}

func main() {

	context := azkaban.NewContext(getClient())
	check.ConfigureCommand(app, context)
	report.ConfigureCommand(app, context)

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
		projectRepo := context.Projects()
		projects, err := projectRepo.ListProjects()
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 8, 4, 1, '\t', 0)
		fmt.Fprintln(w, "ID \t Name")

		for _, project := range projects {
			fmt.Fprintf(
				w,
				"%d \t %s \n",
				project.ID,
				project.Name,
			)
		}
		w.Flush()

	case flowsList.FullCommand():
		client := getClient()

		flows, err := client.ListFlows(*flowsProject)
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
			fmt.Fprintf(
				w,
				"%d \t %s \t %s \t%s \t %s\n",
				e.ID,
				e.Status.Colored(),
				e.StartTime.Time().Format(time.RFC1123),
				e.Duration(),
				humanize.Time(e.StartTime.Time()),
			)
		}
		w.Flush()

	case logs.FullCommand():
		client := getClient()

		var projectID string
		var execID int64

		if *logURL != nil {
			query := (*logURL).Query()
			execID, _ = strconv.ParseInt(query.Get("execid"), 10, 64)
			projectID = query.Get("job")
		} else {
			projectID = *logsJobID
			execID = *logsExecID
		}

		lastOffset, err := client.FetchLogsUntilEnd(execID, projectID, 0, os.Stdout)
		if err != nil {
			panic(err)
		}
		// if follow sleep and fetch more

		follow := *logsFollow
		offset := lastOffset
		if follow {
			ticker := time.NewTicker(time.Second * 2)

			for {
				select {
				case <-ticker.C:
					offset, err = client.FetchLogsUntilEnd(execID, projectID, offset, os.Stdout)
					if err != nil {
						panic(err)
					}
				}
			}
		}
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

	client.DumpResponses = *dumpResponses

	return client
}
