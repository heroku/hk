package main

import (
	"bufio"
	"bytes"
	"code.google.com/p/go-netrc/netrc"
	"encoding/json"
	"flag"
	"fmt"
	"io"
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
	cmdCreate,
	cmdCreds,
	cmdEnv,
	cmdFetchUpdate,
	cmdGet,
	cmdSet,
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
		log.Fatal(err)
	}

	return m.Login, m.Password
}

func apiReqJson(v interface{}, meth, url string, data interface{}) {
	body, err := json.Marshal(data)
	if err != nil {
		log.Fatal(err)
	}
	apiReq(v, meth, url, "application/json", bytes.NewReader(body))
}

func apiReqForm(v interface{}, meth, url string, data url.Values) {
	body := strings.NewReader(data.Encode())
	apiReq(v, meth, url, "application/x-www-form-urlencoded", body)
}

func getApiReq(v interface{}, url string) {
	apiReq(v, "GET", url, "", nil)
}

func apiReq(v interface{}, meth, url, contentType string, body io.Reader) {
	req, err := http.NewRequest(meth, url, body)
	if err != nil {
		log.Fatal(err)
	}

	req.SetBasicAuth(getCreds(req.URL))
	req.Header.Add("User-Agent", "hk/"+Version)
	req.Header.Add("Accept", "application/json")
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	fmt.Fprintf(os.Stderr, "%#v\n", req)
	fmt.Fprintf(os.Stderr, "%#v\n", req.URL)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode == 401 {
		log.Fatal("Unauthorized")
	}
	if res.StatusCode == 403 {
		log.Fatal("Unauthorized")
	}
	if res.StatusCode/100 != 2 { // 200, 201, 202, etc
		log.Fatal("Unexpected error: ", res.Status)
	}

	if msg := res.Header.Get("X-Heroku-Warning"); msg != "" {
		fmt.Fprintln(os.Stderr, strings.TrimSpace(msg))
	}

	if v != nil {
		err = json.NewDecoder(res.Body).Decode(v)
		if err != nil {
			log.Fatal(err)
		}
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
