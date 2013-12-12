package main

import (
	"fmt"
)

var cmdCreds = &Command{
	Run:      runCreds,
	Name:     "creds",
	Category: "hk",
	Short:    "show credentials" + extra,
	Long:     `Creds shows credentials that will be used for API calls.`,
}

func runCreds(cmd *Command, args []string) {
	fmt.Println(getCreds(apiURL))
}
