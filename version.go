package main

import (
	"fmt"
	"runtime"

	"github.com/heroku/hk/cli"
)

var version = &cli.Topic{
	Name:      "version",
	ShortHelp: "print the version",
	Commands: []*cli.Command{
		{
			ShortHelp: "print the version", Run: func(ctx *cli.Context) {
				fmt.Printf("heroku-toolbelt/%s (%s-%s) %s\n", Version, runtime.GOARCH, runtime.GOOS, runtime.Version())
			},
		},
	},
}
