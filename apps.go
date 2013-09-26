package main

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
)

var cmdApps = &Command{
	Run:   runApps,
	Usage: "apps [-l] [app...]",
	Short: "list apps",
	Long: `
Lists all apps.

Options:

    -l       long listing

Long listing for shows the owner, slug size, last release time
(or time the app was created, if it's never been released), and
the app name.

Examples:

	$ hk apps
	myapp
	myapp2

	$ hk apps myapp
	myapp

	$ hk apps -l
	app  me  1234k  Jan 2 12:34  myapp
	app  me  4567k  Jan 2 12:34  myapp2
`,
}

func init() {
	cmdApps.Flag.BoolVar(&flagLong, "l", false, "long listing")
}

func runApps(cmd *Command, names []string) {
	w := tabwriter.NewWriter(os.Stdout, 1, 2, 2, ' ', 0)
	var apps []*App
	if len(names) == 0 {
		must(Get(&apps, "/apps"))
	} else {
		ch := make(chan error, len(names))
		for _, name := range names {
			if name == "" {
				ch <- nil
			} else {
				v, url := new(App), "/apps/"+name
				apps = append(apps, v)
				go func() { ch <- Get(v, url) }()
			}
		}
		for _ = range names {
			if err := <-ch; err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
		}
	}
	printAppList(w, apps)
	w.Flush()
}

func printAppList(w io.Writer, apps []*App) {
	sort.Sort(appsByName(apps))
	abbrevEmailApps(apps)
	for _, a := range apps {
		if a.Name != "" {
			listApp(w, a)
		}
	}
}

func abbrevEmailApps(apps []*App) {
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

func listApp(w io.Writer, a *App) {
	if flagLong {
		size := 0
		if a.SlugSize != nil {
			size = *a.SlugSize
		}
		t := a.CreatedAt
		if a.ReleasedAt != nil {
			t = *a.ReleasedAt
		}
		listRec(w,
			"app",
			abbrev(a.Owner.Email, 10),
			fmt.Sprintf("%6dk", (size+501)/(1000)),
			prettyTime{t},
			a.Name,
		)
	} else {
		fmt.Fprintln(w, a.Name)
	}
}

type appsByName []*App

func (a appsByName) Len() int           { return len(a) }
func (a appsByName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a appsByName) Less(i, j int) bool { return a[i].Name < a[j].Name }
