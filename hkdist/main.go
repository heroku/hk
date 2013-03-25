// Command hkdist provides services for distributing hk binaries and updates.
//
// It has three sub-commands: build, web, and gen.
//
//   $ hkdist build
//
// This command fetches source code from github, builds it, uploads the
// binary to and S3 bucket, and posts its SHA-256 hash to the hk distribution
// server (hk.heroku.com in production).
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
	distURL    = os.Getenv("DISTURL")
	s3DistURL  = os.Getenv("S3DISTURL")
	s3PatchURL = os.Getenv("S3PATCHURL")
	buildName  = os.Getenv("BUILDNAME")
	netrcPath  = filepath.Join(os.Getenv("HOME"), ".netrc")
	branch     = os.Getenv("BUILDBRANCH")
	s3keys     = s3.Keys{
		AccessKey: os.Getenv("S3_ACCESS_KEY"),
		SecretKey: os.Getenv("S3_SECRET_KEY"),
	}
)

type release struct {
	Plat, Cmd, Ver string
	Sha256         []byte
}

func (r release) Name() string {
	return r.Cmd + "-" + r.Ver + "-" + r.Plat
}

func (r release) Gzname() string {
	return r.Name() + ".gz"
}

var subcmds = map[string]func(){
	"gen":   gen,
	"build": build,
	"web":   web,
}

func main() {
	log.SetFlags(log.Lshortfile)
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "Usage: hkdist web|gen|build")
		os.Exit(2)
	}
	f := subcmds[os.Args[1]]
	if f == nil {
		fmt.Fprintln(os.Stderr, "Usage: hkdist web|gen|build")
		os.Exit(2)
	}
	f()
}
