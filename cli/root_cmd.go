package cli

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	HarbormasterSessionID = "HARBORMASTER_SESSION_ID"
	HarbormasterHost      = "HARBORMASTER_HOST"
	HarbormasterProject   = "HARBORMASTER_PROJECT"
)

func NewRootCmd() *cobra.Command {
	var rootCmd = &cobra.Command{
		Use:     "harbormaster",
		Short:   "Azkaban command line power tools",
		Version: "0.1.3",
	}

	rootCmd.PersistentFlags().StringP("project", "p", "", "Azkaban project to use")
	viper.BindEnv("project", HarbormasterProject)
	viper.BindPFlag("project", rootCmd.PersistentFlags().Lookup("project"))

	rootCmd.PersistentFlags().String("session-id", "", "Azkaban session ID")
	viper.BindPFlag("session-id", rootCmd.PersistentFlags().Lookup("session-id"))

	viper.BindEnv("session-id", HarbormasterSessionID)

	rootCmd.PersistentFlags().String("host", "", "Azkaban host")
	viper.BindPFlag("host", rootCmd.PersistentFlags().Lookup("host"))
	viper.BindEnv("host", HarbormasterHost)

	viper.SetDefault("dump-responses", false)
	rootCmd.PersistentFlags().Bool("dump-responses", false, "Dump HTTP responses from Azkaban")

	return rootCmd
}
