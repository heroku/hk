package plugins

import (
	"os"
	"strings"

	"github.com/dickeyxxx/gode"
	"github.com/heroku/hk/cli"
)

type Plugin struct {
	gode.Package
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

func topicFromPackage(pkg gode.Package) *cli.Topic {
	return &cli.Topic{
		Name: pkg.Name,
	}
}

func Plugins() (plugins []*Plugin) {
	packages, err := node.Packages()
	must(err)
	for _, pkg := range packages {
		plugins = append(plugins, pluginFromPackage(pkg))
	}
	return plugins
}

func pluginFromPackage(pkg gode.Package) *Plugin {
	return &Plugin{
		Package: pkg,
	}
}

func PluginTopics() (topics []*cli.Topic) {
	packages, err := node.Packages()
	must(err)
	for _, pkg := range packages {
		topics = append(topics, topicFromPackage(pkg))
	}
	return topics
}
