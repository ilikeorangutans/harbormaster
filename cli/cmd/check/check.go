package check

import (
	"bufio"
	"fmt"
	"log"
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
	project = checkFlowCmd.Arg("project", "Project").Required().String()
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

func (s *checkCmd) checkFlow(ctx *kingpin.ParseContext) error {
	fmt.Printf("Checking status of %s::%s...\n", *project, *flow)
	client := s.ctx.Client()
	flowRepo := azkaban.NewFlowRepository(s.ctx.Client())
	flow, proj, err := flowRepo.Flow(azkaban.Project{Name: *project}, *flow)
	if err != nil {
		return err
	}

	schedule, err := client.FlowSchedule(proj.ID, flow.FlowID)
	if err != nil {
		return err
	}
	executions, err := client.FlowExecutions(proj.Name, flow.FlowID)
	if err != nil {
		return err
	}

	if len(executions) == 0 {
		log.Printf("no executions")
		return nil
	}

	health := executions.Health()
	histogram := executions.Histogram()
	details := executions.HistogramDetails(*detailCountFlag)

	fmt.Printf("%-16s %s\n", "Job health:", health.Colored())
	fmt.Printf("%-16s %d failures, %d successes, %d running, %d total\n", "Stats:", histogram.Failures, histogram.Successes, histogram.Running, histogram.Total)
	lastSuccessMessage := fmt.Sprintf("none in the last %d executions", len(executions))
	if histogram.LastSuccess != nil {
		lastSuccessMessage = humanize.Time(*histogram.LastSuccess)
	}
	fmt.Printf("%-16s %s\n", "Last success:", lastSuccessMessage)

	scheduledMessage := "not scheduled"
	if schedule.IsScheduled() {
		scheduledMessage = fmt.Sprintf("%s", humanize.Time(schedule.NextExecTime.Time()))
	}
	fmt.Printf("%-16s %s\n", "Next execution:", scheduledMessage)
	fmt.Printf("Histogram:       %s\n", histogram.Histogram)
	for _, l := range details {
		fmt.Printf("%-16s %s\n", " ", l)
	}

	if health == azkaban.Critical {
		fmt.Println()

		executionID := executions.MostRecentExecution().ID

		status, err := client.FlowEcecutionStatus(executionID)
		if err != nil {
			return err
		}

		var failedJob azkaban.JobStatus
		for _, n := range status.Nodes {
			if n.Status.IsFailure() {
				failedJob = n
				break
			}
		}

		fmt.Printf("Execution failed in %q, log messages of interest:\n", failedJob.ID)
		l, err := client.ExecutionJobLog(executionID, failedJob.ID)
		if err != nil {
			return err
		}

		scanner := bufio.NewScanner(strings.NewReader(l))

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
	}

	return nil
}
