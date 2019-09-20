package cli

import (
	"github.com/spf13/cobra"
)

func NewCheckCmd(context Context) *cobra.Command {
	checkCmd := &cobra.Command{
		Use:     "check",
		Aliases: []string{"c"},
		Short:   "checks things",
	}

	checkCmd.AddCommand(newCheckFlowCmd(context))
	checkCmd.AddCommand(newCheckProjectCmd(context))

	return checkCmd
}
