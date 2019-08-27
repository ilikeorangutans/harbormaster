package cli

import (
	"fmt"
	"github.com/ilikeorangutans/harbormaster/azkaban"
	"github.com/ilikeorangutans/harbormaster/format"
	"github.com/spf13/cobra"
	"log"
	"os"
	"strings"
	"text/tabwriter"
	"time"
)

func NewReportCmd(context Context) *cobra.Command {
	averageExecutionTimeCmd := &cobra.Command{
		Use:     "average-execution-time",
		Aliases: []string{"aet"},
		Short:   "average execution time ",
		Run: func(cmd *cobra.Command, args []string) {
			project := azkaban.Project{Name: context.Project()}
			flows, err := context.Context().Flows().ListFlows(project)
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
				executions, err := execRepo.ListExecutions(project, f, azkaban.TenMostRecent)
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

			formatter := consoleFormatter
			format, err := cmd.Flags().GetString("format")
			if err != nil {
				log.Fatal(err)
			}
			if format == "tsv" {
				formatter = tsvFormatter
			}

			formatter(data)
		},
	}
	averageExecutionTimeCmd.Flags().StringP("format", "f", "console", "format to display data in, valid are [console, tsv]")

	reportCmd := &cobra.Command{
		Use: "report",
	}
	reportCmd.AddCommand(averageExecutionTimeCmd)

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

func consoleFormatter(data []execReportData) {
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 4, 4, 2, ' ', 0)
	fmt.Fprintf(
		w,
		"%s\t%s\t%s\t%s\n",
		"FlowID",
		"Success Count",
		"Failure Count",
		"Average Time",
	)
	for _, d := range data {
		fmt.Fprintf(
			w,
			"%s\t%d\t%d\t%s\n",
			d.FlowID,
			d.SuccessCount,
			d.TotalCount-d.SuccessCount,
			format.DurationHumanReadable(d.AverageTime),
		)
	}
	w.Flush()

}

func tsvFormatter(data []execReportData) {
	fmt.Printf("FlowID \tSuccess Count \t Average Time\n")
	for _, d := range data {
		fmt.Printf("%s \t %4d \t%s\n", d.FlowID, d.SuccessCount, formatDuration(d.AverageTime))
	}
}
