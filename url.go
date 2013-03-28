package main

import (
	"fmt"
)

var cmdURL = &Command{
	Run:   runURL,
	Usage: "url [-a app]",
	Short: "show app url" + extra,
	Long:  `Prints the web URL for the app.`,
}

func init() {
	cmdURL.Flag.StringVar(&flagApp, "a", "", "app")
}

func runURL(cmd *Command, args []string) {
	fmt.Println("https://" + mustApp() + ".herokuapp.com/")
}
