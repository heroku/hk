package main

import (
	"fmt"
	"log"
	"strings"
)

var cmdEnv = &Command{
	Run:   runEnv,
	Usage: "env",
	Short: "list config vars",
	Long:  `Show all config vars.`,
}

func runEnv(cmd *Command, args []string) {
	var config map[string]string
	getApiReq(&config, fmt.Sprintf(apiURL+"/apps/%s/config_vars", app()))
	for k, v := range config {
		fmt.Printf("%s=%s\n", k, v)
	}
}

var cmdGet = &Command{
	Run:   runGet,
	Usage: "get <name>",
	Short: "get config var",
	Long: `
Get the value of a config var.

Example:

  $ hk get BUILDPACK_URL
  http://github.com/kr/heroku-buildpack-inline.git
`,
}

func runGet(cmd *Command, args []string) {
	if len(args) != 1 {
		log.Fatal("Invalid usage. See 'hk help get'")
	}
	var config map[string]string
	getApiReq(&config, fmt.Sprintf(apiURL+"/apps/%s/config_vars", app()))
	value, found := config[args[0]]
	if !found {
		log.Fatalf("No such key as '%s'", args[0])
	}
	fmt.Println(value)
}

var cmdSet = &Command{
	Run:   runSet,
	Usage: "set <name>=<value> ...",
	Short: "set config var",
	Long: `
Set the value of a config var.

Example:

  $ hk set BUILDPACK_URL=http://github.com/kr/heroku-buildpack-inline.git
`,
}

func runSet(cmd *Command, args []string) {
	config := make(map[string]string)
	for _, arg := range args {
		i := strings.Index(arg, "=")
		if i < 0 {
			log.Fatalf("bad format: %#q. See 'hk help set'", arg)
		}
		config[arg[:i]] = arg[i+1:]
	}
	apiReqJson(nil, "PUT", apiURL+"/apps/"+app()+"/config_vars", config)
}
