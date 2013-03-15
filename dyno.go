package main

import (
	"fmt"
	"log"
	"net/url"
	"sort"
	"strings"
)

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

var cmdPs = &Command{
	Run:   runPs,
	Usage: "ps [-a app]",
	Short: "list running dynos",
	Long:  `List app's running dynos.`,
}

func runPs(cmd *Command, args []string) {
	var procs Procs
	must(APIReq(&v2{&procs}, "GET", "/apps/"+mustApp()+"/ps", nil))
	sort.Sort(procs)
	fmt.Printf("Process           State       Command\n")
	fmt.Printf("----------------  ----------  ------------------------\n")
	for _, proc := range procs {
		fmt.Printf("%-16s  %-10s  %s\n", proc.Name, proc.State, proc.Command)
	}
}

var cmdRestart = &Command{
	Run:   runRestart,
	Usage: "restart [-a app] [type or name]",
	Short: "restart dynos",
	Long: `
Restart all app dynos, all dynos of a specific type, or a single dyno.

Examples:

  $ hk restart
  $ hk restart web
  $ hk restart web.1
`,
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

	must(APIReq(nil, "POST", "/apps/"+mustApp()+"/ps/restart", v))
}
