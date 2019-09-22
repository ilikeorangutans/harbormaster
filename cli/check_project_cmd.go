package cli

import (
	"fmt"
	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/ilikeorangutans/harbormaster/azkaban"
	"github.com/spf13/cobra"
	"log"
	"os"
	"strings"
	"text/tabwriter"
	"time"
)

func newCheckProjectCmd(context Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "project",
		Aliases: []string{"p"},
		Args:    cobra.RangeArgs(1, 2),
		Short:   "check flows of a project",

		Run: func(cmd *cobra.Command, args []string) {

			s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
			spinnerStatus := newSpinnerStatus()
			s.Start()
			s.Suffix = spinnerStatus.String()

			project, err := context.Context().Projects().ByName(args[0])
			if err != nil {
				log.Fatal(err)
			}

			flowNamePredicate := predicateFromArgs(args, 1)

			flows, err := context.Context().Flows().ListFlows(project, azkaban.MatchesAll(flowNamePredicate))
			if err != nil {
				log.Fatal(err)
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

			numberOfExecutions, _ := cmd.Flags().GetUint("count")
			numberOfDetails, _ := cmd.Flags().GetUint("details")
			ignoreHealthy, _ := cmd.Flags().GetBool("ignore-healthy")
			if numberOfDetails > numberOfExecutions {
				numberOfDetails = numberOfExecutions
			}
			spinnerStatus.SetTotal(len(flows))
			for i, f := range flows {
				spinnerStatus.SetProgress(i)
				s.Suffix = spinnerStatus.String()
				executions, err := context.Context().Executions().ListExecutions(project, f, azkaban.NMostRecent(int(numberOfExecutions)))
				if executions.Health().IsHealthy() && ignoreHealthy {
					spinnerStatus.AddIgnored()
					continue
				}
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
				if executions.Health() != azkaban.Healthy {
					for _, line := range executions.HistogramDetails(int(numberOfDetails)) {
						fmt.Fprintf(w, rowFormat, "", color.WhiteString(""), line)
					}
				}
			}
			s.Stop()
			w.Flush()
		},
	}

	cmd.Flags().UintP("count", "c", 5, "how many executions to check")
	cmd.Flags().UintP("details", "d", 3, "for how many executions to show details")
	cmd.Flags().BoolP("ignore-healthy", "i", false, "show only non-healthy flows")

	return cmd
}

type spinnerStatus struct {
	total    int
	ignored  int
	progress int
}

func newSpinnerStatus() *spinnerStatus {
	return &spinnerStatus{
		total:    0,
		ignored:  0,
		progress: 0,
	}
}
func (s *spinnerStatus) SetTotal(n int) {
	s.total = n
}

func (s *spinnerStatus) SetProgress(n int) {
	s.progress = n
}
func (s *spinnerStatus) AddIgnored() {
	s.ignored += 1
}

func (s *spinnerStatus) String() string {
	if s.total == 0 {
		return " fetching flows..."
	}
	if s.ignored > 0 {
		return fmt.Sprintf(" fetching execution %d/%d (%d ignored)", s.progress, s.total, s.ignored)
	} else {
		return fmt.Sprintf(" fetching execution %d/%d", s.progress, s.total)
	}
}
