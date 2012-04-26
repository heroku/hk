package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

func usage() {
	fmt.Printf("Usage: hk <command> [-a <app>] [command-specific-options]\n\n")
	fmt.Printf("Supported hk commands are:\n")
	fmt.Printf("  addons          List add-ons\n")
	fmt.Printf("  addons-add      Add an add-on\n")
	fmt.Printf("  addons-open     Open an add-on page\n")
	fmt.Printf("  addons-remove   Remove an add-on \n")
	fmt.Printf("  create          Create an app\n")
	fmt.Printf("  destroy         Destroy an app\n")
	fmt.Printf("  env             List config vars\n")
	fmt.Printf("  get             GEt config var\n")
	fmt.Printf("  help            Show this help\n")
	fmt.Printf("  info            Show app info\n")
	fmt.Printf("  list            List apps\n")
	fmt.Printf("  login           Log in\n")
	fmt.Printf("  logout          Log out\n")
	fmt.Printf("  logs            Show logs\n")
	fmt.Printf("  open            Open app\n")
	fmt.Printf("  pg              List databases\n")
	fmt.Printf("  pg-info         Show database info\n")
	fmt.Printf("  pg-promote      Promote a database\n")
	fmt.Printf("  ps-psql         Open a psql database shell\n")
	fmt.Printf("  pg-wait         Await a database\n")
	fmt.Printf("  ps              List processes\n")
	fmt.Printf("  release         Show release info\n")
	fmt.Printf("  releases        List releases\n")
	fmt.Printf("  rename          Rename an app\n")
	fmt.Printf("  restart         Restart processes\n")
	fmt.Printf("  rollback        Rollback to a previous release\n")
	fmt.Printf("  run             Run a process\n")
	fmt.Printf("  set             Set config var\n")
	fmt.Printf("  scale           Scale processes\n")
	fmt.Printf("  run             Run a process\n")
	fmt.Printf("  stop            Stop a process\n")
	fmt.Printf("  token           Show auth token\n")
	fmt.Printf("  unset           Unset config vars\n")
	fmt.Printf("  version         Display version\n\n")
	fmt.Printf("See 'hk help <command>' for more information on a specific command.\n")
}

func help(args []string) {
	if len(args) == 2 {
		usage()
	} else if args[2] == "version" {
		versionHelp()
	} else if args[2] == "list" {
		listHelp()
	} else {
		fmt.Fprintf(os.Stderr, "Error: '%s' is not an hk command. See 'hk help'.\n", args[2])
	  os.Exit(1)
	}
}

func listHelp() {
	fmt.Printf("Usage: hk list\n\n")
	fmt.Printf("List accessible apps.\n")
}

func list(args []string) {
	if len(args) == 2 {
		client := &http.Client{}
		req, err := http.NewRequest("GET", "https://api.heroku.com/apps", nil)
		req.SetBasicAuth("x", os.Getenv("HEROKU_API_KEY"))
		req.Header.Add("User-Agent", "hk/0.0.1")
		req.Header.Add("Accept", "application/json")
		res, err := client.Do(req)
			if err != nil {
			panic(err)
		}
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			panic(err)
		}
		var data interface{}
		err = json.Unmarshal(body, &data)
		if err != nil {
			panic(err)
		}
		if (res.StatusCode == 401) {
		  fmt.Fprintf(os.Stderr, "Error: Unauthorized.\n")
		  os.Exit(1)
		} else if (res.StatusCode != 200) {
			fmt.Fprintf(os.Stderr, "Error: Internal Server Error.\n")
			os.Exit(1)
		} else {
			apps := data.([]interface{})
			for i := range apps {
				app := apps[i].(map[string]interface{});
		    fmt.Printf("%s\n", app["name"])
		  }
		}
	} else {
		fmt.Fprintf(os.Stderr, "Error: Unrecognized argument '%s'.\n", args[2])
	}
}

func versionHelp() {
	fmt.Printf("Usage: hk version\n\n")
	fmt.Printf("Show hk client version.\n")
}

func version(args []string) {
	if len(args) == 2 {
		fmt.Printf("0.0.1\n")
	} else {
		fmt.Fprintf(os.Stderr, "Error: Unrecognized argument '%s'.\n", args[2])
		os.Exit(1)
	}
}

func main() {
	args := os.Args;
	if len(args) <= 1 {
		usage()
	} else if args[1] == "help" {
		help(args)
	} else if args[1] == "list" {
		list(args)
	} else if args[1] == "version" {
		version(args)
  } else {
	  fmt.Fprintf(os.Stderr, "Error: '%s' is not an hk command. See 'hk help'.\n", args[2])
	  os.Exit(1)
	}
}
