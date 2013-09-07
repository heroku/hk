package main

import (
	"log"
	"os"
	"os/exec"
)

var cmdDestroy = &Command{
	Run:   runDestroy,
	Usage: "destroy app",
	Short: "destroy an app",
	Long: `
Destroy destroys a heroku app.

There is no going back, so be sure you mean it.

Example:

    $ hk destroy myapp
`,
}

func runDestroy(cmd *Command, args []string) {
	if len(args) != 1 {
		cmd.printUsage()
		os.Exit(2)
	}
	name := args[0]
	var dynos []*Dyno
	err := Get(&dynos, "/apps/"+name+"/dynos")
	if err == nil && len(dynos) > 0 {
		log.Println("App", name, "has", len(dynos), "running dynos")
		log.Println("To destroy this app, first scale to zero")
		log.Println("To list dynos: hk ps")
		log.Println("For more on scaling: hk help scale")
		os.Exit(1)
	}
	if res := getMergedAddons(name); len(res) > 0 {
		log.Println("App", name, "has", len(res), "addons")
		log.Println("To destroy this app, first remove its addons")
		log.Println("To list them: hk ls addons")
		// TODO(kr): uncomment this when the command is known
		//log.Println("For more on removing addons: hk help XXXXX")
		os.Exit(1)
	}
	must(Delete("/apps/" + args[0]))
	for _, remote := range gitRemotes(gitURL(args[0])) {
		exec.Command("git", "remote", "rm", remote).Run()
	}
}
