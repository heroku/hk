package plugins

import (
	"encoding/json"
	"os"

	"github.com/heroku/hk/cli"
)

func runFn(module, topic, command string) func(args []string, flags map[string]string) {
	return func(args []string, flags map[string]string) {
		script := `
		require('` + module + `')
		.topics.filter(function (topic) {
			return topic.name == '` + topic + `'
		})[0]
		.commands.filter(function (command) {
			return command.name == '` + command + `'
		})[0]
		.run([], {}, {
			"app": "dickey-xxx",
			"token": "` + os.Getenv("HEROKU_API_KEY") + `"
		})`

		cmd := node.RunScript(script)
		cmd.Stdout = cli.Stdout
		cmd.Stderr = cli.Stderr
		must(cmd.Run())
	}
}

func getPackageTopics(name string) []*cli.Topic {
	script := `console.log(JSON.stringify(require('` + name + `')))`
	cmd := node.RunScript(script)
	cmd.Stderr = cli.Stderr
	output, err := cmd.StdoutPipe()
	must(err)
	must(cmd.Start())
	var response map[string][]*cli.Topic
	must(json.NewDecoder(output).Decode(&response))
	must(cmd.Wait())
	topics := response["topics"]
	for _, topic := range topics {
		for _, command := range topic.Commands {
			command.Run = runFn(name, topic.Name, command.Name)
		}
	}
	return topics
}

func PluginTopics() (topics []*cli.Topic) {
	packages, err := node.Packages()
	must(err)
	for _, pkg := range packages {
		topics = append(topics, getPackageTopics(pkg.Name)...)
	}
	return topics
}
