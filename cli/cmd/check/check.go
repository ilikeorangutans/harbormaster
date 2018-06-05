package check

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"

	humanize "github.com/dustin/go-humanize"
	"github.com/ilikeorangutans/harbormaster/azkaban"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var cmd *kingpin.CmdClause
var checkFlowCmd *kingpin.CmdClause
var project *string
var flow *string
var getClient func() *azkaban.Client
var checkCountFlag *int
var detailCountFlag *int

func ConfigureCommand(app *kingpin.Application, ctx *azkaban.Context) {
	s := &checkCmd{
		ctx: ctx,
	}
	cmd = app.Command("check", "")
	checkFlowCmd = cmd.Command("flow", "checks a flow").Action(s.checkFlow)
	project = checkFlowCmd.Flag("project", "Project").Envar("HARBORMASTER_PROJECT").Required().String()
	flow = checkFlowCmd.Arg("flow", "Flow").HintAction(s.suggestFlow).Required().String()
	checkCountFlag = checkFlowCmd.Flag("execution-count", "number of executions to check").Short('n').Default("20").Int()
	detailCountFlag = checkFlowCmd.Flag("detail-count", "number of execution details").Short('d').Default("5").Int()
}

type checkCmd struct {
	ctx *azkaban.Context
}

func (s *checkCmd) suggestFlow() []string {
	log.Printf("checkCmd.suggestFlow %s %s...\n", *project, *flow)
	flows, err := s.ctx.Flows().ListFlows(azkaban.Project{Name: *project})
	if err != nil {
		fmt.Printf("error retrieving flows: %s\n", err.Error())
		return []string{}
	}
	var result []string
	log.Printf("Found %d\n", len(flows))
	for _, f := range flows {
		result = append(result, f.FlowID)
	}
	return result
}

func (s *checkCmd) printSchedule(proj azkaban.Project, flow azkaban.Flow) error {
	client := s.ctx.Client()
	schedule, err := client.FlowSchedule(proj.ID, flow.FlowID)
	if err != nil {
		return err
	}
	scheduledMessage := "not scheduled"
	if schedule.IsScheduled() {
		scheduledMessage = fmt.Sprintf("%s", humanize.Time(schedule.NextExecTime.Time()))
	}
	fmt.Printf("%-16s %s\n", "Next execution:", scheduledMessage)
	return nil
}

type FlowStatus struct {
	Project       azkaban.Project
	Flow          azkaban.Flow
	Health        azkaban.Health
	LastExecution azkaban.Execution
	FailedJob     azkaban.JobStatus
}

func (s *checkCmd) printFlowStatus(proj azkaban.Project, flow azkaban.Flow) (status FlowStatus, err error) {
	fmt.Printf("Checking status of %s %s...\n", proj.Name, flow.FlowID)
	client := s.ctx.Client()
	executions, err := client.FlowExecutions(proj.Name, flow.FlowID)
	if err != nil {
		return status, err
	}

	if len(executions) == 0 {
		log.Printf("no executions")
		return status, nil
	}

	status.Health = executions.Health()
	histogram := executions.Histogram()
	details := executions.HistogramDetails(*detailCountFlag)

	fmt.Printf("%-16s %s\n", "Job health:", status.Health.Colored())
	fmt.Printf("%-16s %d failures, %d successes, %d running, %d total\n", "Stats:", histogram.Failures, histogram.Successes, histogram.Running, histogram.Total)
	lastSuccessMessage := fmt.Sprintf("none in the last %d executions", len(executions))
	if histogram.LastSuccess != nil {
		lastSuccessMessage = humanize.Time(*histogram.LastSuccess)
	}
	fmt.Printf("%-16s %s\n", "Last success:", lastSuccessMessage)

	if err := s.printSchedule(proj, flow); err != nil {
		return status, err
	}

	fmt.Printf("Histogram:       %s\n", histogram.Histogram)
	for _, l := range details {
		fmt.Printf("%-16s %s\n", " ", l)
	}

	if status.Health == azkaban.Critical {
		fmt.Println()

		status.LastExecution = executions.MostRecentExecution()
		executionID := executions.MostRecentExecution().ID

		flowExecstatus, err := client.FlowEcecutionStatus(executionID)
		if err != nil {
			return status, err
		}

		for _, n := range flowExecstatus.Nodes {
			if n.Status.IsFailure() {
				status.FailedJob = n
				break
			}
		}
	}

	return status, nil
}

func (s *checkCmd) checkFlow(ctx *kingpin.ParseContext) error {
	flowRepo := azkaban.NewFlowRepository(s.ctx.Client())
	flow, proj, err := flowRepo.Flow(azkaban.Project{Name: *project}, *flow)
	if err != nil {
		return err
	}

	status, err := s.printFlowStatus(proj, flow)
	if err != nil {
		return err
	}

	if status.Health == azkaban.Critical {
		fmt.Printf("Execution failed in %q, log messages of interest:\n", status.FailedJob.ID)
		client := s.ctx.Client()
		buffer := bytes.NewBuffer([]byte{})
		_, err := client.FetchLogsUntilEnd(status.LastExecution.ID, status.FailedJob.ID, 0, buffer)
		if err != nil {
			return err
		}

		scanner := bufio.NewScanner(strings.NewReader(buffer.String()))

		patterns := []string{
			"err",
			"exception",
			"failed",
			"failure",
		}
		fmt.Println(strings.Repeat("-", 80))
		for scanner.Scan() {
			line := scanner.Text()
			lower := strings.ToLower(line)
			ofInterest := false
			for _, p := range patterns {
				if strings.Contains(lower, p) {
					ofInterest = true
					break
				}
			}
			if ofInterest {
				fmt.Println(line)
			}
		}
		fmt.Println(strings.Repeat("-", 80))
		fmt.Println()

		scanner = bufio.NewScanner(os.Stdin)
		run := true

		printActionTerm()
		for run {
			printActionTerm()
			run = run && scanner.Scan()
			input := strings.ToLower(strings.TrimSpace(scanner.Text()))

			if input == "restart" {
				client.RestartFlowNow(proj.Name, flow.FlowID)
			} else if input == "logs" {
				// TODO this might be slow:
				// fmt.Println(l)
				_, err = client.FetchLogsUntilEnd(status.LastExecution.ID, status.FailedJob.ID, 0, os.Stdout)
				if err != nil {
					return err
				}
			} else if input == "status" {
				status, err = s.printFlowStatus(proj, flow)
				if err != nil {
					return err
				}
			} else if input == "unschedule" {
				fmt.Println("not implemented yet")
			} else {
				run = false
			}
		}

	}

	return nil
}

func printActionTerm() {
	fmt.Println("Actions: do nothing|status|restart|unschedule|logs")
	fmt.Print("> ")
}
