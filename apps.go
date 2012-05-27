package main

import (
	"fmt"
	"net/url"
	"os/exec"
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
	apiReq(&info, "GET", fmt.Sprintf(apiURL+"/apps/%s", app()), nil)
	fmt.Printf("Name:     %s\n", info.Name)
	fmt.Printf("Owner:    %s\n", info.Owner)
	fmt.Printf("Stack:    %s\n", info.Stack)
	fmt.Printf("Git URL:  %s\n", info.GitURL)
	fmt.Printf("Web URL:  %s\n", info.WebURL)
}

var cmdOpen = &Command{
	Run:   runOpen,
	Usage: "open",
	Short: "open app",
	Long:  `Open opens the app in a web browser. (Assumes cedar.)`,
}

func runOpen(cmd *Command, args []string) {
	u := "https://" + app() + ".herokuapp.com/"
	command := "open"
	if _, err := exec.LookPath("xdg-open"); err == nil {
		command = "xdg-open"
	}
	exec.Command(command, u).Start()
}

var cmdList = &Command{
	Run:   runList,
	Usage: "list",
	Short: "list apps",
	Long:  `List lists all accessible apps.`,
}

func runList(cmd *Command, args []string) {
	var apps []struct{ Name string }
	apiReq(&apps, "GET", apiURL+"/apps", nil)
	for _, app := range apps {
		fmt.Printf("%s\n", app.Name)
	}
}

var cmdCreate = &Command{
	Run:   runCreate,
	Usage: "create [name]",
	Short: "create an app",
	Long:  `Create creates a new heroku app.`,
}

func runCreate(cmd *Command, args []string) {
	var info struct {
		Name  string
		Stack string
	}

	v := make(url.Values)
	if len(args) > 0 {
		v.Set("app[name]", args[0])
	}
	if len(args) > 1 {
		v.Set("app[stack]", args[1])
	}

	apiReq(&info, "POST", apiURL+"/apps", v)
	fmt.Println(info.Name)
}
