package plugins

import (
	"github.com/dickeyxxx/gode"
	"github.com/heroku/hk/cli"
)

var node = gode.NewClient(cli.AppDir)

var Topic = &cli.Topic{
	Name:      "plugins",
	ShortHelp: "manage plugins",
	Help: `Manage the Heroku CLI Plugins
	
  Example:
  $ heroku plugins:install dickeyxxx/heroku-production-check`,

	Commands: []*cli.Command{
		cmdList,
		cmdInstall,
	},
}
