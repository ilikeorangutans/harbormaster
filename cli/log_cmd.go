package cli

import (
	"github.com/spf13/cobra"
	"log"
	"net/url"
	"os"
	"strconv"
	"time"
)

func NewLogCmd(context Context) *cobra.Command {
	logCmd := &cobra.Command{
		Use:     "log",
		Aliases: []string{"l"},
		Short:   "fetches logs",
		Run: func(cmd *cobra.Command, args []string) {
			follow, _ := cmd.Flags().GetBool("follow")
			client := context.Client()

			var projectID string
			var unparsedExecID string

			if len(args) == 1 {
				u, err := url.Parse(args[0])
				if err != nil {
					log.Fatal(err)
				}
				query := u.Query()
				unparsedExecID = query.Get("execid")
				projectID = query.Get("job")
			} else if len(args) == 2 {
				projectID = args[0]
				unparsedExecID = args[1]
			}

			execID, err := strconv.ParseInt(unparsedExecID, 10, 64)
			if err != nil {
				log.Fatal(err)
			}

			lastOffset, err := client.FetchLogsUntilEnd(execID, projectID, 0, os.Stdout)
			if err != nil {
				log.Fatal(err)
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
		},
	}

	logCmd.Flags().BoolP("follow", "f", false, "follow log, indefinitely updates every 2 seconds")

	return logCmd
}
