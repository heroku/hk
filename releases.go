package main

import (
	"fmt"
	"os"
)

var cmdRels = &Command{
	Run:   runRels,
	Usage: "rels [-a app] [release]",
	Short: "show releases and info",
	Long:  `Rels displays a list of releases and information for a single release.`,
}

func init() {
	cmdRels.Flag.StringVar(&flagApp, "a", "", "app")
}

func runRels(cmd *Command, args []string) {
	if len(args) == 0 {
		var releases []struct{ Name string }
		APIReq("GET", "/apps/"+mustApp()+"/releases").Do(&releases)
		for _, release := range releases {
			fmt.Printf("%s\n", release.Name)
		}
	} else if len(args) == 1 {
		var release struct {
			Name       string
			Descr      string
			User       string
			Commit     string
			Created_At string
		}
		APIReq("GET", "/apps/"+mustApp()+"/releases/"+args[0]).Do(&release)
		fmt.Printf("Name:     %s\n", release.Name)
		fmt.Printf("Desc:     %s\n", release.Descr)
		fmt.Printf("User:     %s\n", release.User)
		fmt.Printf("Commit:   %s\n", release.Commit)
		fmt.Printf("Created:  %s\n", release.Created_At)
		// Should we display addons, pstable, and env? - RM
	} else {
		cmd.printUsage()
		os.Exit(2)
	}
}
