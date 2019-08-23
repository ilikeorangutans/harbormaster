package main

import (
	"fmt"
	"github.com/ilikeorangutans/harbormaster/cli"
	"os"
)

func main() {
	if err := cli.NewCLI().Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
