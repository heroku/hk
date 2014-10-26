package plugins

import (
	"github.com/dickeyxxx/gode"
	"github.com/heroku/hk/cli"
)

var node = gode.NewClient(cli.AppDir)

var Topic = &cli.Topic{
	Name:      "plugins",
	ShortHelp: "manage plugins",
	Run:       Run,
	Help:      Help,
}

func Run(command string, args ...string) {
	switch command {
	case "install":
		if len(args) != 1 {
			Help(command, args...)
			cli.Exit(2)
		}
		install(args[0])
	case "list":
	case "":
		list()
	default:
		Help(command, args...)
	}
}

func Help(command string, args ...string) {
	cli.Stderrln("TODO: help for " + command)
}
