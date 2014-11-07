package main

import (
	"os"
	"strings"

	"github.com/heroku/hk/cli"
)

func help() {
	args := os.Args[1:]
	if len(args) > 0 && args[0] == "help" {
		args = args[1:]
	}
	ctx, _ := Cli.Parse(args)
	cli.Errln("hk version:", Version)
	switch {
	case ctx.Topic == nil:
		cli.Errf("USAGE: heroku COMMAND [--app APP] [command-specific-options]\n\n")
		cli.Errf("Help topics, type \"heroku help TOPIC\" for more details:\n\n")
		for _, topic := range nonHiddenTopics(Cli.Topics) {
			cli.Errf("  heroku %-30s# %s\n", topic.Name, topic.ShortHelp)
		}
	case ctx.Command == nil:
		cli.Errf("USAGE: heroku %s:COMMAND [--app APP] [command-specific-options]\n\n", ctx.Topic.Name)
		cli.Errln(ctx.Topic.Help)
		printTopicCommandsHelp(ctx.Topic)
	case ctx.Command.Name == "":
		cli.Errf("USAGE: heroku %s\n\n", commandSignature(ctx.Topic, ctx.Command))
		cli.Errln(ctx.Command.Help)
		// This is a root command so show the other commands in the topic
		printTopicCommandsHelp(ctx.Topic)
	default:
		cli.Errf("USAGE: heroku %s\n\n", commandSignature(ctx.Topic, ctx.Command))
		cli.Errln(ctx.Command.Help)
	}
	os.Exit(2)
}

func printTopicCommandsHelp(topic *cli.Topic) {
	if len(topic.Commands) > 0 {
		cli.Errf("\nCommands for %s, type \"heroku help %s:COMMAND\" for more details:\n\n", topic.Name, topic.Name)
		for _, command := range nonHiddenCommands(topic.Commands) {
			cli.Errf(" heroku %-30s # %s\n", commandSignature(topic, command), command.ShortHelp)
		}
	}
}

func commandSignature(topic *cli.Topic, command *cli.Command) string {
	cmd := topic.Name
	if command.Name != "" {
		cmd = cmd + ":" + command.Name
	}
	cmd = cmd + commandArgs(command)
	if command.NeedsApp {
		cmd = cmd + " --app APP"
	}
	return cmd
}

func commandArgs(command *cli.Command) string {
	args := ""
	for _, arg := range command.Args {
		if arg.Optional {
			args = args + " [" + strings.ToUpper(arg.Name) + "]"
		} else {
			args = args + " " + strings.ToUpper(arg.Name)
		}
	}
	return args
}
func nonHiddenTopics(from map[string]*cli.Topic) []*cli.Topic {
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

func AppNeededWarning() {
	cli.Errln(" !    No app specified.")
	cli.Errln(" !    Run this command from an app folder or specify which app to use with --app APP.")
	os.Exit(3)
}
