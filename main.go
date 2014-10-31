package main

import (
	"os"
	"path/filepath"
	"runtime/debug"
	"time"

	"github.com/bgentry/go-netrc/netrc"
	"github.com/heroku/hk/apps"
	"github.com/heroku/hk/cli"
	"github.com/heroku/hk/plugins"
)

var Cli = cli.NewCli(
	apps.Apps,
	apps.Info,
	plugins.Plugins,
)

func main() {
	defer handlePanic()
	plugins.Setup()
	for _, topic := range plugins.PluginTopics() {
		Cli.AddTopic(topic)
	}
	ctx, err := Cli.Parse(os.Args[1:])
	if err != nil {
		if err == cli.HelpErr {
			help()
		}
		cli.Errln(err)
		cli.Errf("USAGE: %s %s\n", os.Args[0], commandSignature(ctx.Topic, ctx.Command))
		os.Exit(2)
	}
	if ctx.Command == nil {
		help()
	}
	if ctx.Command.NeedsApp {
		if ctx.App == "" {
			ctx.App = app()
		}
		if ctx.App == "" {
			cli.Errln(" !    No app specified.")
			cli.Errln(" !    Run this command from an app folder or specify which app to use with --app APP.")
			os.Exit(3)
		}
	}
	if ctx.Command.NeedsAuth {
		ctx.Auth.Username, ctx.Auth.Password = auth()
	}
	cli.Logf("Running %s\n", ctx)
	before := time.Now()
	ctx.Command.Run(ctx)
	cli.Logf("Finished in %s\n", (time.Since(before)))
}

func handlePanic() {
	if e := recover(); e != nil {
		cli.Logf("ERROR: %s\n%s", e, debug.Stack())
		cli.Errln("ERROR:", e)
		cli.Exit(1)
	}
}

func app() string {
	if app := os.Getenv("HEROKU_APP"); app != "" {
		return app
	}
	app, err := appFromGitRemote(remoteFromGitConfig())
	if err != nil {
		panic(err)
	}
	return app
}

func auth() (user, password string) {
	netrc, err := netrc.ParseFile(filepath.Join(cli.HomeDir, ".netrc"))
	if err != nil {
		panic(err)
	}
	auth := netrc.FindMachine("api.heroku.com")
	return auth.Login, auth.Password
}
