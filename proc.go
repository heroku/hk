package main

import (
	"fmt"
	"log"
	"net/url"
	"sort"
	"strings"
)

var cmdPs = &Command{
	Run:   runPs,
	Usage: "ps [-a app]",
	Short: "list processes",
	Long:  `List app processes.`,
}

var cmdRestart = &Command{
	Run:   runRestart,
	Usage: "restart [-a app] [type or name]",
	Short: "restart processes",
	Long: `
Restart all app processes, all processes of a specific type, or a single process.

Examples:

  $ hk restart
  $ hk restart web
  $ hk restart web.1
`,
}

func init() {
	cmds := []*Command{cmdPs, cmdRestart}
	for _, c := range cmds {
		c.Flag.StringVar(&flagApp, "a", "", "app")
	}
}

type Procs []*struct {
	Name    string `json:"process"`
	State   string
	Command string
}

func (p Procs) Len() int           { return len(p) }
func (p Procs) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p Procs) Less(i, j int) bool { return p[i].Name < p[j].Name }

func runPs(cmd *Command, args []string) {
	var procs Procs
	APIReq("GET", "/apps/"+mustApp()+"/ps").Do(&procs)
	sort.Sort(procs)
	fmt.Printf("Process           State       Command\n")
	fmt.Printf("----------------  ----------  ------------------------\n")
	for _, proc := range procs {
		fmt.Printf("%-16s  %-10s  %s\n", proc.Name, proc.State, proc.Command)
	}
}

func runRestart(cmd *Command, args []string) {
	if len(args) > 1 {
		log.Fatal("Invalid usage. See 'hk help restart'")
	}

	v := make(url.Values)

	if len(args) == 1 {
		if strings.Index(args[0], ".") > 0 {
			v.Add("ps", args[0])
		} else {
			v.Add("type", args[0])
		}
	}

	req := APIReq("POST", "/apps/"+mustApp()+"/ps/restart")
	req.SetBodyForm(v)
	req.Do(nil)
}
