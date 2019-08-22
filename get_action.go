package main

import (
	"errors"
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/fatih/color"
	"github.com/ilikeorangutans/harbormaster/format"
	"github.com/urfave/cli"
	"os"
	"regexp"
	"strings"
	"text/tabwriter"
	"time"
)

func SetupGetActions() cli.Command {
	handlers := &GetActionsHandler{}
	return cli.Command{
		Name:    "get",
		Usage:   "Get things from Azkaban",
		Aliases: []string{"g"},
		Subcommands: []cli.Command{
			{
				Name:    "projects",
				Aliases: []string{"p"},
				Action:  handlers.GetProjectsAction,
			},
			{
				Name:      "flows",
				Usage:     "Lists flows and optionally filters them either by prefix or regex",
				Aliases:   []string{"f"},
				Action:    handlers.GetFlowsAction,
				ArgsUsage: "[prefix-filter|/regex/]",
			},
			{
				Name:      "executions",
				Aliases:   []string{"e"},
				Usage:     "Get executions",
				Action:    handlers.GetExecutionsAction,
				ArgsUsage: "flow",
			},
		},
	}
}

type GetActionsHandler struct {
	ActionWithContext
}

func (a *GetActionsHandler) GetExecutionsAction(c *cli.Context) error {
	if !c.Args().Present() {
		return errors.New("no flowID given")
	}

	client := a.Client()
	executions, err := client.FlowExecutions(c.GlobalString("project"), c.Args().First())
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
	return w.Flush()
}

func (a *GetActionsHandler) GetFlowsAction(c *cli.Context) error {
	client := a.Client()

	flows, err := client.ListFlows(c.GlobalString("project"))
	if err != nil {
		return err
	}

	var predicate func(string) bool
	if c.Args().Present() {
		regexString := c.Args().First()
		if strings.HasPrefix(regexString, "/") && strings.HasSuffix(regexString, "/") {
			regex, err := regexp.Compile(regexString[1 : len(c.Args().First())-1])
			if err != nil {
				return err
			}
			predicate = func(s string) bool { return regex.MatchString(s) }
		} else {
			predicate = func(s string) bool { return strings.HasPrefix(s, c.Args().First()) }
		}
	} else {
		predicate = func(string) bool { return true }
	}

	for _, f := range flows {
		if predicate(f.FlowID) {
			fmt.Printf("%s\n", f.FlowID)
		}
	}

	return nil
}

func (a *GetActionsHandler) GetProjectsAction(c *cli.Context) error {
	projectRepo := a.Context().Projects()
	projects, err := projectRepo.ListProjects()
	if err != nil {
		return err
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
	return w.Flush()
}
