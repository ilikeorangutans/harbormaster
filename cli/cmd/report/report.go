package report

import (
	"fmt"
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

	reportCmd := app.Command("report", "")
	execReport := reportCmd.Command("exec-time", "average execution time as tab-delimited list").Action(cmd.execReport)
	project = execReport.Arg("project", "").Required().String()
	flowFilter = execReport.Arg("flow-prefix", "prefix filter").String()
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
		fmt.Printf("%s \t %4d \t%s\n", d.FlowID, d.SuccessCount, formatDuration(d.AverageTime))
	}

	return nil
}

func formatDuration(d time.Duration) string {
	hours := 0
	if d.Hours() >= 1.0 {
		hours = int(d.Hours())
	}
	minutes := 0
	if d.Minutes() > 0 {
		minutes = int(d.Minutes()) % 60
	}
	seconds := 0
	if d.Seconds() > 0 {
		seconds = int(d.Seconds()) % 60
	}
	return fmt.Sprintf("%d:%d:%d", hours, minutes, seconds)
}

type execReportData struct {
	FlowID       string
	SuccessCount int
	TotalCount   int
	AverageTime  time.Duration
}
