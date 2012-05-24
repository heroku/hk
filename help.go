package main

import (
	"fmt"
	"os"
)

func fetchUpdateHelp() {
	cmdHelp("hk fetch-update", "Download the latest hk client")
}

func fetchUpdate() {
	if len(os.Args) != 2 {
		unrecArg(os.Args[2], "fetch-update")
	}

	updater.fetchAndApply()
}

func versionHelp() {
	cmdHelp("hk version", "Show hk client version")
}

func version() {
	if len(os.Args) != 2 {
		unrecArg(os.Args[2], "version")
	}
	fmt.Printf("%s\n", Version)
}

func help() {
	if len(os.Args) <= 2 {
		usage()
	} else {
		cmd := os.Args[2]
		switch cmd {
		case "env":
			envHelp()
		case "get":
			getHelp()
		case "info":
			infoHelp()
		case "creds":
			credsHelp()
		case "list":
			listHelp()
		case "ps":
			psHelp()
		case "fetch-update":
			fetchUpdateHelp()
		case "version":
			versionHelp()
		default:
			unrecCmd(cmd)
		}
	}
}

// top-level usage
func usage() {
	fmt.Printf("Usage: hk <command> [-a <app>] [command-specific-options]\n\n")
	fmt.Printf("Supported hk commands are:\n")
	fmt.Printf("  addons          List add-ons\n")
	fmt.Printf("  addons-add      Add an add-on\n")
	fmt.Printf("  addons-open     Open an add-on page\n")
	fmt.Printf("  addons-remove   Remove an add-on \n")
	fmt.Printf("  create          Create an app\n")
	fmt.Printf("  destroy         Destroy an app\n")
	fmt.Printf("  env             List config vars\n")
	fmt.Printf("  get             Get config var\n")
	fmt.Printf("  help            Show this help\n")
	fmt.Printf("  info            Show app info\n")
	fmt.Printf("  list            List apps\n")
	fmt.Printf("  login           Log in\n")
	fmt.Printf("  logout          Log out\n")
	fmt.Printf("  logs            Show logs\n")
	fmt.Printf("  open            Open app\n")
	fmt.Printf("  pg              List databases\n")
	fmt.Printf("  pg-info         Show database info\n")
	fmt.Printf("  pg-promote      Promote a database\n")
	fmt.Printf("  ps-psql         Open a psql database shell\n")
	fmt.Printf("  pg-wait         Await a database\n")
	fmt.Printf("  ps              List processes\n")
	fmt.Printf("  release         Show release info\n")
	fmt.Printf("  releases        List releases\n")
	fmt.Printf("  rename          Rename an app\n")
	fmt.Printf("  restart         Restart processes\n")
	fmt.Printf("  rollback        Rollback to a previous release\n")
	fmt.Printf("  run             Run a process\n")
	fmt.Printf("  set             Set config var\n")
	fmt.Printf("  scale           Scale processes\n")
	fmt.Printf("  stop            Stop a process\n")
	fmt.Printf("  creds           Show auth creds\n")
	fmt.Printf("  unset           Unset config vars\n")
	fmt.Printf("  version         Display version\n\n")
	fmt.Printf("See 'hk help <command>' for more information on a specific command.\n")
}

func cmdHelp(usage string, desc string) {
	fmt.Printf("Usage: %s\n\n", usage)
	fmt.Printf("%s.\n", desc)
}
