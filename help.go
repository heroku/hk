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
	case command != nil && command.Name == "":
		cli.Errf("USAGE: %s %s\n\n", os.Args[0], topic.Name)
		cli.Errln(command.Help)
		// This is a root command so show the other commands in the topic
		printTopicCommandsHelp(topic)
	case command != nil:
		cli.Errf("USAGE: %s %s:%s\n\n", os.Args[0], topic.Name, command.Name)
		cli.Errln(command.Help)
	case topic != nil:
		cli.Errf("USAGE: %s %s:COMMAND [--app APP] [command-specific-options]\n\n", os.Args[0], topic.Name)
		cli.Errln(topic.Help)
		printTopicCommandsHelp(topic)
	default:
		cli.Errf("USAGE: %s COMMAND [--app APP] [command-specific-options]\n\n", os.Args[0])
		cli.Errf("Help topics, type \"%s help TOPIC\" for more details:\n\n", os.Args[0])
		for _, topic := range nonHiddenTopics(topics) {
			cli.Errf("  %s %-30s# %s\n", os.Args[0], topic.Name, topic.ShortHelp)
		}
	}
	os.Exit(2)
}

func printTopicCommandsHelp(topic *cli.Topic) {
	if len(topic.Commands) > 0 {
		cli.Errf("\nCommands for %s, type \"%s help %s:COMMAND\" for more details:\n\n", topic.Name, os.Args[0], topic.Name)
		for _, command := range nonHiddenCommands(topic.Commands) {
			if command.Name == "" {
				cli.Errf("  %s %s                               # %s\n", os.Args[0], topic.Name, command.ShortHelp)
			} else {
				cli.Errf("  %s %s:%-30s# %s\n", os.Args[0], topic.Name, command.Name, command.ShortHelp)
			}
		}
	}
}

func nonHiddenTopics(from cli.TopicSet) []*cli.Topic {
	to := make([]*cli.Topic, 0, len(from))
	for _, topic := range from {
		if !topic.Hidden {
			to = append(to, topic)
		}
	}
	return to
}

func nonHiddenCommands(from []*cli.Command) []*cli.Command {
	to := make([]*cli.Command, 0, len(from))
	for _, command := range from {
		if !command.Hidden {
			to = append(to, command)
		}
	}
	return to
}
