package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
)

const (
	VERSION = "0.0.1"
)

// sort interface-typed arrays by first-class functions
type ByFn struct{
	elems []interface{}
	comp func(a, b interface{}) bool
}
func (c ByFn) Len() int { return len(c.elems)  }
func (c ByFn) Less(i, j int) bool { return c.comp(c.elems[i], c.elems[j]) }
func (c ByFn) Swap(i, j int) { c.elems[i], c.elems[j] = c.elems[j], c.elems[i] }


// generic api requests
func apiReq(meth string, url string) interface{} {
	client := &http.Client{}
	req, err := http.NewRequest(meth, url, nil)
	req.SetBasicAuth("x", os.Getenv("HEROKU_API_KEY"))
	req.Header.Add("User-Agent", fmt.Sprintf("hk/%s", VERSION))
	req.Header.Add("Accept", "application/json")
	res, err := client.Do(req)
		if err != nil {
		panic(err)
	}
	if (res.StatusCode == 401) {
		error("Unauthorized")
	}
	if (res.StatusCode != 200) {
		error("Unexpected error")
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
	return data
}

// error formatting
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

// info formatting
func cmdHelp(usage string, desc string) {
	fmt.Printf("Usage: %s\n\n", usage)
	fmt.Printf("%s.\n", desc)
	os.Exit(0)
}

// commands
func envHelp() {
	cmdHelp("hk env -a <app>", "Show all config vars")
}

func env() {
	if (len(os.Args) != 4) || (os.Args[2] != "-a") {
		error("Invalid usage. See hk help env")
	}
	appName := os.Args[3]
	data := apiReq("GET", fmt.Sprintf("https://api.heroku.com/apps/%s/config_vars", appName))
	config := data.(map[string]interface{})
	for k, v := range config {
		fmt.Printf("%s=%v\n", k, v)
	}
	os.Exit(0)
}

func getHelp() {
	cmdHelp("hk get -a <app> <key>", "Get the value of a config var")
}

func get() {
	if (len(os.Args) != 5) || (os.Args[2] != "-a") {
		error("Invalid usage. See hk help get")
	}
	appName := os.Args[3]
	key := os.Args[4]
	data := apiReq("GET", fmt.Sprintf("https://api.heroku.com/apps/%s/config_vars", appName))
	config := data.(map[string]string)
	value, found := config[key]
	if !found {
		error(fmt.Sprintf("No such key as '%s'", key))
	}
	fmt.Println(value)
	os.Exit(0)
}

func listHelp() {
	cmdHelp("hk list", "List accessible apps")
}

func list() {
	if len(os.Args) != 2 {
		unrecArg(os.Args[2])
	}
	data := apiReq("GET", "https://api.heroku.com/apps")
	apps := data.([]interface{})
	for i := range apps {
		app := apps[i].(map[string]interface{})
		fmt.Printf("%s\n", app["name"])
	}
	os.Exit(0)
}

func psHelp() {
	cmdHelp("hk ps -a <app>", "List app processes")
}

func ps() {
	if (len(os.Args) != 4) || (os.Args[2] != "-a") {
		error("Invalid usage. See hk help ps")
	}
	appName := os.Args[3]
	data := apiReq("GET", fmt.Sprintf("https://api.heroku.com/apps/%s/ps", appName))
	processes := data.([]interface{})
	sort.Sort(ByFn{
		processes,
		func(a, b interface{}) bool {
			p1 := a.(map[string]interface{})["process"].(string)
			p2 := b.(map[string]interface{})["process"].(string)
		  return p1 < p2
	  }})
	fmt.Printf("Process           State       Command\n")
	fmt.Printf("----------------  ----------  ------------------------\n")
	for i := range processes {
		process := processes[i].(map[string]interface{})
		fmt.Printf("%-16s  %-10s  %s\n", process["process"], process["state"], process["command"])
	}
	os.Exit(0)
}

func versionHelp() {
	cmdHelp("hk version", "Show hk client version")
}

func version() {
	if len(os.Args) != 2 {
	  unrecArg(os.Args[2])
  }
	fmt.Printf("%s\n", VERSION)
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

// top-level usage
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

// entry point
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
