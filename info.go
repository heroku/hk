package main

import (
	"fmt"
)

var cmdAppInfo = &Command{
	Run:   runAppInfo,
	Usage: "info",
	Short: "show app info",
	Long:  `Info shows general information about the current app.`,
}

func runAppInfo(cmd *Command, args []string) {
	app, err := client.AppInfo(mustApp())
	must(err)
	fmt.Printf("Name:     %s\n", app.Name)
	fmt.Printf("Owner:    %s\n", app.Owner.Email)
	fmt.Printf("Region:   %s\n", app.Region.Name)
	fmt.Printf("Stack:    %s\n", app.Stack)
	fmt.Printf("Git URL:  %s\n", app.GitURL)
	fmt.Printf("Web URL:  %s\n", app.WebURL)
}
