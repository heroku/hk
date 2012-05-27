package main

import (
	"fmt"
	"log"
)

var cmdEnv = &Command{
	Run:   runEnv,
	Usage: "env",
	Short: "list config vars",
	Long:  `Show all config vars.`,
}

func runEnv(cmd *Command, args []string) {
	var config map[string]string
	apiReq(&config, "GET", fmt.Sprintf(apiURL+"/apps/%s/config_vars", app()))
	for k, v := range config {
		fmt.Printf("%s=%s\n", k, v)
	}
}

var cmdGet = &Command{
	Run:   runGet,
	Usage: "get <var>",
	Short: "get config var",
	Long:  `Get the value of a config var.`,
}

func runGet(cmd *Command, args []string) {
	if len(args) != 1 {
		log.Fatal("Invalid usage. See 'hk help get'")
	}
	var config map[string]string
	apiReq(&config, "GET", fmt.Sprintf(apiURL+"/apps/%s/config_vars", app()))
	value, found := config[args[0]]
	if !found {
		log.Fatalf("No such key as '%s'", args[0])
	}
	fmt.Println(value)
}
