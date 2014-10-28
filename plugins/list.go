package plugins

import "github.com/heroku/hk/cli"

var cmdList = &cli.Command{
	Signature: "plugins",
	ShortHelp: "list installed plugins",
	Help: `Lists installed plugins

  Example:
  $ heroku plugins`,

	Run: func(args []string, flags map[string]string) {
		packages, err := node.Packages()
		must(err)
		for _, pkg := range packages {
			cli.Stdoutln(pkg.Name, pkg.Version)
		}
	},
}
