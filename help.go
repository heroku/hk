package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"text/template"
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

  When set to disable, hk will insecurly skip SSL verification.

HKHEADERS

  A NL-separated list of headers to add to each API request.

HKPATH

  A list of directories to search for plugins. This variable takes
  the same form as the system PATH var. If unset, the value is
  taken to be "/usr/local/lib/hk/plugin" on Unix.

  See 'hk help plugins' for information about the plugin interface.
`,
}

var cmdFetchUpdate = &Command{
	Run:   runFetchUpdate,
	Usage: "fetch-update",
	Long:  `Downloads the next version of hk for later installation.`,
}

func runFetchUpdate(cmd *Command, args []string) {
	updater.fetchAndApply()
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

	fmt.Fprintf(os.Stderr, "Unknown help topic: %q. Run 'hk help'.\n", args[0])
	os.Exit(2)
}

var usageTemplate = template.Must(template.New("usage").Parse(`Usage: hk [command] [options] [arguments]

Supported commands are:
{{range .Commands}}{{if .Runnable}}{{if .ShowUsage}}
  {{.Name | printf "%-8s"}} {{.Short}}{{end}}{{end}}{{end}}
{{range .Plugins}}
  {{. | printf "%-8s"}} (plugin){{end}}

See 'hk help [command]' for more information about a command.

Additional help topics:
{{range .Commands}}{{if not .Runnable}}
  {{.Name | printf "%-8s"}} {{.Short}}{{end}}{{end}}

See 'hk help [topic]' for more information about that topic.

`))

func printUsage() {
	var plugins []string
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
			plugins = append(plugins, name)
		}
	}

	usageTemplate.Execute(os.Stdout, struct {
		Commands []*Command
		Plugins  []string
	}{
		commands,
		plugins,
	})
}

func usage() {
	printUsage()
	os.Exit(2)
}
