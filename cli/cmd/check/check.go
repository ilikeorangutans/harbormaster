package check

import (
	"fmt"
	"log"

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

func ConfigureCommand(app *kingpin.Application, ctx *azkaban.Context) {
	s := &checkCmd{
		ctx: ctx,
	}
	cmd = app.Command("check", "")
	checkFlowCmd = cmd.Command("flow", "checks a flow").Action(s.checkFlow)
	project = checkFlowCmd.Arg("project", "Project").Required().String()
	flow = checkFlowCmd.Arg("flow", "Flow").HintAction(s.suggestFlow).Required().String()
	checkCountFlag = checkFlowCmd.Flag("n", "number of executions to check").Default("20").Int()
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
	details := executions.HistogramDetails(5)

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

	return nil
}
