package plugins

import "github.com/heroku/hk/cli"

var cmdInstall = &cli.Command{
	Name:      "install",
	Signature: "plugins:install [plugin]",
	ShortHelp: "ShortHlp",
	Help: `Install a Heroku plugin

  Example:
  $ heroku plugins:install dickeyxxx/heroku-production-status`,

	Run: func(args []string, flags map[string]string) {
		if len(args) == 0 {
			panic("help")
		}
		name := args[0]
		cli.Stderrf("Installing plugin %s...\n", name)
		must(node.InstallPackage(name))
	},
}
