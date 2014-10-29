package plugins

import (
	"encoding/json"
	"fmt"

	"github.com/heroku/hk/cli"
)

func runFn(module, topic, command string) func(ctx *cli.Context) {
	return func(ctx *cli.Context) {
		ctxJson, err := json.Marshal(ctx)
		must(err)
		script := fmt.Sprintf(`
		require('%s')
		.topics.filter(function (topic) {
			return topic.name == '%s'
		})[0]
		.commands.filter(function (command) {
			return command.name == '%s'
		})[0]
		.run(%s)`, module, topic, command, ctxJson)

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
