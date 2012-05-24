package main

import (
	"fmt"
	"net/url"
)

func credsHelp() {
	cmdHelp("hk creds", "Show API credentials")
}

func creds() {
	u, err := url.Parse(apiURL)
	if err != nil {
		error(err.Error())
	}
	fmt.Println(getCreds(u))
}
