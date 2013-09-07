package main

import (
	"fmt"
)

var cmdInfo = &Command{
	Run:   runInfo,
	Usage: "info",
	Short: "show app info",
	Long:  `Info shows general information about the current app.`,
}

func runInfo(cmd *Command, args []string) {
	var app App
	must(Get(&app, "/apps/"+mustApp()))
	fmt.Printf("Name:     %s\n", app.Name)
	fmt.Printf("Owner:    %s\n", app.Owner.Email)
	fmt.Printf("Stack:    %s\n", app.Stack)
	fmt.Printf("Git URL:  %s\n", app.GitURL)
	fmt.Printf("Web URL:  %s\n", app.WebURL)
}
