package plugins

import "github.com/heroku/hk/cli"

func Setup() {
	if node.IsSetup() {
		return
	}
	cli.Stderrf("Setting up plugins... ")
	must(node.Setup())
	cli.Stderrln("done")
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
