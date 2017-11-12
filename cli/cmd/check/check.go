package check

import (
	"fmt"
	"log"
	"time"

	humanize "github.com/dustin/go-humanize"
	"github.com/fatih/color"
	"github.com/ilikeorangutans/harbormaster/azkaban"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var cmd *kingpin.CmdClause
var checkFlowCmd *kingpin.CmdClause
var project *string
var flow *string
var getClient func() *azkaban.Client
var checkCountFlag *int

func ConfigureCommand(app *kingpin.Application, getClientFunc func() *azkaban.Client) {
	getClient = getClientFunc
	cmd = app.Command("check", "")
	checkFlowCmd = cmd.Command("flow", "checks a flow").Action(checkFlow)
	project = checkFlowCmd.Arg("project", "Project").Required().String()
	flow = checkFlowCmd.Arg("flow", "Flow").Required().String()
	checkCountFlag = checkFlowCmd.Flag("n", "number of executions to check").Default("20").Int()
}

func checkFlow(ctx *kingpin.ParseContext) error {
	fmt.Printf("Checking status of %s::%s...\n", *project, *flow)
	client := getClient()

	executions, err := client.FlowExecutions(*project, *flow)
	if err != nil {
		return err
	}

	if len(executions) == 0 {
		log.Printf("no executions")
		return nil
	}

	health := Healthy
	failures := 0
	successes := 0
	running := 0
	histogram := ""
	var lastSuccess *time.Time

	for _, e := range executions {
		if e.IsFailure() {
			failures++
			histogram += color.RedString("X")
		} else if e.IsSuccess() {
			if lastSuccess == nil {
				endTime := e.EndTime.Time()
				lastSuccess = &endTime
			}
			successes++
			histogram += color.GreenString(".")
		} else {
			running++
			histogram += color.CyanString("?")
		}
	}

	currentlyRunning := false
	for _, e := range executions {
		if e.IsSuccess() {
			health = Healthy
			break
		}
		if e.IsRunning() {
			currentlyRunning = true
		}

		if e.IsFailure() {
			if currentlyRunning {
				health = Unhealthy
			} else {
				health = Broken
			}
			break
		}

	}

	fmt.Printf("%-16s %s\n", "Job health:", health.Colored())
	fmt.Printf("%-16s %d failures, %d successes, %d running, %d total\n", "Stats:", failures, successes, running, len(executions))
	lastSuccessMessage := fmt.Sprintf("none in the last %d executions", len(executions))
	if lastSuccess != nil {
		lastSuccessMessage = humanize.Time(*lastSuccess)
	}
	fmt.Printf("%-16s %s\n", "Last success:", lastSuccessMessage)
	fmt.Printf("Histogram:       %s\n", histogram)

	return nil
}

type Health string

func (h Health) Colored() string {
	switch h {
	case Healthy:
		return color.GreenString(string(h))
	case Unhealthy:
		return color.YellowString(string(h))
	case Broken:
		return color.RedString(string(h))
	default:
		return string(h)

	}
}

const (
	Healthy   Health = "healthy"
	Unhealthy        = "unhealthy"
	Broken           = "broken"
)
