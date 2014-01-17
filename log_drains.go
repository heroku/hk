package main

import (
	"fmt"
	"log"
	"os"
	"text/tabwriter"

	"github.com/bgentry/heroku-go"
)

var cmdLogDrains = &Command{
	Run:      runLogDrains,
	Usage:    "log-drains",
	NeedsApp: true,
	Category: "app",
	Short:    "list log drains" + extra,
	Long: `
Lists log drains on an app. Shows the drain's ID, as well as its
Add-on name (if it's from an Add-on) or its URL.

Example:

    $ hk log-drains
    6af8b744-c513-4217-9f7c-1234567890ab  logging-addon:jumbo
    7f89b6bb-08af-4343-b0b4-d0415dd81712  syslog://my.log.host
    23fcdb8a-3095-46f5-abc2-c5f293c54cf1  syslog://my.other.log.host
`,
}

func runLogDrains(cmd *Command, args []string) {
	if len(args) != 0 {
		cmd.printUsage()
		os.Exit(2)
	}
	appname := mustApp()

	// fetch app's addons concurrently in case we need to resolve addon names
	addonsch := make(chan []heroku.Addon, 1)
	errch := make(chan error, 1)
	go func(appname string) {
		if addons, err := client.AddonList(appname, nil); err != nil {
			errch <- err
		} else {
			addonsch <- addons
		}
	}(appname)

	drains, err := client.LogDrainList(appname, nil)
	must(err)

	hasAddonDrains := false
	merged := make([]*mergedLogDrain, len(drains))
	for i := range drains {
		if !hasAddonDrains && drains[i].Addon != nil {
			hasAddonDrains = true
		}
		merged[i] = &mergedLogDrain{drain: drains[i], hasAddon: drains[i].Addon != nil}
	}

	if hasAddonDrains {
		// resolve addon names, use those instead of URLs
		select {
		case _ = <-errch:
			// couldn't resolve addons, just move on
		case addons := <-addonsch:
			mergeDrainAddonInfo(merged, addons)
		}
	}

	w := tabwriter.NewWriter(os.Stdout, 1, 2, 2, ' ', 0)
	defer w.Flush()

	for _, m := range merged {
		listRec(w, m.drain.Id, m.addonNameOrURL())
	}
}

type mergedLogDrain struct {
	drain    heroku.LogDrain
	hasAddon bool
	addon    *heroku.Addon
}

func (m *mergedLogDrain) addonNameOrURL() string {
	switch {
	case m.hasAddon && m.addon != nil:
		return m.addon.Plan.Name
	case m.hasAddon:
		return "unknown"
	default:
		return m.drain.URL
	}
}

// merge addon info into log drains
func mergeDrainAddonInfo(merged []*mergedLogDrain, addons []heroku.Addon) {
	for i := range merged {
		if merged[i].hasAddon {
			for j := range addons {
				if merged[i].drain.Addon.Id == addons[j].Id {
					merged[i].addon = &addons[j]
					break
				}
			}
		}
	}
}

var cmdLogDrainInfo = &Command{
	Run:      runLogDrainInfo,
	Usage:    "log-drain-info <id or url>",
	NeedsApp: true,
	Category: "app",
	Short:    "show info for a log drain" + extra,
	Long: `
Shows detailed info for a log drain.

Example:

    $ hk log-drain-info syslog://my.other.log.host
    Id:     7f89b6bb-08af-4343-b0b4-d0415dd81712
    Token:  d.a9dc787f-e0a8-43f3-a2c8-1fbf937fd47c
    Addon:  none
    URL:    syslog://my.log.host

    $ hk log-drain-info 23fcdb8a-3095-46f5-abc2-c5f293c54cf1
    ...
`,
}

func runLogDrainInfo(cmd *Command, args []string) {
	if len(args) != 1 {
		cmd.printUsage()
		os.Exit(2)
	}
	appname := mustApp()
	drainIdOrURL := args[0]
	drain, err := client.LogDrainInfo(appname, drainIdOrURL)
	must(err)

	addonName := "none"
	if drain.Addon != nil {
		addon, err := client.AddonInfo(appname, drain.Addon.Id)
		if err != nil {
			addonName = "unknown"
		} else {
			addonName = addon.Name
		}
	}

	fmt.Printf("Id:     %s\n", drain.Id)
	fmt.Printf("Token:  %s\n", drain.Token)
	fmt.Printf("Addon:  %s\n", addonName)
	fmt.Printf("URL:    %s\n", drain.URL)
}

var cmdLogDrainAdd = &Command{
	Run:      runLogDrainAdd,
	Usage:    "log-drain-add <url>",
	NeedsApp: true,
	Category: "app",
	Short:    "add a log drain" + extra,
	Long: `
Adds a log drain to an app.

Example:

    $ hk log-drain-add syslog://my.log.host
    Added log drain to myapp.
`,
}

func runLogDrainAdd(cmd *Command, args []string) {
	if len(args) != 1 {
		cmd.printUsage()
		os.Exit(2)
	}

	url := args[0]
	_, err := client.LogDrainCreate(mustApp(), url)
	must(err)
	log.Printf("Added log drain to %s.", mustApp())
}

var cmdLogDrainRemove = &Command{
	Run:      runLogDrainRemove,
	Usage:    "log-drain-remove <id or url>",
	NeedsApp: true,
	Category: "app",
	Short:    "remove a log drain" + extra,
	Long: `
Removes a log drain from an app.

Example:

    $ hk log-drain-remove 7f89b6bb-08af-4343-b0b4-d0415dd81712
    Removed log drain from myapp.

    $ hk log-drain-remove syslog://my.log.host
    Removed log drain from myapp.
`,
}

func runLogDrainRemove(cmd *Command, args []string) {
	if len(args) != 1 {
		cmd.printUsage()
		os.Exit(2)
	}

	drainId := args[0]
	must(client.LogDrainDelete(mustApp(), drainId))
	log.Printf("Removed log drain from %s.", mustApp())
}
