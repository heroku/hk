package plugins

import (
	"os"
	"strings"

	"github.com/dickeyxxx/gode"
	"github.com/heroku/hk/cli"
)

type Plugin struct {
	*gode.Package
	*cli.Topic
}

func pluginRun(name string) func(command string, args ...string) {
	return func(command string, args ...string) {
		context := `{
			"app": "dickey-xxx",
			"token": "` + os.Getenv("HEROKU_API_KEY") + `"
		}`
		script := `require('` + name + `').run("` + command + `", [], {}, ` + context + `)`
		cmd := node.RunScript(script)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		must(cmd.Run())
	}
}

func pluginShortHelp(name string) string {
	script := `console.log(require('` + name + `').shortHelp)`
	output, err := node.RunScript(script).Output()
	must(err)
	return strings.TrimSpace(string(output))
}

func pluginHelp(name string) func(command string, args ...string) {
	return func(command string, args ...string) {
		script := `require('` + name + `').help()`
		cmd := node.RunScript(script)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		must(cmd.Run())
	}
}

func Plugins() []*Plugin {
	var plugins []*Plugin
	packages, err := node.Packages()
	must(err)
	for _, pkg := range packages {
		plugin := &Plugin{
			Package: &pkg,
			Topic: &cli.Topic{
				Name: pkg.Name,
			},
		}
		plugins = append(plugins, plugin)
	}
	return plugins
}
