package report

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/ilikeorangutans/harbormaster/azkaban"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	project    *string
	flowFilter *string
)

func ConfigureCommand(app *kingpin.Application, ctx *azkaban.Context) {
	cmd := &execReportCmd{
		ctx: ctx,
	}
	log.Println(cmd)

	reportCmd := app.Command("report", "")
	execReport := reportCmd.Command("exec-time", "").Action(cmd.execReport)
	project = execReport.Arg("project", "").Required().String()
	flowFilter = execReport.Arg("flow-prefix", "").String()
}

type execReportCmd struct {
	ctx *azkaban.Context
}

func (e *execReportCmd) execReport(ctx *kingpin.ParseContext) error {
	proj := azkaban.Project{Name: *project}
	flows, err := e.ctx.Flows().ListFlows(proj)
	if err != nil {
		return err
	}

	var filteredFlows []azkaban.Flow
	if len(*flowFilter) > 0 {
		for _, f := range flows {
			if strings.HasPrefix(f.FlowID, *flowFilter) {
				filteredFlows = append(filteredFlows, f)
			}
		}
	} else {
		filteredFlows = flows
	}

	fmt.Printf("Fetching %d flows...\n", len(filteredFlows))

	data := []execReportData{}
	execRepo := e.ctx.Executions()

	for _, f := range filteredFlows {
		executions, err := execRepo.ListExecutions(proj, f, azkaban.TenMostRecent)
		if err != nil {
			return err
		}

		execData := execReportData{
			FlowID: f.FlowID,
		}

		var totalTimeSeconds time.Duration = 0.0
		for _, exec := range executions {
			execData.TotalCount++
			if !exec.IsSuccess() {
				continue
			}
			execData.SuccessCount++

			totalTimeSeconds += exec.Duration()
		}

		if execData.SuccessCount > 0 {
			execData.AverageTime = time.Duration(totalTimeSeconds / time.Duration(execData.SuccessCount))
		}
		data = append(data, execData)
	}

	fmt.Printf("FlowID \tSuccess Count \t Average Time\n")
	for _, d := range data {
		fmt.Printf("%s \t %4d \t%s\n", d.FlowID, d.SuccessCount, d.AverageTime)
	}

	return nil
}

type execReportData struct {
	FlowID       string
	SuccessCount int
	TotalCount   int
	AverageTime  time.Duration
}
