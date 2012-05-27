package main

import (
	"bufio"
	"code.google.com/p/go-netrc/netrc"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
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

type Command struct {
	// args does not include the command name
	Run func(cmd *Command, args []string)

	Usage string // first word is the command name
	Short string // `hk help` output
	Long  string // `hk help <cmd>` output
}

func (c *Command) Name() string {
	name := c.Usage
	i := strings.Index(name, " ")
	if i >= 0 {
		name = name[:i]
	}
	return name
}

// Running `hk help` will list commands in this order.
var commands = []*Command{
	cmdCreds,
	cmdEnv,
	cmdFetchUpdate,
	cmdGet,
	cmdInfo,
	cmdList,
	cmdOpen,
	cmdPs,
	cmdVersion,
	cmdHelp,
}

var flagApp = flag.String("a", "", "app")

func main() {
	defer updater.run() // doesn't run if os.Exit is called

	if s := os.Getenv("HEROKU_API_URL"); s != "" {
		apiURL = strings.TrimRight(s, "/")
	}

	log.SetFlags(0)
	flag.Usage = usage
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		usage()
	}

	name := args[0]
	os.Args = args
	flag.Parse()
	args = flag.Args()

	for _, cmd := range commands {
		if cmd.Name() == name {
			cmd.Run(cmd, args)
			return
		}
	}

	fmt.Fprintf(os.Stderr, "Unknown command: %s\n", name)
	usage()
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
	req.Header.Add("User-Agent", "hk/"+Version)
	req.Header.Add("Accept", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	if res.StatusCode == 401 {
		log.Fatal("Unauthorized")
	}
	if res.StatusCode == 403 {
		log.Fatal("Unauthorized")
	}
	if res.StatusCode != 200 {
		fmt.Printf("%v\n", res)
		log.Fatal("Unexpected error")
	}

	err = json.NewDecoder(res.Body).Decode(v)
	if err != nil {
		panic(err)
	}
}

func app() string {
	if *flagApp != "" {
		return *flagApp
	}
	out, err := exec.Command("git", "remote", "show", "-n", "heroku").Output()
	if err != nil {
		log.Fatal(err)
	}
	s := string(out)
	const sign = "Fetch URL: "
	i := strings.Index(s, sign)
	if i < 0 {
		log.Fatal("could not find git remote named 'heroku'")
	}
	s = s[i+len(sign):]
	i = strings.Index(s, "\n")
	if i >= 0 {
		s = s[:i]
	}
	const pre = "git@heroku.com:"
	const suf = ".git"
	if !strings.HasPrefix(s, pre) || !strings.HasSuffix(s, suf) {
		log.Fatal("could not find app name in heroku git remote")
	}
	return s[len(pre) : len(s)-len(suf)]
}
