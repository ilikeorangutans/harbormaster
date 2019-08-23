package cli

import (
	"fmt"
	"github.com/ilikeorangutans/harbormaster/azkaban"
	"github.com/spf13/cobra"
	"log"
	"strings"
	"time"
)

func NewReportCmd(context Context) *cobra.Command {
	reportCmd := &cobra.Command{
		Use:   "report",
		Short: "average execution time as tab-delimited list",
		Run: func(cmd *cobra.Command, args []string) {
			proj := azkaban.Project{Name: context.Project()}
			flows, err := context.Context().Flows().ListFlows(proj)
			if err != nil {
				log.Fatal(err)
			}

			var filteredFlows []azkaban.Flow
			if len(args) > 0 {
				for _, f := range flows {
					if strings.HasPrefix(f.FlowID, args[0]) {
						filteredFlows = append(filteredFlows, f)
					}
				}
			} else {
				filteredFlows = flows
			}

			var data []execReportData
			execRepo := context.Context().Executions()

			for _, f := range filteredFlows {
				executions, err := execRepo.ListExecutions(proj, f, azkaban.TenMostRecent)
				if err != nil {
					log.Fatal(err)
				}

				execData := execReportData{
					FlowID: f.FlowID,
				}

				var totalTime time.Duration = 0.0
				for _, exec := range executions {
					execData.TotalCount++
					if !exec.IsSuccess() {
						continue
					}
					execData.SuccessCount++

					totalTime += exec.Duration()
				}

				if execData.SuccessCount > 0 {
					execData.AverageTime = totalTime / time.Duration(execData.SuccessCount)
				}
				data = append(data, execData)
			}

			fmt.Printf("FlowID \tSuccess Count \t Average Time\n")
			for _, d := range data {
				fmt.Printf("%s \t %4d \t%s\n", d.FlowID, d.SuccessCount, formatDuration(d.AverageTime))
			}
		},
	}

	return reportCmd
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
