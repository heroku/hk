package plugins

import "github.com/heroku/hk/cli"

var cmdInstall = &cli.Command{
	Name:      "install",
	ShortHelp: "Installs a plugin into the CLI",
	Help: `Install a Heroku plugin

  Example:
  $ heroku plugins:install dickeyxxx/heroku-production-status`,

	Run: func(ctx *cli.Context) {
		if len(ctx.Args) == 0 {
			panic("help")
		}
		name := ctx.Args[0]
		cli.Errf("Installing plugin %s... ", name)
		must(node.InstallPackage(name))
		cli.Errln("done")
	},
}
