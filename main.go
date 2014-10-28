package main

import (
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"

	"github.com/bgentry/go-netrc/netrc"
	"github.com/heroku/hk/cli"
	"github.com/heroku/hk/plugins"
)

var topics = cli.NewTopicSet(
	plugins.Topic,
)

func main() {
	defer handlePanic()
	plugins.Setup()
	for _, topic := range plugins.PluginTopics() {
		topics.AddTopic(topic)
	}
	topic, command, args, flags := parse(os.Args[1:])
	if command == nil {
		help(os.Args[1:])
	}
	runCommand(topic, command, args, flags)
}

func handlePanic() {
	if e := recover(); e != nil {
		if e == "help" {
			help(os.Args[1:])
		}
		cli.Logf("ERROR: %s\n%s", e, debug.Stack())
		cli.Errln("ERROR:", e)
		cli.Exit(1)
	}
}

func runCommand(topic *cli.Topic, command *cli.Command, args []string, flags map[string]string) {
	ctx := &cli.Context{}
	if command.NeedsApp {
		app, err := app()
		if err != nil {
			cli.Errln(err)
			os.Exit(3)
		}
		ctx.App = app
	}
	if command.NeedsToken {
		ctx.Token = apiToken()
		if ctx.Token == "" {
			panic("error reading netrc")
		}
	}
	cli.Logf("Running %s:%s %s\n", topic, command, args, flags)
	before := time.Now()
	command.Run(ctx, args, flags)
	cli.Logf("Finished in %s\n", (time.Since(before)))
}

func parse(input []string) (topic *cli.Topic, command *cli.Command, args []string, flags map[string]string) {
	if len(input) == 0 {
		return
	}
	tc := strings.SplitN(input[0], ":", 2)
	topic = topics[tc[0]]
	if topic != nil {
		command = topic.GetCommand("")
		if len(tc) == 2 {
			command = topic.GetCommand(tc[1])
		}
	}
	args = input[1:]
	return topic, command, args, flags
}

func app() (string, error) {
	if app := os.Getenv("HEROKU_APP"); app != "" {
		return app, nil
	}
	return appFromGitRemote(remoteFromGitConfig())
}

func apiToken() string {
	netrc, err := netrc.ParseFile(filepath.Join(cli.HomeDir, ".netrc"))
	if err != nil {
		return ""
	}
	m := netrc.FindMachine("api.heroku.com")
	return m.Password
}
