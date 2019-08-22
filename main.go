package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli"
)

const (
	AzkabanSessionIDEnv = "AZKABAN_SESSION_ID"
	AzkabanHostEnv      = "AZKABAN_HOST"
)

func main() {
	app := cli.NewApp()
	app.Name = "harbormaster"
	app.Version = "0.1.3"
	app.Author = "Jakob Külzer"
	app.Email = "jakob.kuelzer@gmail.com"
	app.EnableBashCompletion = true
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "dump-responses",
			Usage: "Dump all HTTP responses from Azkaban to stdout",
		},
		cli.StringFlag{
			Name:   "project",
			Usage:  "the project you want to interact with",
			EnvVar: "HARBORMASTER_PROJECT",
		},
	}
	app.Commands = []cli.Command{
		SetupLoginAction(),
		SetupGetActions(),
		SetupLogAction(),
		SetupCheckFlowAction(),
		SetupReportActions(),
	}
	if err := app.Run(os.Args); err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}
}
