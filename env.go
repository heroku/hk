package main

import (
	"fmt"
	"os"
)

func envHelp() {
	cmdHelp("hk env -a <app>", "Show all config vars")
}

func env() {
	if (len(os.Args) != 4) || (os.Args[2] != "-a") {
		error("Invalid usage. See 'hk help env'")
	}
	appName := os.Args[3]
	var config map[string]string
	apiReq(&config, "GET", fmt.Sprintf(apiURL+"/apps/%s/config_vars", appName))
	for k, v := range config {
		fmt.Printf("%s=%s\n", k, v)
	}
}

func getHelp() {
	cmdHelp("hk get -a <app> <key>", "Get the value of a config var")
}

func get() {
	if (len(os.Args) != 5) || (os.Args[2] != "-a") {
		error("Invalid usage. See 'hk help get'")
	}
	appName := os.Args[3]
	key := os.Args[4]
	var config map[string]string
	apiReq(&config, "GET", fmt.Sprintf(apiURL+"/apps/%s/config_vars", appName))
	value, found := config[key]
	if !found {
		error(fmt.Sprintf("No such key as '%s'", key))
	}
	fmt.Println(value)
}
