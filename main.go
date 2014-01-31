package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"

	"github.com/bgentry/heroku-go"
	"github.com/heroku/hk/postgresql"
	"github.com/heroku/hk/term"
	"github.com/mgutz/ansi"
)

var (
	apiURL = "https://api.heroku.com"
	stdin  = bufio.NewReader(os.Stdin)
)

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
		fmt.Printf("Usage: hk %s\n\n", c.FullUsage())
	}
	fmt.Println(strings.Trim(c.Long, "\n"))
}

func (c *Command) FullUsage() string {
	if c.NeedsApp {
		return c.Name() + " [-a <app>]" + strings.TrimPrefix(c.Usage, c.Name())
	}
	return c.Usage
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
	cmdVersion,
	cmdHelp,

	helpEnviron,
	helpPlugins,
	helpMore,
	helpAbout,

	// listed by hk help more
	cmdAccess,
	cmdAccessAdd,
	cmdAccessRemove,
	cmdAccountFeatures,
	cmdAccountFeatureInfo,
	cmdAccountFeatureEnable,
	cmdAccountFeatureDisable,
	cmdAddonOpen,
	cmdAPI,
	cmdCreds,
	cmdDrains,
	cmdDrainInfo,
	cmdDrainAdd,
	cmdDrainRemove,
	cmdFeatures,
	cmdFeatureInfo,
	cmdFeatureEnable,
	cmdFeatureDisable,
	cmdGet,
	cmdKeys,
	cmdKeyAdd,
	cmdKeyRemove,
	cmdLogin,
	cmdLogout,
	cmdMaintenance,
	cmdMaintenanceEnable,
	cmdMaintenanceDisable,
	cmdOpen,
	cmdPgInfo,
	cmdPsql,
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
	client    *heroku.Client
	pgclient  *postgresql.Client
	hkAgent   = "hk/" + Version + " (" + runtime.GOOS + "; " + runtime.GOARCH + ")"
	userAgent = hkAgent + " " + heroku.DefaultUserAgent
)

func main() {
	log.SetFlags(0)

	// make sure command is specified, disallow global args
	args := os.Args[1:]
	if len(args) < 1 || strings.IndexRune(args[0], '-') == 0 {
		usage()
	}

	// Run the update command as early as possible to avoid the possibility of
	// installations being stranded without updates due to errors in other code
	if args[0] == cmdUpdate.Name() {
		cmdUpdate.Run(cmdUpdate, args)
		return
	} else if updater != nil {
		defer updater.backgroundRun() // doesn't run if os.Exit is called
	}

	if !term.IsTerminal(os.Stdout) {
		ansi.DisableColors(true)
	}

	initClients()

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
			if flagApp != "" {
				if gitRemoteApp, err := appFromGitRemote(flagApp); err == nil {
					flagApp = gitRemoteApp
				}
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
		if g := suggest(args[0]); len(g) > 0 {
			fmt.Fprintf(os.Stderr, "Possible alternatives: %v\n", strings.Join(g, " "))
		}
		fmt.Fprintf(os.Stderr, "Run 'hk help' for usage.\n")
		os.Exit(2)
	}
	err := execPlugin(path, args)
	printError("exec error: %s", err)
}

func initClients() {
	disableSSLVerify := false
	apiURL = heroku.DefaultAPIURL
	if s := os.Getenv("HEROKU_API_URL"); s != "" {
		apiURL = s
		disableSSLVerify = true
	}
	user, pass := getCreds(apiURL)
	debug := os.Getenv("HKDEBUG") != ""
	client = &heroku.Client{
		URL:       apiURL,
		Username:  user,
		Password:  pass,
		UserAgent: userAgent,
		Debug:     debug,
	}
	pgclient = &postgresql.Client{
		Username:  user,
		Password:  pass,
		UserAgent: userAgent,
		Debug:     debug,
	}
	if disableSSLVerify || os.Getenv("HEROKU_SSL_VERIFY") == "disable" {
		client.HTTP = &http.Client{Transport: http.DefaultTransport}
		client.HTTP.Transport.(*http.Transport).TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
		pgclient.HTTP = client.HTTP
	}
	if s := os.Getenv("HEROKU_POSTGRESQL_HOST"); s != "" {
		pgclient.URL = s
	}
	client.AdditionalHeaders = http.Header{}
	pgclient.AdditionalHeaders = http.Header{}
	for _, h := range strings.Split(os.Getenv("HKHEADER"), "\n") {
		if i := strings.Index(h, ":"); i >= 0 {
			client.AdditionalHeaders.Set(
				strings.TrimSpace(h[:i]),
				strings.TrimSpace(h[i+1:]),
			)
			pgclient.AdditionalHeaders.Set(
				strings.TrimSpace(h[:i]),
				strings.TrimSpace(h[i+1:]),
			)
		}
	}
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
		printError(err.Error())
	}
	return name
}
