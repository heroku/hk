package main

import (
	"fmt"
	"os"
)

var cmdInfo = &Command{
	Run:      runInfo,
	Usage:    "info",
	NeedsApp: true,
	Category: "app",
	Short:    "show app info",
	Long:     `Info shows general information about the current app.`,
}

func runInfo(cmd *Command, args []string) {
	if len(args) != 0 {
		cmd.PrintUsage()
		os.Exit(2)
	}
	app, err := client.AppInfo(mustApp())
	must(err)
	fmt.Printf("Name:     %s\n", app.Name)
	fmt.Printf("Owner:    %s\n", app.Owner.Email)
	fmt.Printf("Region:   %s\n", app.Region.Name)
	fmt.Printf("Stack:    %s\n", app.Stack.Name)
	fmt.Printf("Git URL:  %s\n", app.GitURL)
	fmt.Printf("Web URL:  %s\n", app.WebURL)
}
