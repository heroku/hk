package main

import (
	"fmt"
	"sort"
)

var cmdPs = &Command{
	Run:   runPs,
	Usage: "ps [-a app]",
	Short: "list processes",
	Long:  `List app processes.`,
}

var cmdRestart = &Command{
	Run:   runRestart,
	Usage: "restart [-a app]",
	Short: "restart processes",
	Long:  `Restart app processes.`,
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
	APIReq("POST", "/apps/"+mustApp()+"/ps/restart").Do(nil)
}
