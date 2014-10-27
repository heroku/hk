package main

import (
	"os"

	"github.com/heroku/hk/cli"
)

func help(args []string) {
	if len(args) > 0 && args[0] == "help" {
		help(args[1:])
		return
	}
	topic, command, _, _ := parse(args)
	switch {
	case command != nil:
		cli.Stderrf("USAGE: %s %s\n\n", os.Args[0], command.Signature)
		cli.Stderrln(command.Help)
		printTopicCommandsHelp(topic)
	case topic != nil:
		cli.Stderrf("USAGE: %s %s:COMMAND [--app APP] [command-specific-options]\n\n", os.Args[0], topic.Name)
		cli.Stderrln(topic.Help)
		printTopicCommandsHelp(topic)
	default:
		cli.Stderrf("USAGE: %s COMMAND [--app APP] [command-specific-options]\n\n", os.Args[0])
		cli.Stderrf("Help topics, type \"%s help TOPIC\" for more details:\n\n", os.Args[0])
		for _, topic := range topics {
			cli.Stderrf("  %-30s# %s\n", topic.Name, topic.ShortHelp)
		}
	}
	os.Exit(2)
}

func printTopicCommandsHelp(topic *cli.Topic) {
	if len(topic.Commands) > 0 {
		cli.Stderrf("\nCommands for %s, type \"%s help %s:COMMAND\" for more details:\n\n", topic.Name, os.Args[0], topic.Name)
		for _, command := range topic.Commands {
			cli.Stderrf("  %-30s# %s\n", command.Signature, command.ShortHelp)
		}
	}
}
