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
		cli.Errf("USAGE: %s %s\n\n", os.Args[0], command.Signature)
		cli.Errln(command.Help)
		printTopicCommandsHelp(topic)
	case topic != nil:
		cli.Errf("USAGE: %s %s:COMMAND [--app APP] [command-specific-options]\n\n", os.Args[0], topic.Name)
		cli.Errln(topic.Help)
		printTopicCommandsHelp(topic)
	default:
		cli.Errf("USAGE: %s COMMAND [--app APP] [command-specific-options]\n\n", os.Args[0])
		cli.Errf("Help topics, type \"%s help TOPIC\" for more details:\n\n", os.Args[0])
		for _, topic := range topics {
			cli.Errf("  %s %-30s# %s\n", os.Args[0], topic.Name, topic.ShortHelp)
		}
	}
	os.Exit(2)
}

func printTopicCommandsHelp(topic *cli.Topic) {
	if len(topic.Commands) > 0 {
		cli.Errf("\nCommands for %s, type \"%s help %s:COMMAND\" for more details:\n\n", topic.Name, os.Args[0], topic.Name)
		for _, command := range topic.Commands {
			cli.Errf("  %s %-30s# %s\n", os.Args[0], command.Signature, command.ShortHelp)
		}
	}
}
