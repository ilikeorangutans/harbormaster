package main

import (
	"errors"
	"github.com/urfave/cli"
	"net/url"
	"os"
	"strconv"
	"time"
)

func SetupLogAction() cli.Command {
	logAction := &LogActionHandler{}

	return cli.Command{
		Action:    logAction.LogAction,
		Aliases:   []string{"l"},
		ArgsUsage: "(URL|job_id exec_id)",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "follow,f",
				Usage: "follow log, indefinitely updates every 2 seconds",
			},
		},
		Name:  "logs",
		Usage: "Fetches logs for an execution, either via URL or by job and exec ID",
	}
}

type LogActionHandler struct {
	ActionWithContext
}

func (a *LogActionHandler) LogAction(c *cli.Context) error {
	follow := c.Bool("f")
	client := a.Client()

	var projectID string
	var unparsedExecID string

	if len(c.Args()) == 1 {
		u, err := url.Parse(c.Args().Get(0))
		if err != nil {
			return err
		}
		query := u.Query()
		unparsedExecID = query.Get("execid")
		projectID = query.Get("job")
	} else if len(c.Args()) == 2 {
		projectID = c.Args().Get(0)
		unparsedExecID = c.Args().Get(1)
	} else {
		return errors.New("not enough or too many arguments")
	}

	execID, err := strconv.ParseInt(unparsedExecID, 10, 64)
	if err != nil {
		return err
	}

	lastOffset, err := client.FetchLogsUntilEnd(execID, projectID, 0, os.Stdout)
	if err != nil {
		return err
	}

	offset := lastOffset
	if follow {
		ticker := time.NewTicker(time.Second * 2)

		for {
			select {
			case <-ticker.C:
				offset, err = client.FetchLogsUntilEnd(execID, projectID, offset, os.Stdout)
				if err != nil {
					panic(err)
				}
			}
		}
	}

	return nil
}
