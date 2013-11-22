package main

import (
	"fmt"
	"github.com/bgentry/heroku-go"
	"os/exec"
)

var cmdAppCreate = &Command{
	Run:   runAppCreate,
	Usage: "create [-r region] [app]",
	Short: "create an app",
	Long:  `Create creates a new heroku app.`,
}

var flagRegion string

func init() {
	cmdAppCreate.Flag.StringVar(&flagRegion, "r", "", "region name")
}

func runAppCreate(cmd *Command, args []string) {
	var opts heroku.AppCreateOpts
	if flagRegion != "" {
		opts.Region = &flagRegion
	}
	if len(args) > 0 {
		opts.Name = &args[0]
	}
	app, err := client.AppCreate(opts)
	must(err)
	exec.Command("git", "remote", "add", "heroku", app.GitURL).Run()
	fmt.Println(app.Name)
}
