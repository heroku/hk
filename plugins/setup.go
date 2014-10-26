package plugins

import "github.com/heroku/hk/cli"

func Setup() {
	if node.IsSetup() {
		return
	}
	cli.Stderrf("Setting up plugins... ")
	node.Setup()
	cli.Stderrln("done")
}
