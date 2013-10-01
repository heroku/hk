package main

import (
	"bytes"
	"github.com/kr/binarydist"
	"log"
)

func gen(args []string) {
	from := release{Plat: args[1], Cmd: args[0], Ver: args[2]}
	to := release{Plat: args[1], Cmd: args[0], Ver: args[3]}
	genPatch(from, to)
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
	return s3put(patch, s3PatchURL+patchFilename(a.Cmd, a.Plat, a.Ver, b.Ver))
}

func patchFilename(cmd, plat, from, to string) string {
	return cmd + "/" + from + "/" + to + "/" + plat
}
