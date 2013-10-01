// Command hkdist provides services for distributing hk binaries and updates.
//
// It has three sub-commands: build, web, and gen.
//
//   $ hkdist build [platforms]
//
// This command builds cross-compiled binaries. The tool builds all known
// platforms by default, but will optionally build for a specified list of
// platforms instead.  It first fetches the source code and termines the latest
// git tag on BUILDBRANCH.  Then, for each platform, it builds a binary
// executable, uploads the binary to an S3 bucket, and posts its SHA-256 hash
// to the hk distribution server (hk.heroku.com in production).
//
//   $ hkdist web
//
// This command provides directory service for hk binary hashes.
//
//   $ hkdist gen
//
// This command polls the distribution server to learn about new releases,
// then generates byte-sequence patches between each pair of releases on
// each platform. It puts these patches in an S3 bucket so the hk client
// can use them for self-update instead of downloading a (much larger) full
// release.
package main

import (
	"fmt"
	"github.com/kr/s3"
	"log"
	"os"
	"path/filepath"
)

var (
	distURL      = os.Getenv("DISTURL")
	s3DistURL    = os.Getenv("S3DISTURL")
	s3PatchURL   = os.Getenv("S3PATCHURL")
	buildName    = os.Getenv("BUILDNAME")
	netrcPath    = filepath.Join(os.Getenv("HOME"), ".netrc")
	buildbranch  = os.Getenv("BUILDBRANCH")
	hkgenAppName = os.Getenv("HKGENAPPNAME")
	s3keys       = s3.Keys{
		AccessKey: os.Getenv("S3_ACCESS_KEY"),
		SecretKey: os.Getenv("S3_SECRET_KEY"),
	}
)

type release struct {
	Plat, Cmd, Ver string
	Sha256         []byte
}

func (r release) Name() string {
	return r.Cmd + "/" + r.Ver + "/" + r.Plat
}

func (r release) Gzname() string {
	return r.Name() + ".gz"
}

var subcmds = map[string]func([]string){
	"gen":   gen,
	"build": build,
	"web":   web,
}

func usage() {
	fmt.Fprintln(os.Stderr, "Usage: hkdist (web|gen|build [platforms])")
	os.Exit(2)
}

func main() {
	log.SetFlags(log.Lshortfile)
	if len(os.Args) < 2 {
		usage()
	} else if os.Args[1] == "web" && len(os.Args) != 2 {
		usage()
	} else if os.Args[1] == "gen" && len(os.Args) != 6 {
		usage()
	}
	f := subcmds[os.Args[1]]
	if f == nil {
		usage()
	}
	f(os.Args[2:])
}
