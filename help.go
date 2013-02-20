package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"text/template"
	"time"
)

var helpEnviron = &Command{
	Usage: "environ",
	Short: "environment variables used by hk",
	Long: `
Several environment variables affect hk's behavior.

HEROKU_API_URL

  The base URL hk will use to make api requests in the format:
  https://[username][:password]@<host>[:port]/

  If username and password are present in the URL, they will
  override .netrc.

  Its default value is https://api.heroku.com/

HEROKU_SSL_VERIFY

  When set to disable, hk will insecurely skip SSL verification.

HKHEADERS

  A NL-separated list of headers to add to each API request.

HKPATH

  A list of directories to search for plugins. This variable takes
  the same form as the system PATH var. If unset, the value is
  taken to be "/usr/local/lib/hk/plugin" on Unix.

  See 'hk help plugins' for information about the plugin interface.
`,
}

var cmdUpdate = &Command{
	Run:   runUpdate,
	Usage: "update",
	Long: `
Update downloads and installs the next version of hk.

This command is unlisted, since users never have to run it directly.
`,
}

func runUpdate(cmd *Command, args []string) {
	if err := updater.update(); err != nil {
		log.Fatal(err)
	}
}

var cmdVersion = &Command{
	Run:   runVersion,
	Usage: "version",
	Short: "show hk version",
	Long:  `Version shows the hk client version string.`,
}

func runVersion(cmd *Command, args []string) {
	fmt.Println(Version)
}

var cmdHelp = &Command{
	Usage: "help [command]",
	Short: "show help",
	Long:  `Help shows usage for a command.`,
}

func init() {
	cmdHelp.Run = runHelp // break init loop
}

func runHelp(cmd *Command, args []string) {
	if len(args) == 0 {
		printUsage()
		return // not os.Exit(2); success
	}
	if len(args) != 1 {
		log.Fatal("too many arguments")
	}

	for _, cmd := range commands {
		if cmd.Name() == args[0] {
			cmd.printUsage()
			return
		}
	}

	if lookupPlugin(args[0]) != "" {
		_, _, long := pluginInfo(string(args[0]))
		fmt.Println(long)
		return
	}

	fmt.Fprintf(os.Stderr, "Unknown help topic: %q. Run 'hk help'.\n", args[0])
	os.Exit(2)
}

var usageTemplate = template.Must(template.New("usage").Parse(`Usage: hk [command] [options] [arguments]

Supported commands are:
{{range .Commands}}{{if .Runnable}}{{if .ShowUsage}}
  {{.Name | printf "%-8s"}} {{.Short}}{{end}}{{end}}{{end}}
{{range .Plugins}}
  {{.Name | printf "%-8s"}} {{.Short}} (plugin){{end}}

See 'hk help [command]' for more information about a command.

Additional help topics:
{{range .Commands}}{{if not .Runnable}}
  {{.Name | printf "%-8s"}} {{.Short}}{{end}}{{end}}

See 'hk help [topic]' for more information about that topic.

{{if .Dev}}This dev build of hk will expire at {{.Expiration}}
{{end}}`))

func printUsage() {
	var plugins []plugin
	for _, path := range strings.Split(hkPath, ":") {
		d, err := os.Open(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			log.Fatal(err)
		}
		names, err := d.Readdirnames(-1)
		if err != nil {
			log.Fatal(err)
		}
		for _, name := range names {
			plugins = append(plugins, plugin(name))
		}
	}

	usageTemplate.Execute(os.Stdout, struct {
		Commands   []*Command
		Plugins    []plugin
		Dev        bool
		Expiration time.Time
	}{
		commands,
		plugins,
		Version == "dev",
		hkExpiration(),
	})
}

func usage() {
	printUsage()
	os.Exit(2)
}
