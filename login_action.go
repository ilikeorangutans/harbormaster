package main

import (
	"errors"
	"fmt"
	"github.com/ilikeorangutans/harbormaster/azkaban"
	"github.com/urfave/cli"
)

func SetupLoginAction() cli.Command {
	return cli.Command{
		Name:  "login",
		Usage: "create a new session",
		UsageText: `Creates a new session using the given URL, username, and password
and outputs the Azkaban session ID and host to be used to set these values as
environment variables. Ideally you run this command in a subshell, like so:

# $(harbormaster login <url> <username> <password>)

This will automatically set these values for your current shell and subsequent
invocations of harbormaster will use these values.`,
		Action:    LoginAction,
		ArgsUsage: "<URL> <username> <password>",
	}
}

func LoginAction(c *cli.Context) error {
	if !c.Args().Present() {
		return errors.New("expected <URL> <username> <password>")
	}
	client, err := azkaban.ConnectWithUsernameAndPassword(c.Args().Get(0), c.Args().Get(1), c.Args().Get(2))
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	fmt.Printf("export %s=%s\n", AzkabanSessionIDEnv, client.SessionID)
	fmt.Printf("export %s=%s\n", AzkabanHostEnv, c.Args().Get(0))
	return nil
}
