package main

import (
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/heroku/hk/cli"
	"github.com/heroku/hk/plugins"
)

var topics = cli.NewTopicSet(
	plugins.Topic,
)

func main() {
	defer handlePanic()
	plugins.Setup()
	for _, topic := range plugins.PluginTopics() {
		topics.AddTopic(topic)
	}
	topic, command, args, flags := parse(os.Args[1:])
	if command == nil {
		help(os.Args[1:])
	}
	runCommand(topic, command, args, flags)
}

func handlePanic() {
	if e := recover(); e != nil {
		switch e := e.(type) {
		case int:
			// This is for when we stub out ctx.Exit
			panic(e)
		}
		if e == "help" {
			help(os.Args[1:])
		}
		cli.Logf("ERROR: %s\n%s", e, debug.Stack())
		cli.Stderrln("ERROR:", e)
		cli.Exit(1)
	}
}

func runCommand(topic *cli.Topic, command *cli.Command, args []string, flags map[string]string) {
	cli.Logf("Running %s:%s %s\n", topic, command, args, flags)
	before := time.Now()
	command.Run(args, flags)
	cli.Logf("Finished in %s\n", (time.Since(before)))
}

func parse(input []string) (topic *cli.Topic, command *cli.Command, args []string, flags map[string]string) {
	if len(input) == 0 {
		return
	}
	tc := strings.SplitN(input[0], ":", 2)
	topic = topics[tc[0]]
	if topic != nil {
		command = topic.Commands[""]
		if len(tc) == 2 {
			command = topic.Commands[tc[1]]
		}
	}
	args = input[1:]
	return topic, command, args, flags
}
