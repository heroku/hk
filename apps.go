package main

import (
	"fmt"
	"os"
	"os/exec"
)

var cmdInfo = &Command{
	Run:   runInfo,
	Usage: "info [-a app]",
	Short: "show app info",
	Long:  `Info shows general information about the current app.`,
}

func init() {
	cmdInfo.Flag.StringVar(&flagApp, "a", "", "app")
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

var cmdURL = &Command{
	Run:   runURL,
	Usage: "url [-a app]",
	Short: "show app url" + extra,
	Long:  `Prints the web URL for the app.`,
}

func init() {
	cmdURL.Flag.StringVar(&flagApp, "a", "", "app")
}

func runURL(cmd *Command, args []string) {
	fmt.Println("https://" + mustApp() + ".herokuapp.com/")
}

var cmdOpen = &Command{
	Run:   runOpen,
	Usage: "open",
	Short: "open app in a web browser",
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

var cmdCreate = &Command{
	Run:   runCreate,
	Usage: "create [app]",
	Short: "create an app",
	Long:  `Create creates a new heroku app.`,
}

func runCreate(cmd *Command, args []string) {
	var app App
	var v struct {
		Name string `json:"name,omitempty"`
	}
	if len(args) > 0 {
		v.Name = args[0]
	}
	must(Post(&app, "/apps", v))
	exec.Command("git", "remote", "add", "heroku", app.GitURL).Run()
	fmt.Println(app.Name)
}

var cmdRename = &Command{
	Run:   runRename,
	Usage: "rename old new",
	Short: "rename an app",
	Long: `
Rename renames a heroku app.

Example:

  $ hk rename myapp myapp2
`,
}

func runRename(cmd *Command, args []string) {
	if len(args) != 2 {
		cmd.printUsage()
		os.Exit(2)
	}
	oldname, newname := args[0], args[1]
	var app App
	p := map[string]string{"name": newname}
	must(Put(&app, "/apps/"+oldname, p))
	fmt.Println("Renamed app to: " + app.Name)
	fmt.Println("Ensure you update your git remote URL.")
	// should we automatically update the remote if they specify an app
	// or via mustApp + conditional logic - RM
}
