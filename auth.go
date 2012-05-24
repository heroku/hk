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
		errorf("%v", err)
	}
	fmt.Println(getCreds(u))
}
