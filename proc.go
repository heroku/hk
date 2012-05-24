package main

import (
	"fmt"
	"os"
	"sort"
)

func psHelp() {
	cmdHelp("hk ps -a <app>", "List app processes")
}

type Proc struct {
	Name    string `json:"process"`
	State   string
	Command string
}

type Procs []*Proc

func (p Procs) Len() int           { return len(p) }
func (p Procs) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p Procs) Less(i, j int) bool { return p[i].Name < p[j].Name }

func ps() {
	if (len(os.Args) != 4) || (os.Args[2] != "-a") {
		error("Invalid usage. See 'hk help ps'")
	}
	appName := os.Args[3]
	var procs Procs
	apiReq(&procs, "GET", fmt.Sprintf(apiURL+"/apps/%s/ps", appName))
	sort.Sort(procs)
	fmt.Printf("Process           State       Command\n")
	fmt.Printf("----------------  ----------  ------------------------\n")
	for _, proc := range procs {
		fmt.Printf("%-16s  %-10s  %s\n", proc.Name, proc.State, proc.Command)
	}
}
