package main

import (
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/dickeyxxx/gonpm/cli"
	"github.com/dickeyxxx/gonpm/plugins"
)

var topics []*cli.Topic = []*cli.Topic{
	plugins.Topic,
}

func main() {
	defer handlePanic()
	plugins.Setup()
	for _, plugin := range plugins.ListPlugins() {
		topics = append(topics, plugin.Topic)
	}
	topicName, command, args := parse(os.Args[1:])
	topic := topicByName(topicName)
	if topic == nil {
		help(command, args...)
		cli.Exit(2)
	}
	cli.Logf("Running %s:%s %s\n", topicName, command, args)
	before := time.Now()
	topic.Run(command, args...)
	cli.Logf("Finished in %s\n", (time.Since(before)))
}

func handlePanic() {
	if e := recover(); e != nil {
		switch e := e.(type) {
		case int:
			// This is for when we stub out ctx.Exit
			panic(e)
		default:
			cli.Logf("ERROR: %s\n%s", e, debug.Stack())
			cli.Stderrln("ERROR:", e)
			cli.Exit(1)
		}
	}
}

func parse(input []string) (topic, command string, args []string) {
	if len(input) == 0 {
		return
	}
	tc := strings.SplitN(input[0], ":", 2)
	topic = tc[0]
	if len(tc) == 2 {
		command = tc[1]
	}
	args = input[1:]
	return topic, command, args
}

func topicByName(name string) *cli.Topic {
	for _, topic := range topics {
		if name == topic.Name {
			return topic
		}
	}
	return nil
}
