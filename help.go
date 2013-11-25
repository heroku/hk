package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
)

var helpEnviron = &Command{
	Usage:    "environ",
	Category: "hk",
	Short:    "environment variables used by hk",
	Long: `
Several environment variables affect hk's behavior.

HEROKU_API_URL

  The base URL hk will use to make api requests in the format:
  https://[username][:password]@host[:port]/

  If username and password are present in the URL, they will
  override .netrc.

  Its default value is https://api.heroku.com/

HEROKU_SSL_VERIFY

  When set to disable, hk will insecurely skip SSL verification.

HKHEADER

  A NL-separated list of fields to set in each API request header.
  These override any fields set by hk if they have the same name.

HKPATH

  A list of directories to search for plugins. This variable takes
  the same form as the system PATH var. If unset, the value is
  taken to be "/usr/local/lib/hk/plugin" on Unix.

  See 'hk help plugins' for information about the plugin interface.

HKDEBUG

  When this is set, hk prints the wire representation of each API
  request to stderr just before sending the request, and prints the
  response. This will most likely include your secret API key in
  the Authorization header field, so be careful with the output.
`,
}

var cmdVersion = &Command{
	Run:      runVersion,
	Usage:    "version",
	Category: "hk",
	Short:    "show hk version",
	Long:     `Version shows the hk client version string.`,
}

func runVersion(cmd *Command, args []string) {
	fmt.Println(Version)
}

var cmdHelp = &Command{
	Usage:    "help [<topic>]",
	Category: "hk",
	Long:     `Help shows usage for a command or other topic.`,
}

var helpMore = &Command{
	Usage:    "more",
	Category: "hk",
	Short:    "additional commands, less frequently used",
	Long:     "(not displayed; see special case in runHelp)",
}

var helpCommands = &Command{
	Usage:    "commands",
	Category: "hk",
	Short:    "list all commands with usage",
	Long:     "(not displayed; see special case in runHelp)",
}

var helpStyleGuide = &Command{
	Usage:    "styleguide",
	Category: "hk",
	Short:    "generate an html styleguide for all commands with usage",
	Long:     "(not displayed; see special case in runHelp)",
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
	switch args[0] {
	case helpMore.Name():
		printExtra()
		return
	case helpCommands.Name():
		printAllUsage()
		return
	case helpStyleGuide.Name():
		printStyleGuide()
		return
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

func maxStrLen(strs []string) (strlen int) {
	for i := range strs {
		if len(strs[i]) > strlen {
			strlen = len(strs[i])
		}
	}
	return
}

var usageTemplate = template.Must(template.New("usage").Parse(`
Usage: hk <command> [-a app] [options] [arguments]


Commands:
{{range .Commands}}{{if .Runnable}}{{if .List}}
    {{.Name | printf (print "%-" $.MaxRunListName "s")}}  {{.Short}}{{end}}{{end}}{{end}}
{{range .Plugins}}
    {{.Name | printf (print "%-" $.MaxRunListName "s")}}  {{.Short}} (plugin){{end}}

Run 'hk help [command]' for details.


Additional help topics:
{{range .Commands}}{{if not .Runnable}}
    {{.Name | printf "%-8s"}}  {{.Short}}{{end}}{{end}}

{{if .Dev}}This dev build of hk cannot auto-update itself.
{{end}}`[1:]))

var extraTemplate = template.Must(template.New("usage").Parse(`
Additional commands:
{{range .Commands}}{{if .Runnable}}{{if .ListAsExtra}}
    {{.Name | printf (print "%-" $.MaxRunExtraName "s")}}  {{.ShortExtra}}{{end}}{{end}}{{end}}

Run 'hk help [command]' for details.

`[1:]))

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
		fi, err := d.Readdir(-1)
		if err != nil {
			log.Fatal(err)
		}
		for _, f := range fi {
			if !f.IsDir() && f.Mode()&0111 != 0 {
				plugins = append(plugins, plugin(f.Name()))
			}
		}
	}

	var runListNames []string
	for i := range commands {
		if commands[i].Runnable() && commands[i].List() {
			runListNames = append(runListNames, commands[i].Name())
		}
	}
	for i := range plugins {
		runListNames = append(runListNames, plugins[i].Name())
	}

	usageTemplate.Execute(os.Stdout, struct {
		Commands       []*Command
		Plugins        []plugin
		Dev            bool
		MaxRunListName int
	}{
		commands,
		plugins,
		Version == "dev",
		maxStrLen(runListNames),
	})
}

