package main

import (
	"fmt"
)

var cmdURL = &Command{
	Run:      runURL,
	Usage:    "url",
	Category: "app",
	Short:    "show app url" + extra,
	Long:     `Prints the web URL for the app.`,
}

func runURL(cmd *Command, args []string) {
	fmt.Println("https://" + mustApp() + ".herokuapp.com/")
}
