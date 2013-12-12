package main

import (
	"errors"
	"github.com/bgentry/heroku-go"
	"os"
	"strconv"
	"strings"
)

var cmdScale = &Command{
	Run:      runScale,
	Usage:    "scale <type>=[<qty>]:[<size>]...",
	NeedsApp: true,
	Category: "dyno",
	Short:    "change dyno quantities and sizes",
	Long: `
Scale changes the quantity of dynos (horizontal scale) and/or the
dyno size (vertical scale) for each process type. Note that
changing dyno size will restart all dynos of that type.

Example:

	$ hk scale web=2 worker=5

	$ hk scale web=2:1X worker=5:2X

	$ hk scale web=2X worker=1X
`,
}

// takes args of the form "web=1", "worker=3X", web=4:2X etc
func runScale(cmd *Command, args []string) {
	todo := make([]heroku.FormationBatchUpdateOpts, len(args))
	types := make(map[string]bool)
	for i, arg := range args {
		pstype, qty, size, err := parseScaleArg(arg)
		if err != nil {
			cmd.printUsage()
			os.Exit(2)
		}
		if _, exists := types[pstype]; exists {
			// can only specify each process type once
			cmd.printUsage()
			os.Exit(2)
		}
		types[pstype] = true

		opt := heroku.FormationBatchUpdateOpts{Process: pstype}
		if qty != -1 {
			opt.Quantity = &qty
		}
		if size != "" {
			opt.Size = &size
		}
		todo[i] = opt
	}

	_, err := client.FormationBatchUpdate(mustApp(), todo)
	must(err)
}

var errInvalidScaleArg = errors.New("invalid argument")

func parseScaleArg(arg string) (pstype string, qty int, size string, err error) {
	qty = -1
	iEquals := strings.IndexRune(arg, '=')
	if fields := strings.Fields(arg); len(fields) > 1 || iEquals == -1 {
		err = errInvalidScaleArg
		return
	}
	pstype = arg[:iEquals]

	rem := strings.ToUpper(arg[iEquals+1:])
	if len(rem) == 0 {
		err = errInvalidScaleArg
		return
	}

	if iColon := strings.IndexRune(rem, ':'); iColon == -1 {
		if iX := strings.IndexRune(rem, 'X'); iX == -1 {
			qty, err = strconv.Atoi(rem)
			if err != nil {
				return pstype, -1, "", errInvalidScaleArg
			}
		} else {
			size = rem
		}
	} else {
		if iColon > 0 {
			qty, err = strconv.Atoi(rem[:iColon])
			if err != nil {
				return pstype, -1, "", errInvalidScaleArg
			}
		}
		if len(rem) > iColon+1 {
			size = rem[iColon+1:]
		}
	}
	if err != nil || qty == -1 && size == "" {
		err = errInvalidScaleArg
	}
	return
}
