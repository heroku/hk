package main

import (
	"bufio"
	"code.google.com/p/go-netrc/netrc"
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/bgentry/heroku-go"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
)

var (
	apiURL    = "https://api.heroku.com"
	hkHome    = filepath.Join(homePath, ".hk")
	netrcPath = filepath.Join(os.Getenv("HOME"), ".netrc")
	stdin     = bufio.NewReader(os.Stdin)
)

type Namespace struct {
	Name  string // single-word namespace
	Short string // `hk help` output

	// Commands that live in this namespace
	Commands []*Command
}

func (n *Namespace) printUsage() {
	for i := range n.Commands {
		fmt.Printf("hk %s %s\n", n.Name, n.Commands[i].Usage)
	}
}

type Command struct {
	// args does not include the command name
	Run      func(cmd *Command, args []string)
	Flag     flag.FlagSet
	NeedsApp bool // whether command needs the app param

	Usage string // first word is the command name
	Short string // `hk help` output
	Long  string // `hk help cmd` output
}

func (c *Command) printUsage() {
	if c.Runnable() {
		fmt.Printf("Usage: hk %s\n\n", c.Usage)
	}
	fmt.Println(strings.Trim(c.Long, "\n"))
}

func (c *Command) Name() string {
	name := c.Usage
	i := strings.Index(name, " ")
	if i >= 0 {
		name = name[:i]
	}
	return name
}

func (c *Command) Runnable() bool {
	return c.Run != nil
}

const extra = " (extra)"

func (c *Command) List() bool {
	return c.Short != "" && !strings.HasSuffix(c.Short, extra)
}

func (c *Command) ListAsExtra() bool {
	return c.Short != "" && strings.HasSuffix(c.Short, extra)
}

func (c *Command) ShortExtra() string {
	return c.Short[:len(c.Short)-len(extra)]
}

var namespaces = []*Namespace{
	nsApp,
	nsDyno,
	nsEnv,
}

// Running `hk help` will list commands in this order.
var commands = []*Command{
	cmdReleases,
	cmdAddons,
	cmdLog,
	cmdOpen,
	cmdSSHAuth,
	cmdVersion,
	cmdHelp,

	helpEnviron,
	helpPlugins,
	helpMore,
	helpAbout,

	// listed by hk help more
	cmdAPI,
	cmdCreds,
	cmdURL,

	// unlisted
	cmdUpdate,
}

var (
	flagApp   string
	flagLong  bool
	client    heroku.Client
	hkAgent   = "hk/" + Version + " (" + runtime.GOOS + "; " + runtime.GOARCH + ")"
	userAgent = hkAgent + " " + heroku.DefaultUserAgent
)

func main() {
	if updater != nil {
		defer updater.backgroundRun() // doesn't run if os.Exit is called
	}
	log.SetFlags(0)

	args := parseAppFromArgs(os.Args[1:])

	if len(args) < 1 {
		usage()
	}

	apiURL = heroku.DefaultAPIURL
	if s := os.Getenv("HEROKU_API_URL"); s != "" {
		apiURL = s
	}
	user, pass := getCreds(apiURL)
	debug := os.Getenv("HKDEBUG") != ""
	client = heroku.Client{
		URL:       apiURL,
		Username:  user,
		Password:  pass,
		UserAgent: userAgent,
		Debug:     debug,
	}
	if os.Getenv("HEROKU_SSL_VERIFY") == "disable" {
		client.HTTP.Transport.(*http.Transport).TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}
	if s := os.Getenv("HEROKU_API_URL"); s != "" {
		client.URL = s
	}
	client.AdditionalHeaders = http.Header{}
	for _, h := range strings.Split(os.Getenv("HKHEADER"), "\n") {
		if i := strings.Index(h, ":"); i >= 0 {
			client.AdditionalHeaders.Set(
				strings.TrimSpace(h[:i]),
				strings.TrimSpace(h[i+1:]),
			)
		}
	}

	for _, ns := range namespaces {
		if ns.Name == args[0] {
			if len(args) >= 2 && runFromCmds(ns.Commands, args[1:]) {
				return
			}
			ns.printUsage()
			os.Exit(2)
		}
	}

	if runFromCmds(commands, args) {
		return
	}

	path := findPlugin(args[0])
	if path == "" {
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", args[0])
		usage()
	}
	err := execPlugin(path, args)
	log.Fatal("exec error: ", err)
}

