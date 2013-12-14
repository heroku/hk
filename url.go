package main

import (
	"fmt"
)

var cmdURL = &Command{
	Run:      runURL,
	Name:     "url",
	Usage:    "[-a <app>]",
	Category: "app",
	Short:    "show app url" + extra,
	Long:     `Prints the web URL for the app.`,
}

func init() {
	cmdURL.Flag.StringVar(&flagApp, "a", "", "app name")
}

func runURL(cmd *Command, args []string) {
	fmt.Println("https://" + mustApp() + ".herokuapp.com/")
}
