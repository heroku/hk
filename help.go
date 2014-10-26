package main

import (
	"os"

	"github.com/dickeyxxx/gonpm/cli"
)

func help(command string, args ...string) {
	var topic *cli.Topic
	cli.Stderrf("USAGE: %s COMMAND [--app APP] [command-specific-options]\n\n", os.Args[0])

	if len(args) > 0 {
		topic = topicByName(args[0])
		if topic != nil {
			topic.Help(command, args...)
			return
		}
	}
	cli.Stderrf("Help topics, type \"%s help TOPIC\" for more details:\n\n", os.Args[0])
	for _, topic := range topics {
		cli.Stderrf("  %-30s# %s\n", topic.Name, topic.ShortHelp)
	}
}
