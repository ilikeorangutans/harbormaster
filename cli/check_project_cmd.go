package cli

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/ilikeorangutans/harbormaster/azkaban"
	"github.com/spf13/cobra"
	"log"
	"os"
	"strings"
	"text/tabwriter"
)

func newCheckProjectCmd(context Context) *cobra.Command {
	return &cobra.Command{
		Use:     "project",
		Aliases: []string{"p"},
		Args:    cobra.RangeArgs(1, 2),
		Short:   "check flows of a project",
		Run: func(cmd *cobra.Command, args []string) {
			flowNamePredicate := predicateFromArgs(args, 1)
			project, err := context.Context().Projects().ByName(args[0])
			if err != nil {
				log.Fatal(err)
			}

			allFlows, err := context.Context().Flows().ListFlows(project)
			if err != nil {
				log.Fatal(err)
			}
			var flows []azkaban.Flow
			for _, flow := range allFlows {
				if flowNamePredicate(flow.FlowID) {
					flows = append(flows, flow)
				}
			}

			w := new(tabwriter.Writer)
			w.Init(os.Stdout, 2, 8, 1, ' ', 0)
			columns := []interface{}{
				"Flow",
				color.WhiteString("Health"),
				color.WhiteString("Histogram (← most recent)"),
			}
			headerFormat := strings.Join([]string{"%s", "%-20s", "%s\n"}, "\t")
			rowFormat := strings.Join([]string{"%s", "%-20s", "%s\n"}, "\t")
			fmt.Fprintf(w, headerFormat, columns...)

			for _, f := range flows {
				executions, err := context.Context().Executions().ListExecutions(project, f, azkaban.TenMostRecent)
				if err != nil {
					log.Fatal(err)
				}

				fmt.Fprintf(
					w,
					rowFormat,
					f.FlowID,
					executions.Health().Colored(),
					executions.Histogram().Histogram,
				)
				if executions.Histogram().Failures > 0 {
					for _, line := range executions.HistogramDetails(3) {
						fmt.Fprintf(w, rowFormat, "", color.WhiteString(""), line)
					}
				}
			}
			w.Flush()
		},
	}
}