func printExtra() {
	var runExtraNames []string
	for i := range commands {
		if commands[i].Runnable() && commands[i].ListAsExtra() {
			runExtraNames = append(runExtraNames, commands[i].Name())
		}
	}

	extraTemplate.Execute(os.Stdout, struct {
		Commands        []*Command
		MaxRunExtraName int
	}{
		commands,
		maxStrLen(runExtraNames),
	})
}

func printAllUsage() {
	w := tabwriter.NewWriter(os.Stdout, 1, 2, 2, ' ', 0)
	defer w.Flush()
	cl := commandList(commands)
	sort.Sort(cl)
	for i := range cl {
		if cl[i].Runnable() {
			listRec(w, "hk "+cl[i].Usage, "# "+cl[i].Short)
		}
	}
}

func printStyleGuide() {
	cmap := make(map[string]commandList)
	// group by category
	for i := range commands {
		if _, exists := cmap[commands[i].Category]; !exists {
			cmap[commands[i].Category] = commandList{commands[i]}
		} else {
			cmap[commands[i].Category] = append(cmap[commands[i].Category], commands[i])
		}
	}
	// sort each category
	for _, cl := range cmap {
		sort.Sort(cl)
	}
	err := styleGuideTemplate.Execute(os.Stdout, struct {
		CommandMap commandMap
	}{
		cmap,
	})
	if err != nil {
		log.Fatal(err)
	}
}

func (c *Command) UsageJSON() commandJSON {
	return commandJSON{Root: c.Name(), Arguments: strings.TrimLeft(c.Usage, c.Name()+" "), Comment: c.Short}
}

type commandJSON struct {
	Root      string `json:"root"`
	Arguments string `json:"arguments"`
	Comment   string `json:"comment"`
}

type commandList []*Command

func (cl commandList) Len() int           { return len(cl) }
func (cl commandList) Swap(i, j int)      { cl[i], cl[j] = cl[j], cl[i] }
func (cl commandList) Less(i, j int) bool { return cl[i].Name() < cl[j].Name() }

func (cl commandList) UsageJSON() []commandJSON {
	a := make([]commandJSON, len(cl))
	for i := range cl {
		a[i] = cl[i].UsageJSON()
	}
	return a
}

type commandMap map[string]commandList

func (cm commandMap) UsageJSON(prefix string) template.JS {
	a := make([]map[string]interface{}, 0)
	for k, cl := range cm {
		m := map[string]interface{}{"title": k, "commands": cl.UsageJSON()}
		a = append(a, m)
	}
	buf, err := json.MarshalIndent(a, prefix, "  ")
	if err != nil {
		return template.JS(fmt.Sprintf("{\"error\": %q}", err.Error))
	}
	resp := strings.Replace(string(buf), "\\u003c", "<", -1)
	resp = strings.Replace(resp, "\\u003e", ">", -1)
	return template.JS(resp)
}

func usage() {
	printUsage()
	os.Exit(2)
}

var styleGuideTemplate = template.Must(template.New("styleguide").Delims("{{{", "}}}").Parse(`
<!DOCTYPE html>
<html>
  <head>
    <title>hk style guide</title>

    <style>
      body {
        background: #282A36;
        color: white;
        font-family: Helvetica;
      }

      #viewing-options {
        padding: 0;
      }

      #viewing-options li {
        display: inline-block;
        margin-right: 20px;
      }

      td {
        font-family: monospace;
        padding-right: 10px;
      }

      td:first-child {
        width: 460px;
      }

      h2 {
        color: #5A5D6E;
      }

      .prompt,
      .comment {
        color: #6272A4;
      }

      .command {
        color: white;
      }

      .root {
        color: #FF79C6;
        font-weight: bold;
      }

      .arguments {
        color: #66D9D0;
      }
    </style>
  </head>

  <body>
    <script id="command-structure" type="text/x-handlebars-template">
      {{#groups}}
      <h2>{{title}}</h2>

      <table>
        {{#commands}}
        <tr>
          <td>
            <span class='prompt'>$</span>
            <span class='command'>hk</span>
            <span class='root'>{{root}}</span>
            <span class='arguments'>{{arguments}}</span>
          </td>
          <td class='comment'># {{comment}}</td>
        </tr>
        {{/commands}}
      </table>
      {{/groups}}
    </script>

    <script src="https://ajax.googleapis.com/ajax/libs/jquery/1.10.2/jquery.min.js"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/handlebars.js/1.1.2/handlebars.min.js"></script>

    <script>
      var source = $('#command-structure').html();
      var template = Handlebars.compile(source);

      var data = {{{.CommandMap.UsageJSON "      "}}}

      $('body').append(template({groups: data}));
    </script>
  </body>
</html>
`[1:]))
