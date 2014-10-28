package plugins

import "github.com/heroku/hk/cli"

var cmdList = &cli.Command{
	Signature: "plugins",
	ShortHelp: "Lists the installed plugins",
	Help: `Lists installed plugins

  Example:
  $ heroku plugins`,

	Run: func(ctx *cli.Context) {
		packages, err := node.Packages()
		must(err)
		for _, pkg := range packages {
			cli.Println(pkg.Name, pkg.Version)
		}
	},
}
