package main

import (
	"errors"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/heroku/hk/Godeps/_workspace/src/github.com/bgentry/heroku-go"
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

Examples:

    $ hk scale web=2
    Scaled myapp to web=2:1X.

    $ hk scale web=2:1X worker=5:2X
    Scaled myapp to web=2:1X, worker=5:2X.

    $ hk scale web=PX worker=1X
    Scaled myapp to web=2:PX, worker=5:1X.
`,
}

// takes args of the form "web=1", "worker=3X", web=4:2X etc
func runScale(cmd *Command, args []string) {
	appname := mustApp()
	if len(args) == 0 {
		cmd.PrintUsage()
		os.Exit(2)
	}
	todo := make([]heroku.FormationBatchUpdateOpts, len(args))
	types := make(map[string]bool)
	for i, arg := range args {
		pstype, qty, size, err := parseScaleArg(arg)
		if err != nil {
			cmd.PrintUsage()
			os.Exit(2)
		}
		if _, exists := types[pstype]; exists {
			// can only specify each process type once
			printError("process type '%s' specified more than once", pstype)
			cmd.PrintUsage()
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

	formations, err := client.FormationBatchUpdate(appname, todo)
	must(err)

	sortedFormations := formationsByType(formations)
	sort.Sort(sortedFormations)
	results := make([]string, len(types))
	rindex := 0
	for _, f := range sortedFormations {
		if _, exists := types[f.Type]; exists {
			results[rindex] = f.Type + "=" + strconv.Itoa(f.Quantity) + ":" + f.Size
			rindex += 1
		}
	}
	log.Printf("Scaled %s to %s.", appname, strings.Join(results, ", "))
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

type formationsByType []heroku.Formation

func (f formationsByType) Len() int           { return len(f) }
func (f formationsByType) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }
func (f formationsByType) Less(i, j int) bool { return f[i].Type < f[j].Type }
