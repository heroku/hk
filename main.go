package main

import (
	"bufio"
	"code.google.com/p/go-netrc/netrc"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
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
