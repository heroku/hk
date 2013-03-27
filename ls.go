package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"time"
)

var cmdLs = &Command{
	Run:   runLs,
	Usage: "ls [-l] [app...]",
	Short: "list apps, addons, dynos, and releases",
	Long: `
       hk ls [-l] [-a app] releases [name...]

       hk ls [-l] [-a app] addons [name...]

Command hk ls lists apps, releases, and addons.

Options:

    -l       long listing
    -a=name  app name

Long listing for apps shows the owner, slug size, last release
time (or time the app was created, if it's never been released),
and the app name. Long listing for releases shows the git commit
id, who made the release, time of the release, name of the release
(e.g. v1), and description. Long listing for addons shows the type
of the addon, owner, name of the resource, and the config var it's
attached to.

Examples:

    $ hk ls
    myapp
    myapp2

    $ hk ls -l
    app  me  1234k  Jan 2 12:34  myapp
    app  me  4567k  Jan 2 12:34  myapp2

    $ hk ls rel
    v1
    v2

    $ hk ls -l rel
    3ae20c2  me  Jun 12 18:28  v1  Deploy 3ae20c2
    0fda0ae  me  Jun 13 18:14  v2  Deploy 0fda0ae
    ed39b69  me  Jun 13 18:31  v3  Deploy ed39b69

    $ hk ls -l rel v3
    ed39b69  me  Jun 13 18:31  v3  Deploy ed39b69

    $ hk ls addons
    DATABASE_URL
    REDIS_URL

    $ hk ls -l addons REDIS_URL
    redistogo:nano  me  soaring-ably-1234  REDIS_URL
`,
}

func init() {
	cmdLs.Flag.StringVar(&flagApp, "a", "", "app")
	cmdLs.Flag.BoolVar(&flagLong, "l", false, "long listing")
}

func runLs(cmd *Command, args []string) {
	w := tabwriter.NewWriter(os.Stdout, 1, 2, 2, ' ', 0)
	list(w, cmd, args)
	w.Flush()
}

func list(w io.Writer, cmd *Command, args []string) {
	if len(args) == 0 {
		var apps []*App
		must(Get(&apps, "/apps"))
		sort.Sort(appsByName(apps))
		abbrevEmailApps(apps)
		for _, a := range apps {
			listApp(w, a)
		}
		return
	}
	switch a0 := args[0]; {
	case strings.HasPrefix("releases", a0):
		listRels(w, args[1:])
	case strings.HasPrefix("addons", a0):
		listAddons(w, args[1:])
	case strings.HasPrefix("dynos", a0):
		listDynos(w, args[1:])
	default:
		listApps(w, args)
	}
}

func listApps(w io.Writer, names []string) {
	ch := make(chan error, len(names))
	var apps []*App
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
	sort.Sort(appsByName(apps))
	abbrevEmailApps(apps)
	for _, a := range apps {
		if a.Name != "" {
			listApp(w, a)
		}
	}
}

type appsByName []*App

func (a appsByName) Len() int           { return len(a) }
func (a appsByName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a appsByName) Less(i, j int) bool { return a[i].Name < a[j].Name }

func listRels(w io.Writer, names []string) {
	if len(names) == 0 {
		var rels []*Release
		must(Get(&rels, "/apps/"+mustApp()+"/releases"))
		gitDescribe(rels)
		abbrevEmailReleases(rels)
		for _, r := range rels {
			listRelease(w, r)
		}
		return
	}

	app := mustApp()
	ch := make(chan error, len(names))
	var rels []*Release
	for _, name := range names {
		if name == "" {
			ch <- nil
		} else {
			r, url := new(Release), "/apps/"+app+"/releases/"+name
			rels = append(rels, r)
			go func() { ch <- Get(r, url) }()
		}
	}
	for _ = range names {
		if err := <-ch; err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}
	sort.Sort(releasesByName(rels))
	gitDescribe(rels)
	abbrevEmailReleases(rels)
	for _, r := range rels {
		if r.Name != "" {
			listRelease(w, r)
		}
	}
}

type DynosByName []*Dyno

func (p DynosByName) Len() int           { return len(p) }
func (p DynosByName) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p DynosByName) Less(i, j int) bool { return p[i].Name < p[j].Name }

func listDynos(w io.Writer, names []string) {
	var dynos []*Dyno
	must(Get(&v2{&dynos}, "/apps/"+mustApp()+"/ps"))
	sort.Sort(DynosByName(dynos))

	if len(names) == 0 {
		for _, d := range dynos {
			listDyno(w, d)
		}
		return
	}

	for _, name := range names {
		for _, d := range dynos {
			if d.Name == name {
				listDyno(w, d)
			}
		}
	}
	return
}

