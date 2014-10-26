package plugins

import "github.com/heroku/hk/cli"

func install(name string) {
	cli.Stderrf("Installing plugin %s...\n", name)
	must(node.InstallPackage(name))
}
