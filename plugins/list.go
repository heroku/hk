package plugins

import "github.com/heroku/hk/cli"

var cmdList = &cli.Command{
	Signature: "plugins",
	ShortHelp: "list installed plugins",
	Help: `Lists installed plugins

  Example:
  $ heroku plugins`,

	Run: func(args []string, flags map[string]string) {
		for _, plugin := range Plugins() {
			cli.Stdoutln(plugin.Package.Name, plugin.Version)
		}
	},
}