func abbrevEmailReleases(rels []*Release) {
	domains := make(map[string]int)
	for _, r := range rels {
		parts := strings.SplitN(r.User, "@", 2)
		if len(parts) == 2 {
			domains["@"+parts[1]]++
		}
	}
	smax, nmax := "", 0
	for s, n := range domains {
		if n > nmax {
			smax = s
		}
	}
	for _, r := range rels {
		if strings.HasSuffix(r.User, smax) {
			r.User = r.User[:len(r.User)-len(smax)]
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

func abbrevEmailResources(ms []*mergedAddon) {
	domains := make(map[string]int)
	for _, m := range ms {
		parts := strings.SplitN(m.Owner, "@", 2)
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
	for _, m := range ms {
		if strings.HasSuffix(m.Owner, smax) {
			m.Owner = m.Owner[:len(m.Owner)-len(smax)]
		}
	}
}

type releasesByName []*Release

func (a releasesByName) Len() int           { return len(a) }
func (a releasesByName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a releasesByName) Less(i, j int) bool { return a[i].Name < a[j].Name }

func listAddons(w io.Writer, names []string) {
	ms := getMergedAddons(mustApp())
	abbrevEmailResources(ms)
	for i, s := range names {
		names[i] = strings.ToLower(s)
	}
	for _, m := range ms {
		if len(names) == 0 || addonMatch(m, names) {
			listAddon(w, m)
		}
	}
}

func addonMatch(m *mergedAddon, a []string) bool {
	for _, s := range a {
		if s == strings.ToLower(m.Type) {
			return true
		}
		if s == strings.ToLower(m.Name) {
			return true
		}
		if s == strings.ToLower(m.ConfigVar) {
			return true
		}
	}
	return false
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

func listRelease(w io.Writer, r *Release) {
	if flagLong {
		listRec(w,
			abbrev(GitRef(r.Commit), 10),
			abbrev(r.User, 10),
			prettyTime{r.CreatedAt.Time},
			r.Name,
			r.Descr,
		)
	} else {
		fmt.Fprintln(w, r.Name)
	}
}

func listDyno(w io.Writer, d *Dyno) {
	if flagLong {
		listRec(w,
			d.Name,
			d.State,
			prettyDuration{d.Age()},
			maybeQuote(d.Command),
		)
	} else {
		fmt.Fprintln(w, d.Name)
	}
}

// quotes s as a json string if it contains any weird chars
// currently weird is anything other than [alnum]_-
func maybeQuote(s string) string {
	for _, r := range s {
		if !('0' <= r && r <= '9' || 'a' <= r && r <= 'z' ||
			'A' <= r && r <= 'Z' || r == '-' || r == '_') {
			return quote(s)
		}
	}
	return s
}

// quotes s as a json string
func quote(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

func listAddon(w io.Writer, m *mergedAddon) {
	if flagLong {
		listRec(w,
			m.Type,
			abbrev(m.Owner, 10),
			m.Name,
			m.ConfigVar,
		)
	} else {
		name := m.ConfigVar
		if name == "" {
			name = "(" + m.Type + ")"
		}
		fmt.Fprintln(w, m.String())
	}
}

type prettyTime struct {
	time.Time
}

func (s prettyTime) String() string {
	if time.Now().Sub(s.Time) < 12*30*24*time.Hour {
		return s.Local().Format("Jan _2 15:04")
	}
	return s.Local().Format("Jan _2  2006")
}

type prettyDuration struct {
	time.Duration
}

func (a prettyDuration) String() string {
	switch d := a.Duration; {
	case d > 2*24*time.Hour:
		return a.Unit(24*time.Hour, "d")
	case d > 2*time.Hour:
		return a.Unit(time.Hour, "h")
	case d > 2*time.Minute:
		return a.Unit(time.Minute, "m")
	}
	return a.Unit(time.Second, "s")
}

func (a prettyDuration) Unit(u time.Duration, s string) string {
	return fmt.Sprintf("%2d", roundDur(a.Duration, u)) + s
}

func roundDur(d, k time.Duration) int {
	return int((d + k/2 - 1) / k)
}

func abbrev(s string, n int) string {
	if len(s) > n {
		return s[:n-1] + "…"
	}
	return s
}

func listRec(w io.Writer, a ...interface{}) {
	for i, x := range a {
		fmt.Fprint(w, x)
		if i+1 < len(a) {
			w.Write([]byte{'\t'})
		} else {
			w.Write([]byte{'\n'})
		}
	}
}
