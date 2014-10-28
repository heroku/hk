package plugins

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/dickeyxxx/gode"
	"github.com/heroku/hk/cli"
)

type Plugin struct {
	gode.Package
}

func (p *Plugin) Topics() (topics []*cli.Topic) {
	script := `
	var commands = require('` + p.Name + `').commands
	console.log(JSON.stringify(commands))`
	var response map[string]map[string]*cli.Command
	cmd := node.RunScript(script)
	cmd.Stderr = cli.Stderr
	output, err := cmd.StdoutPipe()
	must(err)
	must(cmd.Start())
	must(json.NewDecoder(output).Decode(&response))
	must(cmd.Wait())
	for topicName, topic := range response {
		topic := &cli.Topic{Name: topicName, Commands: topic}
		for name, command := range topic.Commands {
			command.Run = runFn(p.Name, topic.Name, name)
		}
		topics = append(topics, topic)
	}
	return topics
}

func runFn(module, topic, command string) func(args []string, flags map[string]string) {
	return func(args []string, flags map[string]string) {
		context := `{
			"app": "dickey-xxx",
			"token": "` + os.Getenv("HEROKU_API_KEY") + `"
		}`
		script := `
		var commands = require('` + module + `').commands
		commands['` + topic + `']['` + command + `'].run([], {}, ` + context + `)`
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
	for _, plugin := range Plugins() {
		topics = append(topics, plugin.Topics()...)
	}
	return topics
}
