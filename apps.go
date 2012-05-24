package main

import (
	"fmt"
	"os"
)

func infoHelp() {
	cmdHelp("hk info -a <app>", "Show app info")
}

func info() {
	if (len(os.Args) != 4) || (os.Args[2] != "-a") {
		errorf("Invalid usage. See 'hk help info'")
	}
	appName := os.Args[3]
	var info struct {
		Name   string
		Owner  string `json:"owner_email"`
		Stack  string
		GitURL string `json:"git_url"`
		WebURL string `json:"web_url"`
	}
	apiReq(&info, "GET", fmt.Sprintf(apiURL+"/apps/%s", appName))
	fmt.Printf("Name:     %s\n", info.Name)
	fmt.Printf("Owner:    %s\n", info.Owner)
	fmt.Printf("Stack:    %s\n", info.Stack)
	fmt.Printf("Git URL:  %s\n", info.GitURL)
	fmt.Printf("Web URL:  %s\n", info.WebURL)
}

func listHelp() {
	cmdHelp("hk list", "List accessible apps")
}

func list() {
	if len(os.Args) != 2 {
		unrecArg(os.Args[2], "list")
	}
	var apps []struct{ Name string }
	apiReq(&apps, "GET", apiURL+"/apps")
	for _, app := range apps {
		fmt.Printf("%s\n", app.Name)
	}
}
