package main

import (
	"fmt"
	"net/url"
)

var cmdCreds = &Command{
	Run:   runCreds,
	Usage: "creds",
	Short: "show auth creds",
	Long:  `Creds shows credentials that will be used for API calls.`,
}

func runCreds(cmd *Command, args []string) {
	u, err := url.Parse(apiURL)
	if err != nil {
		errorf("%v", err)
	}
	fmt.Println(getCreds(u))
}
