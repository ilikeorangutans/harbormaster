package cli

import (
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/fatih/color"
	"github.com/ilikeorangutans/harbormaster/azkaban"
	"github.com/ilikeorangutans/harbormaster/format"
	"github.com/spf13/cobra"
	"log"
	"os"
	"text/tabwriter"
	"time"
)

func NewGetCmd(context Context) *cobra.Command {
	getCmd := &cobra.Command{
		Use:     "get",
		Aliases: []string{"g"},
		Short:   "get things from Azkaban",
	}

	getCmd.AddCommand(newGetProjectsCmd(context))
	getCmd.AddCommand(newGetFlowsCmd(context))
	getCmd.AddCommand(newGetExecutionsCmd(context))
	getCmd.AddCommand(newGetRunningCmd(context))

	return getCmd
}

func newGetProjectsCmd(context Context) *cobra.Command {
	return &cobra.Command{
		Use:     "projects",
		Aliases: []string{"p"},
		Short:   "get a list of projects",
		Run: func(cmd *cobra.Command, args []string) {
			projectRepo := context.Context().Projects()
			projects, err := projectRepo.ListProjects()
			if err != nil {
				log.Fatal(err)
			}

			w := new(tabwriter.Writer)
			w.Init(os.Stdout, 4, 4, 1, '\t', 0)
			fmt.Fprintln(w, "ID \t Name")

			for _, project := range projects {
				fmt.Fprintf(
					w,
					"%d \t %s \n",
					project.ID,
					project.Name,
				)
			}
			w.Flush()

		},
	}
}

func predicateFromArgs(args []string, position int) func(azkaban.Flow) bool {
	input := ""
	if len(args) > position {
		input = args[position]
	}
	return azkaban.MatchesFlowName(input)
}

func newGetFlowsCmd(context Context) *cobra.Command {
	return &cobra.Command{
		Use:     "flows",
		Aliases: []string{"f"},
		Short:   "Lists flows and optionally filters them either by prefix or regex",
		Args:    cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			client := context.Client()

			context.Context().Flows().ListFlows(azkaban.Project{Name: context.Project()}, azkaban.MatchesAll(predicateFromArgs(args, 0)))
			flows, err := client.ListFlows(context.Project())
			if err != nil {
				log.Fatal(err)
			}

			for _, f := range flows {
				fmt.Printf("%s\n", f.FlowID)
			}
		},
	}
}

func newGetExecutionsCmd(context Context) *cobra.Command {
	return &cobra.Command{
		Use:     "executions",
		Aliases: []string{"e"},
		Short:   "Get executions for a flow",
		Args:    cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			client := context.Client()
			executions, err := client.FlowExecutions(context.Project(), args[0], azkaban.TenMostRecent)
			if err != nil {
				panic(err)
			}

			w := new(tabwriter.Writer)
			w.Init(os.Stdout, 4, 4, 2, ' ', 0)
			fmt.Fprintf(
				w,
				"%s\t%s\t%s\t%s\t%s\n",
				"ExecID",
				color.WhiteString("Status"),
				"Runtime",
				"When",
				"Start Time",
			)

			for _, e := range executions {
				fmt.Fprintf(
					w,
					"%d \t%s \t%s \t%s \t%s\n",
					e.ID,
					e.Status.Colored(),
					format.DurationHumanReadable(e.Duration()),
					humanize.Time(e.StartTime.Time()),
					e.StartTime.Time().Format(time.RFC1123),
				)
			}
			w.Flush()
		},
	}
}

func newGetRunningCmd(context Context) *cobra.Command {
	return &cobra.Command{
		Use:     "running",
		Aliases: []string{"r"},
		Short:   "Get currently running flows",
		Run: func(cmd *cobra.Command, args []string) {
			executions, err := context.Client().Running()
			if err != nil {
				log.Fatal(err)
			}

			w := new(tabwriter.Writer)
			w.Init(os.Stdout, 4, 4, 2, ' ', 0)
			fmt.Fprintf(
				w,
				"%s\t%s\t%s\t%s\t%s\n",
				"Project",
				"FlowID",
				color.WhiteString("Status"),
				"Start time",
				"Runtime",
			)

			for _, execution := range executions {
				fmt.Fprintf(
					w,
					"%s\t%s\t%s\t%s\t%s\n",
					execution.Project,
					execution.FlowID,
					execution.Status.Colored(),
					humanize.Time(execution.StartTime.Time()),
					format.DurationHumanReadable(time.Since(execution.StartTime.Time())),
				)
			}

			w.Flush()
		},
	}
}
