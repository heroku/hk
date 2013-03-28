package main

import (
	"fmt"
	"os/exec"
)

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
