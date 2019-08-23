package cli

import (
	"fmt"
	"github.com/ilikeorangutans/harbormaster/azkaban"
	"github.com/spf13/cobra"
)

func NewLoginCmd(context Context) *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "create a new session",
		Long: `Creates a new session using the given URL, username, and password
and outputs the Azkaban session ID and Host to be used to set these values as
environment variables. Ideally you run this command in a subshell, like so:

# $(harbormaster login <url> <username> <password>)

This will automatically set these values for your current shell and subsequent
invocations of harbormaster will use these values.`,
		Args: cobra.ExactArgs(3),
		Run: func(cmd *cobra.Command, args []string) {
			client, err := azkaban.ConnectWithUsernameAndPassword(args[0], args[1], args[2])
			if err != nil {
				fmt.Println(err.Error())
				return
			}

			fmt.Printf("export %s=%s\n", HarbormasterSessionID, client.SessionID)
			fmt.Printf("export %s=%s\n", HarbormasterHost, args[0])
		},
	}
}
