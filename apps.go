package main

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/bgentry/heroku-go"
)

var cmdApps = &Command{
	Run:      runApps,
	Usage:    "apps [<name>...]",
	Category: "app",
	Short:    "list apps",
	Long: `
Lists apps. Shows the app name, owner, and last release time (or
time the app was created, if it's never been released).

Examples:

    $ hk apps
    myapp   user@test.com         Jan 2 12:34
    myapp2  user@longdomainname…  Jan 2 12:34

    $ hk apps myapp
    myapp  user@test.com  Jan 2 12:34
`,
}

func runApps(cmd *Command, names []string) {
	w := tabwriter.NewWriter(os.Stdout, 1, 2, 2, ' ', 0)
	defer w.Flush()
	var apps []heroku.App
	if len(names) == 0 {
		var err error
		apps, err = client.AppList(&heroku.ListRange{Field: "name", Max: 1000})
		must(err)
	} else {
		appch := make(chan *heroku.App, len(names))
		errch := make(chan error, len(names))
		for _, name := range names {
			if name == "" {
				appch <- nil
			} else {
				go func(appname string) {
					if app, err := client.AppInfo(appname); err != nil {
						errch <- err
					} else {
						appch <- app
					}
				}(name)
			}
		}
		for _ = range names {
			select {
			case err := <-errch:
				fmt.Fprintln(os.Stderr, err)
			case app := <-appch:
				if app != nil {
					apps = append(apps, *app)
				}
			}
		}
	}
	printAppList(w, apps)
}

func printAppList(w io.Writer, apps []heroku.App) {
	sort.Sort(appsByName(apps))
	abbrevEmailApps(apps)
	for _, a := range apps {
		if a.Name != "" {
			listApp(w, a)
		}
	}
}

func abbrevEmailApps(apps []heroku.App) {
	domains := make(map[string]int)
	for _, a := range apps {
		parts := strings.SplitN(a.Owner.Email, "@", 2)
		if len(parts) == 2 {
			domains["@"+parts[1]]++
		}
	}
	smax, nmax := "", 0
	for s, n := range domains {
		if n > nmax {
			smax = s
			nmax = n
		}
	}
	for _, a := range apps {
		if strings.HasSuffix(a.Owner.Email, smax) {
			a.Owner.Email = a.Owner.Email[:len(a.Owner.Email)-len(smax)]
		}
	}
}

func listApp(w io.Writer, a heroku.App) {
	t := a.CreatedAt
	if a.ReleasedAt != nil {
		t = *a.ReleasedAt
	}
	listRec(w,
		a.Name,
		abbrev(a.Owner.Email, 20),
		prettyTime{t},
	)
}

type appsByName []heroku.App

func (a appsByName) Len() int           { return len(a) }
func (a appsByName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a appsByName) Less(i, j int) bool { return a[i].Name < a[j].Name }