// returns whether the command ran
func runFromCmds(commands []*Command, args []string) bool {
	for _, cmd := range commands {
		if cmd.Name() == args[0] && cmd.Run != nil {
			cmd.Flag.Usage = func() {
				cmd.printUsage()
			}
			if !cmd.NeedsApp && flagApp != "" {
				log.Fatalf("flag provided but not defined: -a")
			}
			if cmd.NeedsApp {
				if a, _ := app(); a == "" {
					log.Fatal("no app specified")
				}
			}
			if err := cmd.Flag.Parse(args[1:]); err != nil {
				os.Exit(2)
			}
			cmd.Run(cmd, cmd.Flag.Args())
			return true
		}
	}
	return false
}

func getCreds(u string) (user, pass string) {
	apiURL, err := url.Parse(u)
	if err != nil {
		log.Fatalf("invalid API URL: %s", err)
	}
	if apiURL.User != nil {
		pw, _ := apiURL.User.Password()
		return apiURL.User.Username(), pw
	}

	m, err := netrc.FindMachine(netrcPath, apiURL.Host)
	if err != nil {
		log.Fatalf("netrc error (%s): %v", apiURL.Host, err)
	}

	return m.Login, m.Password
}

func parseAppFromArgs(args []string) []string {
	if len(args) >= 2 {
		if i := stringsIndex(args[1:], "-a"); i != -1 {
			if len(args[1:]) < i+2 {
				log.Fatal("missing value for app param")
			}
			if strings.IndexRune(args[1:][i+1], '-') == 0 {
				log.Fatalf("invalid value for app param: %s", args[1:][i+1])
			}
			flagApp = args[i+2]
			args = append(args[:1], append(args[1:i+1], args[i+3:]...)...)

			if gitRemoteApp, err := appFromGitRemote(flagApp); err == nil {
				flagApp = gitRemoteApp
			}
		}
	}
	return args
}

func app() (string, error) {
	if flagApp != "" {
		return flagApp, nil
	}

	if app := os.Getenv("HKAPP"); app != "" {
		return app, nil
	}

	gitRemoteApp, err := appFromGitRemote("heroku")
	if err != nil {
		return "", err
	}

	return gitRemoteApp, nil
}

func appFromGitRemote(remote string) (string, error) {
	b, err := exec.Command("git", "config", "remote."+remote+".url").Output()
	if err != nil {
		if isNotFound(err) {
			wdir, _ := os.Getwd()
			return "", fmt.Errorf("could not find git remote "+remote+" in %s", wdir)
		}
		return "", err
	}

	out := strings.Trim(string(b), "\r\n ")

	if !strings.HasPrefix(out, gitURLPre) || !strings.HasSuffix(out, gitURLSuf) {
		return "", fmt.Errorf("could not find app name in " + remote + " git remote")
	}

	return out[len(gitURLPre) : len(out)-len(gitURLSuf)], nil
}

func isNotFound(err error) bool {
	if ee, ok := err.(*exec.ExitError); ok {
		if ws, ok := ee.ProcessState.Sys().(syscall.WaitStatus); ok {
			return ws.ExitStatus() == 1
		}
	}
	return false
}

func mustApp() string {
	name, err := app()
	if err != nil {
		log.Fatal(err)
	}
	return name
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func listRec(w io.Writer, a ...interface{}) {
	for i, x := range a {
		fmt.Fprint(w, x)
		if i+1 < len(a) {
			w.Write([]byte{'\t'})
		} else {
			w.Write([]byte{'\n'})
		}
	}
}

type prettyTime struct {
	time.Time
}

func (s prettyTime) String() string {
	if time.Now().Sub(s.Time) < 12*30*24*time.Hour {
		return s.Local().Format("Jan _2 15:04")
	}
	return s.Local().Format("Jan _2  2006")
}

func stringsIndex(a []string, val string) int {
	for i := range a {
		if a[i] == val {
			return i
		}
	}
	return -1
}
