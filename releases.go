package main

import (
	"fmt"
	"github.com/bgentry/heroku-go"
	"io"
	"log"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"time"
)

var cmdReleases = &Command{
	Run:      runReleases,
	Usage:    "releases [-l] [name...]",
	Category: "release",
	Short:    "list releases",
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
	defer w.Flush()
	listReleases(w, versions)
}

func listReleases(w io.Writer, versions []string) {
	appname := mustApp()
	if len(versions) == 0 {
		hrels, err := client.ReleaseList(appname, nil)
		must(err)
		rels := make([]*Release, len(hrels))
		for i := range hrels {
			rels[i] = newRelease(&hrels[i])
		}
		sort.Sort(releasesByVersion(rels))
		gitDescribe(rels)
		abbrevEmailReleases(rels)
		for _, r := range rels {
			listRelease(w, r)
		}
		return
	}

	var rels []*Release
	relch := make(chan *heroku.Release, len(versions))
	errch := make(chan error, len(versions))
	for _, name := range versions {
		if name == "" {
			relch <- nil
		} else {
			go func(relname string) {
				if rel, err := client.ReleaseInfo(appname, relname); err != nil {
					errch <- err
				} else {
					relch <- rel
				}
			}(name)
		}
	}
	for _ = range versions {
		select {
		case err := <-errch:
			fmt.Fprintln(os.Stderr, err)
		case rel := <-relch:
			if rel != nil {
				rels = append(rels, newRelease(rel))
			}
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

func newRelease(rel *heroku.Release) *Release {
	return &Release{*rel, "", ""}
}

var cmdReleaseInfo = &Command{
	Run:      runReleaseInfo,
	Usage:    "release-info <version>",
	Category: "release",
	Short:    "show release info",
	Long:     `release-info shows detailed information about a release.`,
}

func runReleaseInfo(cmd *Command, args []string) {
	if len(args) != 1 {
		log.Fatal("Invalid usage. See 'hk help release-info'")
	}
	ver := strings.TrimPrefix(args[0], "v")
	rel, err := client.ReleaseInfo(mustApp(), ver)
	must(err)

	fmt.Printf("Version:  v%d\n", rel.Version)
	fmt.Printf("By:       %s\n", rel.User.Email)
	fmt.Printf("Change:   %s\n", rel.Description)
	fmt.Printf("When:     %s\n", rel.CreatedAt.UTC().Format(time.RFC3339))
	fmt.Printf("Id:       %s\n", rel.Id)
	fmt.Printf("Slug:     %s\n", rel.Slug.Id)
}

var cmdRollback = &Command{
	Run:      runRollback,
	Usage:    "rollback <version>",
	Category: "release",
	Short:    "rolback to a previous release",
}

func runRollback(cmd *Command, args []string) {
	if len(args) != 1 {
		log.Fatal("Invalid usage. See 'hk help rollback'")
	}
	ver := strings.TrimPrefix(args[0], "v")
	rel, err := client.ReleaseRollback(mustApp(), ver)
	must(err)
	fmt.Printf("Rolled back to v%s as v%d.\n", ver, rel.Version)
}
