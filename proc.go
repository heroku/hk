package main

import (
	"fmt"
	"log"
	"os"
	"sort"
)

var cmdPs = &Command{
	Run:   runPs,
	Usage: "ps",
	Short: "list processes",
	Long:  `List app processes.`,
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
	if (len(os.Args) != 4) || (os.Args[2] != "-a") {
		log.Fatal("Invalid usage. See 'hk help ps'")
	}
	appName := os.Args[3]
	var procs Procs
	APIReq("GET", "/apps/"+appName+"/ps").Do(&procs)
	sort.Sort(procs)
	fmt.Printf("Process           State       Command\n")
	fmt.Printf("----------------  ----------  ------------------------\n")
	for _, proc := range procs {
		fmt.Printf("%-16s  %-10s  %s\n", proc.Name, proc.State, proc.Command)
	}
}
