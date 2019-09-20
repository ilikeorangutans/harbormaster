package cli

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/ilikeorangutans/harbormaster/azkaban"
	"github.com/spf13/cobra"
	"log"
	"os"
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
			maxFlowNameWidth := 40
			var flows []azkaban.Flow
			for _, flow := range allFlows {
				if flowNamePredicate(flow.FlowID) {
					flows = append(flows, flow)
					if len(flow.FlowID) > maxFlowNameWidth {
						maxFlowNameWidth = len(flow.FlowID)
					}
				}
			}

			formatString := fmt.Sprintf("%%-%ds \t %%s \t %%s \n", maxFlowNameWidth)

			w := new(tabwriter.Writer)
			w.Init(os.Stdout, 2, 8, 1, '\t', 0)
			fmt.Fprintf(w, fmt.Sprintf("%%-%ds \t %%s \t %%s\n", maxFlowNameWidth), "Flow", color.WhiteString("Health"), color.WhiteString("Histogram (← most recent)"))

			for _, f := range flows {
				executions, err := context.Context().Executions().ListExecutions(project, f, azkaban.TenMostRecent)
				if err != nil {
					log.Fatal(err)
				}

				fmt.Fprintf(
					w,
					formatString,
					f.FlowID,
					executions.Health().Colored(),
					executions.Histogram().Histogram,
				)

				w.Flush()
			}

		},
	}
}
