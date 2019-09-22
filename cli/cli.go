package cli

import (
	"github.com/ilikeorangutans/harbormaster/azkaban"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
	"net/url"
	"os"
)

// NewCLI builds the full cobra root command with all sub commands attached
func NewCLI() *cobra.Command {
	context := Context{
		DumpResponses: viper.GetBool("dump-responses"),
		client:        nil,
		context:       nil,
	}

	rootCmd := NewRootCmd()
	rootCmd.AddCommand(NewLoginCmd(context))
	rootCmd.AddCommand(NewGetCmd(context))
	rootCmd.AddCommand(NewLogCmd(context))
	rootCmd.AddCommand(NewCheckCmd(context))
	rootCmd.AddCommand(NewReportCmd(context))

	completionCommand := &cobra.Command{
		Use:   "completion",
		Short: "generate autocomplete scripts",
	}

	completionCommand.AddCommand(&cobra.Command{
		Use:   "bash",
		Short: "Generates bash completion scripts",
		Long: `To load completion run

. <(harbormaster completion)

To configure your bash shell to load completions for each session add to your bashrc

# ~/.bashrc or ~/.profile
. <(harbormaster completion zsh)
`,
		Run: func(cmd *cobra.Command, args []string) {
			rootCmd.GenBashCompletion(os.Stdout)
		},
	})

	completionCommand.AddCommand(&cobra.Command{
		Use:   "zsh",
		Short: "Generates zsh completion scripts",
		Long:  "Place the output in a file called _harbormaster somewhere in your $fpath, usually ~/.zsh/completions/_harbormaster",
		Run: func(cmd *cobra.Command, args []string) {
			rootCmd.GenZshCompletion(os.Stdout)
		},
	})

	rootCmd.AddCommand(completionCommand)
	return rootCmd
}

type Context struct {
	DumpResponses bool
	client        *azkaban.Client
	context       *azkaban.Context
}

func (c *Context) SessionID() string {
	return viper.GetString("session-id")
}
func (c *Context) Project() string {
	return viper.GetString("project")
}

func (c *Context) Host() string {
	azkabanURL, err := url.Parse(viper.GetString("host"))
	if err != nil {
		log.Fatal(err)
	}

	return azkabanURL.String()
}

func (c *Context) Context() *azkaban.Context {
	if c.context != nil {
		return c.context
	}
	c.context = azkaban.NewContext(c.Client())
	return c.context
}

func (c *Context) Client() *azkaban.Client {
	if c.client != nil {
		return c.client
	}

	var err error
	c.client, err = azkaban.ConnectWithSessionID(c.Host(), c.SessionID())
	if err != nil {
		panic(err)
	}

	c.client.DumpResponses = c.DumpResponses

	return c.client
}
