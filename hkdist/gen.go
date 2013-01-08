package main

import (
	"bytes"
	"github.com/kr/binarydist"
	"log"
	"net/http"
	"time"
)

func gen() {
	var mod time.Time
	for {
		genPatches(&mod)
		time.Sleep(time.Minute)
	}
}

func genPatches(mod *time.Time) {
	r := rels(mod)
	for i, a := range r {
		for _, b := range r[i+1:] {
			if a.Plat == b.Plat && a.Cmd == b.Cmd {
				if exists, err := patchExists(a, b); !exists && err == nil {
					genPatch(a, b)
				} else if err != nil {
					log.Println(err)
				}
			}
		}
	}
}

func patchExists(a, b release) (bool, error) {
	res, err := http.Head(s3PatchURL + patchFilename(a, b))
	if err != nil {
		return false, err
	}
	return res.StatusCode == 200, nil
}

func rels(mod *time.Time) (a []release) {
	url := distURL + "release.json"
	log.Println("fetch rels", url)
	err := fetchJSON(url, mod, &a)
	if err != nil {
		log.Println("fetch rels: ", err)
		return nil
	}
	log.Println("fetch rels finish")
	return a
}

func genPatch(a, b release) {
	log.Println("genPatch", a, b)
	if err := computeAndStorePatch(a, b); err != nil {
		log.Println("genPatch: ", err)
	} else {
		log.Println("genPatch finish")
	}
}

func computeAndStorePatch(a, b release) error {
	ar := newGzReader(fetch(s3DistURL+a.Gzname(), nil))
	defer ar.Close()
	br := newGzReader(fetch(s3DistURL+b.Gzname(), nil))
	defer br.Close()
	patch := new(bytes.Buffer)
	if err := binarydist.Diff(ar, br, patch); err != nil {
		return err
	}
	return s3put(patch, s3PatchURL+patchFilename(a, b))
}

func patchFilename(a, b release) string {
	return s3DistURL + a.Cmd + "-" + a.Ver + "-" + a.Plat + "-to-" + b.Ver
}
