package plugins

import (
	"os"
	"path/filepath"

	"github.com/dickeyxxx/gonpm/cli"
)

func init() {
	os.Setenv("NODE_PATH", filepath.Join(cli.AppDir, "lib", "node_modules"))
	os.Setenv("NPM_CONFIG_GLOBAL", "true")
	os.Setenv("NPM_CONFIG_PREFIX", cli.AppDir)
	os.Setenv("NPM_CONFIG_SPINNER", "false")
}

var Topic = &cli.Topic{
	Name:      "plugins",
	ShortHelp: "manage plugins",
	Run:       Run,
	Help:      Help,
}

func Run(command string, args ...string) {
	switch command {
	case "install":
		if len(args) != 1 {
			Help(command, args...)
			cli.Exit(2)
		}
		install(args[0])
	case "list":
	case "":
		list()
	default:
		Help(command, args...)
	}
}

func Help(command string, args ...string) {
	cli.Stderrln("TODO: help for " + command)
}
