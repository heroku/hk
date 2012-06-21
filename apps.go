package main

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
)

var cmdInfo = &Command{
	Run:   runInfo,
	Usage: "info [-a APP]",
	Short: "show app info",
	Long:  `Info shows general information about the current app.`,
}

func init() {
	cmdInfo.Flag.StringVar(&flagApp, "a", "", "app")
}

func runInfo(cmd *Command, args []string) {
	var info struct {
		Name   string
		Owner  string `json:"owner_email"`
		Stack  string
		GitURL string `json:"git_url"`
		WebURL string `json:"web_url"`
	}
	APIReq("GET", "/apps/"+mustApp()).Do(&info)
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
	u := "https://" + mustApp() + ".herokuapp.com/"
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
	APIReq("GET", "/apps").Do(&apps)
	for _, app := range apps {
		fmt.Printf("%s\n", app.Name)
	}
}

var cmdCreate = &Command{
	Run:   runCreate,
	Usage: "create [APP]",
	Short: "create an app",
	Long:  `Create creates a new heroku app.`,
}

func runCreate(cmd *Command, args []string) {
	var info struct {
		Name   string
		Stack  string
		GitURL string `json:"git_url"`
	}

	v := make(url.Values)
	v.Set("app[stack]", "cedar")
	if len(args) > 0 {
		v.Set("app[name]", args[0])
	}

	r := APIReq("POST", "/apps")
	r.SetBodyForm(v)
	r.Do(&info)
	exec.Command("git", "remote", "add", "heroku", info.GitURL).Run()
	fmt.Println(info.Name)
}

var cmdDestroy = &Command{
	Run:   runDestroy,
	Usage: "destroy <APP>",
	Short: "destroy an app",
	Long:  `Destroy destroys a heroku app.`,
}

func runDestroy(cmd *Command, args []string) {
	if len(args) != 1 {
		cmd.printUsage()
		os.Exit(2)
	}

	APIReq("DELETE", "/apps/"+args[0]).Do(nil)
	for _, remote := range gitRemotes(gitURL(args[0])) {
		exec.Command("git", "remote", "rm", remote).Run()
	}
}
