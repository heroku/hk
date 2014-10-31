package plugins

import "github.com/heroku/hk/cli"

var cmdInstall = &cli.Command{
	Name:      "install",
	Args:      []*cli.Arg{{Name: "name"}},
	ShortHelp: "Installs a plugin into the CLI",
	Help: `Install a Heroku plugin

  Example:
  $ heroku plugins:install dickeyxxx/heroku-production-status`,

	Run: func(ctx *cli.Context) {
		name := ctx.Args["name"]
		cli.Errf("Installing plugin %s... ", name)
		must(node.InstallPackage(name))
		cli.Errln("done")
	},
}
