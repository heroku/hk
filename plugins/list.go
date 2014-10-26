package plugins

import "github.com/heroku/hk/cli"

func list() {
	cli.Logln("Listing plugins...")
	for _, plugin := range Plugins() {
		cli.Stdoutln(plugin.Package.Name, plugin.Version)
	}
}
