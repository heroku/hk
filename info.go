package main

import (
	"fmt"
)

var cmdInfo = &Command{
	Run:      runInfo,
	Name:     "info",
	Usage:    "[-a <app>]",
	Category: "app",
	Short:    "show app info",
	Long:     `Info shows general information about the current app.`,
}

func init() {
	cmdInfo.Flag.StringVar(&flagApp, "a", "", "app name")
}

func runInfo(cmd *Command, args []string) {
	app, err := client.AppInfo(mustApp())
	must(err)
	fmt.Printf("Name:     %s\n", app.Name)
	fmt.Printf("Owner:    %s\n", app.Owner.Email)
	fmt.Printf("Region:   %s\n", app.Region.Name)
	fmt.Printf("Stack:    %s\n", app.Stack.Name)
	fmt.Printf("Git URL:  %s\n", app.GitURL)
	fmt.Printf("Web URL:  %s\n", app.WebURL)
}
