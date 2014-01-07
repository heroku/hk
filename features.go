package main

import (
	"io"
	"log"
	"os"
	"text/tabwriter"

	"github.com/bgentry/heroku-go"
)

var cmdFeatures = &Command{
	Run:      runFeatures,
	Usage:    "features",
	NeedsApp: true,
	Category: "app",
	Short:    "list app features",
	Long: `
Features lists Heroku Labs features for an app.

Example:

  $ hk features
  +  preboot
     user-env-compile
  +  websockets
`,
}

func runFeatures(cmd *Command, args []string) {
	if len(args) != 0 {
		cmd.printUsage()
		os.Exit(2)
	}
	w := tabwriter.NewWriter(os.Stdout, 1, 2, 2, ' ', 0)
	defer w.Flush()

	features, err := client.AppFeatureList(mustApp(), &heroku.ListRange{Field: "name"})
	must(err)

	listFeatures(w, features)
}

func listFeatures(w io.Writer, features []heroku.AppFeature) {
	for _, f := range features {
		enabled := " "
		if f.Enabled {
			enabled = "+"
		}
		listRec(w,
			enabled,
			f.Name,
		)
	}
}

var cmdFeatureEnable = &Command{
	Run:      runFeatureEnable,
	Usage:    "feature-enable <feature>",
	NeedsApp: true,
	Category: "app",
	Short:    "enable an app feature",
	Long: `
Enables a Heroku Labs feature on an app.

Example:

  $ hk feature-enable preboot
  Enabled preboot on myapp.
`,
}

func runFeatureEnable(cmd *Command, args []string) {
	if len(args) != 1 {
		cmd.printUsage()
		os.Exit(2)
	}
	appname := mustApp()
	featureName := args[0]
	feature, err := client.AppFeatureUpdate(appname, featureName, true)
	must(err)
	log.Printf("Enabled %s on %s.", feature.Name, appname)
}

var cmdFeatureDisable = &Command{
	Run:      runFeatureDisable,
	Usage:    "feature-disable <feature>",
	NeedsApp: true,
	Category: "app",
	Short:    "disable an app feature",
	Long: `
Disables a Heroku Labs feature on an app.

Example:

  $ hk feature-disable websockets
  Disabled websockets on myapp.
`,
}

func runFeatureDisable(cmd *Command, args []string) {
	if len(args) != 1 {
		cmd.printUsage()
		os.Exit(2)
	}
	appname := mustApp()
	featureName := args[0]
	feature, err := client.AppFeatureUpdate(appname, featureName, false)
	must(err)
	log.Printf("Disabled %s on %s.", feature.Name, appname)
}
