package main

import (
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
	Usage: "ls [-l] [resource...]",
	Short: "list addons and releases",
	Long: `
       hk ls [-l] releases [name...]

       hk ls [-l] addons [name...]

Command hk ls lists releases and addons.

Options:

    -l       long listing

Long listing for releases shows the git commit id, who made the
release, time of the release, version of the release (e.g. 1),
and description. Long listing for addons shows the type of the
addon, owner, name of the resource, and the config var it's
attached to.

Examples:

    $ hk ls rel
    v1
    v2

    $ hk ls -l rel
    3ae20c2  me  Jun 12 18:28  v1  Deploy 3ae20c2
    0fda0ae  me  Jun 13 18:14  v2  Deploy 0fda0ae
    ed39b69  me  Jun 13 18:31  v3  Deploy ed39b69

    $ hk ls -l rel 3
    ed39b69  me  Jun 13 18:31  v3  Deploy ed39b69

    $ hk ls addons
    DATABASE_URL
    REDIS_URL

    $ hk ls -l addons REDIS_URL
    redistogo:nano  me  soaring-ably-1234  REDIS_URL
`,
}

func init() {
	cmdLs.Flag.BoolVar(&flagLong, "l", false, "long listing")
}

func runLs(cmd *Command, args []string) {
	w := tabwriter.NewWriter(os.Stdout, 1, 2, 2, ' ', 0)
	list(w, cmd, args)
	w.Flush()
}

func list(w io.Writer, cmd *Command, args []string) {
	switch a0 := args[0]; {
	case strings.HasPrefix("releases", a0):
		listRels(w, args[1:])
	case strings.HasPrefix("addons", a0):
		listAddons(w, args[1:])
	}
}

func listRels(w io.Writer, versions []string) {
	if len(versions) == 0 {
		var rels []*Release
		must(Get(&rels, "/apps/"+mustApp()+"/releases"))
		gitDescribe(rels)
		abbrevEmailReleases(rels)
		sort.Sort(releasesByVersion(rels))
		for _, r := range rels {
			listRelease(w, r)
		}
		return
	}

	app := mustApp()
	ch := make(chan error, len(versions))
	var rels []*Release
	for _, name := range versions {
		if name == "" {
			ch <- nil
		} else {
			r, url := new(Release), "/apps/"+app+"/releases/"+name
			rels = append(rels, r)
			go func() { ch <- Get(r, url) }()
		}
	}
	for _ = range versions {
		if err := <-ch; err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}
	sort.Sort(releasesByVersion(rels))
	gitDescribe(rels)
	abbrevEmailReleases(rels)
	for _, r := range rels {
		listRelease(w, r)
	}
}

func abbrevEmailReleases(rels []*Release) {
	domains := make(map[string]int)
	for _, r := range rels {
		r.Who = r.User.Email
		if a := strings.SplitN(r.Who, "@", 2); len(a) == 2 {
			domains["@"+a[1]]++
		}
	}
	smax, nmax := "", 0
	for s, n := range domains {
		if n > nmax {
			smax = s
		}
	}
	for _, r := range rels {
		if strings.HasSuffix(r.Who, smax) {
			r.Who = r.Who[:len(r.Who)-len(smax)]
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

type releasesByVersion []*Release

func (a releasesByVersion) Len() int           { return len(a) }
func (a releasesByVersion) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a releasesByVersion) Less(i, j int) bool { return a[i].Version < a[j].Version }

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
		if s == strings.ToLower(m.ID) {
			return true
		}
	}
	return false
}

func listRelease(w io.Writer, r *Release) {
	if flagLong {
		listRec(w,
			abbrev(r.Commit, 10),
			abbrev(r.Who, 10),
			prettyTime{r.CreatedAt},
			fmt.Sprintf("%d", r.Version),
			r.Description,
		)
	} else {
		fmt.Fprintln(w, fmt.Sprintf("v%d", r.Version))
	}
}

func listAddon(w io.Writer, m *mergedAddon) {
	if flagLong {
		listRec(w,
			m.Type,
			abbrev(m.Owner, 10),
			m.ID,
		)
	} else {
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
