package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/bgentry/go-netrc/netrc"
	"github.com/bgentry/heroku-go"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
)

var (
	apiURL = "https://api.heroku.com"
	stdin  = bufio.NewReader(os.Stdin)
)

func hkHome() string {
	return filepath.Join(homePath(), ".hk")
}

func homePath() string {
	u, err := user.Current()
	if err != nil {
		panic("couldn't determine user: " + err.Error())
	}
	return u.HomeDir
}

func netrcPath() string {
	if runtime.GOOS == "windows" {
		return filepath.Join(homePath(), "_netrc")
	}
	return filepath.Join(homePath(), ".netrc")
}

type Command struct {
	// args does not include the command name
	Run      func(cmd *Command, args []string)
	Flag     flag.FlagSet
	NeedsApp bool

	Usage    string // first word is the command name
	Category string // i.e. "App", "Account", etc.
	Short    string // `hk help` output
	Long     string // `hk help cmd` output
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

// Running `hk help` will list commands in this order.
var commands = []*Command{
	cmdCreate,
	cmdApps,
	cmdDynos,
	cmdReleases,
	cmdReleaseInfo,
	cmdRollback,
	cmdAddons,
	cmdAddonAdd,
	cmdAddonRemove,
	cmdScale,
	cmdRestart,
	cmdSet,
	cmdUnset,
	cmdEnv,
	cmdRun,
	cmdLog,
	cmdInfo,
	cmdRename,
	cmdDestroy,
	cmdDomains,
	cmdDomainAdd,
	cmdDomainRemove,
	cmdSSHKeyAdd,
	cmdVersion,
	cmdHelp,

	helpEnviron,
	helpPlugins,
	helpMore,
	helpAbout,

	// listed by hk help more
	cmdAddonOpen,
	cmdAPI,
	cmdGet,
	cmdCreds,
	cmdOpen,
	cmdTransfer,
	cmdTransfers,
	cmdTransferAccept,
	cmdTransferDecline,
	cmdTransferCancel,
	cmdURL,
	cmdWhichApp,

	// unlisted
	cmdUpdate,
}

var (
	flagApp   string
	client    heroku.Client
	hkAgent   = "hk/" + Version + " (" + runtime.GOOS + "; " + runtime.GOARCH + ")"
	userAgent = hkAgent + " " + heroku.DefaultUserAgent
)

func main() {
	if updater != nil {
		defer updater.backgroundRun() // doesn't run if os.Exit is called
	}
	log.SetFlags(0)

	args := os.Args[1:]

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

	for _, cmd := range commands {
		if cmd.Name() == args[0] && cmd.Run != nil {
			cmd.Flag.Usage = func() {
				cmd.printUsage()
			}
			if cmd.NeedsApp {
				cmd.Flag.StringVar(&flagApp, "a", "", "app name")
			}
			if err := cmd.Flag.Parse(args[1:]); err != nil {
				os.Exit(2)
			}
			if cmd.NeedsApp {
				if a, _ := app(); a == "" {
					log.Println("no app specified")
					cmd.printUsage()
					os.Exit(2)
				}
			}
			cmd.Run(cmd, cmd.Flag.Args())
			return
		}
	}

	path := findPlugin(args[0])
	if path == "" {
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", args[0])
		usage()
	}
	err := execPlugin(path, args)
	log.Fatal("exec error: ", err)
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

	m, err := netrc.FindMachine(netrcPath(), apiURL.Host)
	if err != nil {
		log.Fatalf("netrc error (%s): %v", apiURL.Host, err)
	}

	return m.Login, m.Password
}

func app() (string, error) {
	if flagApp != "" {
		return flagApp, nil
	}

	if app := os.Getenv("HKAPP"); app != "" {
		return app, nil
	}

	gitRemote := remoteFromGit()
	gitRemoteApp, err := appFromGitRemote(gitRemote)
	if err != nil {
		return "", err
	}

	return gitRemoteApp, nil
}

func remoteFromGit() string {
	b, err := exec.Command("git", "config", "heroku.remote").Output()
	if err != nil {
		return "heroku"
	}
	return strings.TrimSpace(string(b))
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

	out := strings.TrimSpace(string(b))

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

func openURL(url string) error {
	var command string
	var args []string
	switch runtime.GOOS {
	case "darwin":
		command = "open"
		args = []string{command, url}
	case "windows":
		command = "cmd"
		args = []string{"/c", "start " + url}
	default:
		if _, err := exec.LookPath("xdg-open"); err != nil {
			fmt.Println("xdg-open is required to open web pages on " + runtime.GOOS)
			os.Exit(2)
		}
		command = "xdg-open"
		args = []string{command, url}
	}
	if runtime.GOOS != "windows" {
		p, err := exec.LookPath(command)
		if err != nil {
			fmt.Printf("Error finding path to %q: %s\n", command, err)
			os.Exit(2)
		}
		command = p
	}
	return sysExec(command, args, os.Environ())
}
