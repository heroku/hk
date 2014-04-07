package main

import (
	"os"
	"text/tabwriter"
)

var cmdRegions = &Command{
	Run:      runRegions,
	Usage:    "regions",
	Category: "misc",
	Short:    "list regions" + extra,
	Long: `
Lists regions. Shows the region name and description.

Examples:

    $ hk regions
    eu  Europe
    us  United States
`,
}

func runRegions(cmd *Command, args []string) {
	if len(args) != 0 {
		cmd.PrintUsage()
		os.Exit(2)
	}
	regions, err := client.RegionList(nil)
	must(err)

	w := tabwriter.NewWriter(os.Stdout, 1, 2, 2, ' ', 0)
	defer w.Flush()

	for _, r := range regions {
		listRec(w,
			r.Name,
			r.Description,
		)
	}
}
