package main

import (
	"bufio"
	"code.google.com/p/go-netrc/netrc"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
)

const (
	Version = "0.0.1"
)

var (
	apiURL = "https://api.heroku.com"
	hkHome = os.Getenv("HOME") + "/.hk"
)

var stdin = bufio.NewReader(os.Stdin)

var updater = Updater{
	url: "https://github.com/downloads/kr/hk/",
	dir: hkHome + "/update/",
}

func getCreds(u *url.URL) (user, pass string) {
	if u.User != nil {
		pw, _ := u.User.Password()
		return u.User.Username(), pw
	}

	m, err := netrc.FindMachine(os.Getenv("HOME")+"/.netrc", u.Host)
	if err != nil {
		panic(err)
	}

	return m.Login, m.Password
}

// generic api requests
func apiReq(v interface{}, meth string, url string) {
	req, err := http.NewRequest(meth, url, nil)
	if err != nil {
		panic(err)
	}

	req.SetBasicAuth(getCreds(req.URL))
	req.Header.Add("User-Agent", fmt.Sprintf("hk/%s", Version))
	req.Header.Add("Accept", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	if res.StatusCode == 401 {
		error("Unauthorized")
	}
	if res.StatusCode == 403 {
		error("Unauthorized")
	}
	if res.StatusCode != 200 {
		fmt.Printf("%v\n", res)
		error("Unexpected error")
	}

	err = json.NewDecoder(res.Body).Decode(v)
	if err != nil {
		panic(err)
	}
}

// error formatting
func error(msg string) {
	fmt.Fprintf(os.Stderr, "Error: %s.\n", msg)
	os.Exit(1)
}

func unrecArg(arg, cmd string) {
	error(fmt.Sprintf("Unrecognized argument '%s'. See 'hk help %s'", arg, cmd))
}

func unrecCmd(cmd string) {
	error(fmt.Sprintf("'%s' is not an hk command. See 'hk help'", cmd))
}

// info formatting
func cmdHelp(usage string, desc string) {
	fmt.Printf("Usage: %s\n\n", usage)
	fmt.Printf("%s.\n", desc)
}

// commands
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

func infoHelp() {
	cmdHelp("hk info -a <app>", "Show app info")
}

func info() {
	if (len(os.Args) != 4) || (os.Args[2] != "-a") {
		error("Invalid usage. See 'hk help info'")
	}
	appName := os.Args[3]
	var info struct {
		Name   string
		Owner  string `json:"owner_email"`
		Stack  string
		GitURL string `json:"git_url"`
		WebURL string `json:"web_url"`
	}
	apiReq(&info, "GET", fmt.Sprintf(apiURL+"/apps/%s", appName))
	fmt.Printf("Name:     %s\n", info.Name)
	fmt.Printf("Owner:    %s\n", info.Owner)
	fmt.Printf("Stack:    %s\n", info.Stack)
	fmt.Printf("Git URL:  %s\n", info.GitURL)
	fmt.Printf("Web URL:  %s\n", info.WebURL)
}

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

func listHelp() {
	cmdHelp("hk list", "List accessible apps")
}

func list() {
	if len(os.Args) != 2 {
		unrecArg(os.Args[2], "list")
	}
	var apps []struct{ Name string }
	apiReq(&apps, "GET", apiURL+"/apps")
	for _, app := range apps {
		fmt.Printf("%s\n", app.Name)
	}
}

func psHelp() {
	cmdHelp("hk ps -a <app>", "List app processes")
}

type Proc struct {
	Name    string `json:"process"`
	State   string
	Command string
}

type Procs []*Proc

func (p Procs) Len() int           { return len(p) }
func (p Procs) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p Procs) Less(i, j int) bool { return p[i].Name < p[j].Name }

func ps() {
	if (len(os.Args) != 4) || (os.Args[2] != "-a") {
		error("Invalid usage. See 'hk help ps'")
	}
	appName := os.Args[3]
	var procs Procs
	apiReq(&procs, "GET", fmt.Sprintf(apiURL+"/apps/%s/ps", appName))
	sort.Sort(procs)
	fmt.Printf("Process           State       Command\n")
	fmt.Printf("----------------  ----------  ------------------------\n")
	for _, proc := range procs {
		fmt.Printf("%-16s  %-10s  %s\n", proc.Name, proc.State, proc.Command)
	}
}

func fetchUpdateHelp() {
	cmdHelp("hk fetch-update", "Download the latest hk client")
}

func fetchUpdate() {
	if len(os.Args) != 2 {
		unrecArg(os.Args[2], "fetch-update")
	}

	updater.fetchAndApply()
}

func versionHelp() {
	cmdHelp("hk version", "Show hk client version")
}

func version() {
	if len(os.Args) != 2 {
		unrecArg(os.Args[2], "version")
	}
	fmt.Printf("%s\n", Version)
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
		case "info":
			infoHelp()
		case "creds":
			credsHelp()
		case "list":
			listHelp()
		case "ps":
			psHelp()
		case "fetch-update":
			fetchUpdateHelp()
		case "version":
			versionHelp()
		default:
			unrecCmd(cmd)
		}
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
	fmt.Printf("  stop            Stop a process\n")
	fmt.Printf("  creds           Show auth creds\n")
	fmt.Printf("  unset           Unset config vars\n")
	fmt.Printf("  version         Display version\n\n")
	fmt.Printf("See 'hk help <command>' for more information on a specific command.\n")
}

// entry point
func main() {
	if s := os.Getenv("HEROKU_API_URL"); s != "" {
		apiURL = strings.TrimRight(s, "/")
	}

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
		case "info":
			info()
		case "creds":
			creds()
		case "list":
			list()
		case "ps":
			ps()
		case "fetch-update":
			fetchUpdate()
		case "version":
			version()
		default:
			unrecCmd(cmd)
		}
	}

	updater.run()
}
