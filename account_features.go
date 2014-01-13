package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"text/tabwriter"

	"github.com/bgentry/heroku-go"
)

var cmdAccountFeatures = &Command{
	Run:      runAccountFeatures,
	Usage:    "account-features",
	Category: "account",
	Short:    "list account features",
	Long: `
Account-features lists Heroku Labs features for your account.

Example:

    $ hk account-features
    +  pipelines
`,
}

func runAccountFeatures(cmd *Command, args []string) {
	if len(args) != 0 {
		cmd.printUsage()
		os.Exit(2)
	}
	w := tabwriter.NewWriter(os.Stdout, 1, 2, 2, ' ', 0)
	defer w.Flush()

	features, err := client.AccountFeatureList(&heroku.ListRange{Field: "name"})
	must(err)

	listAccountFeatures(w, features)
}

func listAccountFeatures(w io.Writer, features []heroku.AccountFeature) {
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

var cmdAccountFeatureInfo = &Command{
	Run:      runAccountFeatureInfo,
	Usage:    "account-feature-info <feature>",
	Category: "app",
	Short:    "show info for an account feature",
	Long: `
Shows detailed info for a Heroku Labs feature on an account.

Example:

    $ hk feature-info preboot
    ...
`,
}

func runAccountFeatureInfo(cmd *Command, args []string) {
	if len(args) != 1 {
		cmd.printUsage()
		os.Exit(2)
	}
	feature, err := client.AccountFeatureInfo(args[0])
	must(err)
	fmt.Printf("Name:         %s\n", feature.Name)
	fmt.Printf("Docs:         %s\n", feature.DocURL)
	fmt.Printf("Enabled:      %t\n", feature.Enabled)
	fmt.Printf("Description:  %s\n", feature.Description)
}

var cmdAccountFeatureEnable = &Command{
	Run:      runAccountFeatureEnable,
	Usage:    "account-feature-enable <feature>",
	Category: "account",
	Short:    "enable an account feature",
	Long: `
Enables a Heroku Labs feature on your account.

Example:

    $ hk account-feature-enable pipelines
    Enabled pipelines.
`,
}

func runAccountFeatureEnable(cmd *Command, args []string) {
	if len(args) != 1 {
		cmd.printUsage()
		os.Exit(2)
	}
	featureName := args[0]
	feature, err := client.AccountFeatureUpdate(featureName, true)
	must(err)
	log.Printf("Enabled %s.", feature.Name)
}

var cmdAccountFeatureDisable = &Command{
	Run:      runAccountFeatureDisable,
	Usage:    "account-feature-disable <feature>",
	Category: "account",
	Short:    "disable an account feature",
	Long: `
Disables a Heroku Labs feature on your account.

Example:

    $ hk account-feature-disable pipelines
    Disabled pipelines.
`,
}

func runAccountFeatureDisable(cmd *Command, args []string) {
	if len(args) != 1 {
		cmd.printUsage()
		os.Exit(2)
	}
	featureName := args[0]
	feature, err := client.AccountFeatureUpdate(featureName, false)
	must(err)
	log.Printf("Disabled %s.", feature.Name)
}
