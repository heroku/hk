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
	var info struct {
		Name   string
		Owner  string `json:"owner_email"`
		Stack  string
		GitURL string `json:"git_url"`
		WebURL string `json:"web_url"`
	}
	apiReq(&info, "GET", fmt.Sprintf(apiURL+"/apps/%s", *flagApp))
	fmt.Printf("Name:     %s\n", info.Name)
	fmt.Printf("Owner:    %s\n", info.Owner)
	fmt.Printf("Stack:    %s\n", info.Stack)
	fmt.Printf("Git URL:  %s\n", info.GitURL)
	fmt.Printf("Web URL:  %s\n", info.WebURL)
}

var cmdList = &Command{
	Run:   runList,
	Usage: "list",
	Short: "list apps",
	Long:  `List lists all accessible apps.`,
}

func runList(cmd *Command, args []string) {
	var apps []struct{ Name string }
	apiReq(&apps, "GET", apiURL+"/apps")
	for _, app := range apps {
		fmt.Printf("%s\n", app.Name)
	}
}
