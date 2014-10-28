package plugins

import "github.com/heroku/hk/cli"

func Setup() {
	if node.IsSetup() {
		return
	}
	cli.Err("Setting up plugins... ")
	must(node.Setup())
	cli.Errln("done")
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
