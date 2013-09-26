package main

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
)

var cmdReleases = &Command{
	Run:   runReleases,
	Usage: "releases [-l] [name...]",
	Short: "list releases",
	Long: `
Lists releases.

Options:

    -l       long listing

Long listing shows the git commit id, who made the release, time
of the release, version of the release (e.g. 1), and description.

Examples:

    $ hk releases
    v1
    v2

    $ hk releases -l
    3ae20c2  me  Jun 12 18:28  v1  Deploy 3ae20c2
    0fda0ae  me  Jun 13 18:14  v2  Deploy 0fda0ae
    ed39b69  me  Jun 13 18:31  v3  Deploy ed39b69

    $ hk releases -l 3
    ed39b69  me  Jun 13 18:31  v3  Deploy ed39b69
`,
}

func init() {
	cmdReleases.Flag.BoolVar(&flagLong, "l", false, "long listing")
}

func runReleases(cmd *Command, versions []string) {
	w := tabwriter.NewWriter(os.Stdout, 1, 2, 2, ' ', 0)
	listReleases(w, versions)
	w.Flush()
}

func listReleases(w io.Writer, versions []string) {
	var rels []*Release

	if len(versions) == 0 {
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

type releasesByVersion []*Release

func (a releasesByVersion) Len() int           { return len(a) }
func (a releasesByVersion) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a releasesByVersion) Less(i, j int) bool { return a[i].Version < a[j].Version }

