package main

import (
	"io"
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

func runRegions(cmd *Command, names []string) {
	w := tabwriter.NewWriter(os.Stdout, 1, 2, 2, ' ', 0)
	defer w.Flush()

	if len(names) > 1 {
		cmd.printUsage()
		os.Exit(2)
	}
	listRegions(w, names)
}

func listRegions(w io.Writer, names []string) {
	regions, err := client.RegionList(nil)
	must(err)

	for _, d := range regions {
		listRec(w,
			d.Name,
			d.Description,
		)
	}
}
