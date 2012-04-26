package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

const (
	VERSION = "0.0.1"
)

func apiReq(meth string, url string) (res *http.Response) {
	client := &http.Client{}
	req, err := http.NewRequest(meth, url, nil)
	req.SetBasicAuth("x", os.Getenv("HEROKU_API_KEY"))
	req.Header.Add("User-Agent", fmt.Sprintf("hk/%s", VERSION))
	req.Header.Add("Accept", "application/json")
	res, err = client.Do(req)
		if err != nil {
		panic(err)
	}
	if (res.StatusCode == 401) {
		error("Unauthorized")
	}
	if (res.StatusCode != 200) {
		error("Unexpected error")
	}
	return res
}

func error(msg string) {
	fmt.Fprintf(os.Stderr, "Error: %s.\n", msg)
	os.Exit(1)
}

func unrecArg(arg string) {
	error(fmt.Sprintf("Unrecognized argument '%s'", arg))
}

func unrecCmd(cmd string) {
	error(fmt.Sprintf("'%s' is not an hk command. See 'hk help'", cmd))
}

func envHelp() {
	fmt.Printf("Usage: hk env -a <app>\n\n")
	fmt.Printf("Show all config vars.")
	os.Exit(0)
}

func env() {
	if (len(os.Args) != 4) || (os.Args[2] != "-a") {
		error("Invalid usage. See hk help env")
	}
	appName := os.Args[3]
	res := apiReq("GET", fmt.Sprintf("https://api.heroku.com/apps/%s/config_vars", appName))
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	var data interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		panic(err)
	}
	config := data.(map[string]interface{})
	for k, v := range config {
		fmt.Printf("%s=%v\n", k, v)
	}
	os.Exit(0)
}

func getHelp() {
	fmt.Printf("Usage: hk get -a <app> <key>\n\n")
	fmt.Printf("Get the value of a config var.\n")
	os.Exit(0)
}

func get() {
	if (len(os.Args) != 5) || (os.Args[2] != "-a") {
		error("Invalid usage. See hk help get")
	}
	appName := os.Args[3]
	key := os.Args[4]
	res := apiReq("GET", fmt.Sprintf("https://api.heroku.com/apps/%s/config_vars", appName))
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	var data interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		panic(err)
	}
	config := data.(map[string]interface{})
	value, found := config[key]
	if !found {
		error(fmt.Sprintf("No such key as '%s'", key))
	}
	fmt.Println(value)
	os.Exit(0)
}

func listHelp() {
	fmt.Printf("Usage: hk list\n\n")
	fmt.Printf("List accessible apps.\n")
	os.Exit(0)
}

func list() {
	if len(os.Args) != 2 {
		unrecArg(os.Args[2])
	}
	res := apiReq("GET", "https://api.heroku.com/apps")
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	var data interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		panic(err)
	}
	apps := data.([]interface{})
	for i := range apps {
		app := apps[i].(map[string]interface{})
		fmt.Printf("%s\n", app["name"])
	}
	os.Exit(0)
}

func psHelp() {
	fmt.Printf("Usage: hk ps -a <app>\n\n")
	fmt.Printf("List apps processes.\n")
	os.Exit(0)
}

func ps() {
	if (len(os.Args) != 4) || (os.Args[2] != "-a") {
		error("Invalid usage. See hk help ps")
	}
	appName := os.Args[3]
	res := apiReq("GET", fmt.Sprintf("https://api.heroku.com/apps/%s/ps", appName))
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	var data interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		panic(err)
	}
	processes := data.([]interface{})
	for i := range processes {
		process := processes[i].(map[string]interface{})
		fmt.Printf("%v\n", process)
	}
	os.Exit(0)
}

func versionHelp() {
	fmt.Printf("Usage: hk version\n\n")
	fmt.Printf("Show hk client version.\n")
	os.Exit(0)
}

func version() {
	if len(os.Args) != 2 {
	  unrecArg(os.Args[2])
  }
	fmt.Printf("%s\n", VERSION)
	os.Exit(0)
}

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
	fmt.Printf("  get             Get config var\n")
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
	os.Exit(0)
}

func help() {
	if len(os.Args) <= 2 {
		usage()
	} else {
		cmd := os.Args[2]
		switch cmd {
	  case "env":
		  envHelp()
	  case "get":
			getHelp()
		case "list":
		  listHelp()
	  case "ps":
		  psHelp()
		case "version":
			versionHelp()
		}
		unrecCmd(cmd)
	}
}

func main() {
	if len(os.Args) <= 1 {
		usage()
	} else {
		cmd := os.Args[1]
		switch cmd {
		case "env":
				env()
		case "get":
			get()
		case "help":
			help()
		case "list":
			list()
		case "ps":
			ps()
		case "version":
			version()
		}
		unrecCmd(cmd)
	}
}
