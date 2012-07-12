package main

import (
	"bytes"
	"compress/gzip"
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"github.com/kr/binarydist"
	"io"
	"log"
	"net/http"
	"time"
)

var magic = [8]byte{'h', 'k', 'D', 'I', 'F', 'F', '0', '1'}

func gen() {
	for {
		genPatches()
		converge()
		time.Sleep(time.Minute)
	}
}

type edge struct {
	plat, cmd       string
	newver, newsha1 string
	oldver, oldsha1 string
}

func converge() {
	_, err := db.Exec(`
		delete from next n
		where exists (
			select 1 from cur c
			where n.plat = c.plat and n.cmd = c.cmd and n.oldver = c.curver
		)
	`)
	if err != nil {
		log.Println(err)
	}

	_, err = db.Exec(`
		insert into next (plat, cmd, oldver, newver)
		select p.plat, p.cmd, p.oldver, p.newver
		from patch p join cur c
		on (p.plat = c.plat and p.cmd = c.cmd and p.newver = c.curver)
		where not exists (
			select 1 from next n
			where n.plat = p.plat and n.cmd = p.cmd and n.oldver = p.oldver
		)
	`)
	if err != nil {
		log.Println(err)
	}

	_, err = db.Exec(`
		update next n
		set newver = p.newver
		from patch p join cur c
		on (p.plat = c.plat and p.cmd = c.cmd and p.newver = c.curver)
		where n.plat = p.plat and n.cmd = p.cmd and n.oldver = p.oldver
			and n.newver != p.newver
	`)
	if err != nil {
		log.Println(err)
	}
}

func genPatches() {
	rows, err := db.Query(`
		select t.plat, t.cmd, t.ver, t.sha1, f.ver, f.sha1
		from release f join (
			select r.plat, r.cmd, r.ver, r.sha1
			from cur c join release r
				on (r.plat=c.plat and r.cmd=c.cmd and r.ver=c.curver)
		) as t
		on (f.plat=t.plat and f.cmd=t.cmd and f.ver!=t.ver)
		where not exists (
			select 1 from patch p
			where p.plat=t.plat and p.cmd=t.cmd and
				p.oldver=f.ver and p.newver=t.ver
		)
	`)
	if err != nil {
		log.Print(err)
		return
	}
	var edges []edge
	for rows.Next() {
		var e edge
		err = rows.Scan(&e.plat, &e.cmd, &e.newver, &e.newsha1, &e.oldver, &e.oldsha1)
		if err != nil {
			log.Print(err)
		} else {
			edges = append(edges, e)
		}
	}
	for _, e := range edges {
		logj(J{"event": "gen", "state": "start", "plat": e.plat, "cmd": e.cmd, "newver": e.newver, "newsha1": e.newsha1, "oldver": e.oldver, "oldsha1": e.oldsha1})
		if err = computeAndStorePatch(e); err != nil {
			log.Println(err)
		} else {
			logj(J{"event": "gen", "state": "finish", "plat": e.plat, "cmd": e.cmd, "newver": e.newver, "newsha1": e.newsha1, "oldver": e.oldver, "oldsha1": e.oldsha1})
		}
	}
}

func computeAndStorePatch(e edge) (err error) {
	var oldbuf, newbuf []byte
	if oldbuf, err = fetchGzBytes(s3url(e.oldsha1)); err != nil {
		return err
	}
	if newbuf, err = fetchGzBytes(s3url(e.newsha1)); err != nil {
		return err
	}

	var patch bytes.Buffer
	binarydist.Diff(bytes.NewReader(oldbuf), bytes.NewReader(newbuf), &patch)

	var buf bytes.Buffer
	var header struct {
		Magic    [8]byte
		OldHash  [sha1.Size]byte
		NewHash  [sha1.Size]byte
		DiffHash [sha1.Size]byte
	}
	header.Magic = magic
	sha1Hash(header.OldHash[:], oldbuf)
	sha1Hash(header.NewHash[:], newbuf)
	sha1Hash(header.DiffHash[:], patch.Bytes())
	err = binary.Write(&buf, binary.BigEndian, &header)
	if err != nil {
		return err
	}
	buf.Write(patch.Bytes())

	hkdiffsha1, err := s3put(buf.Bytes(), e.cmd+"-"+e.oldver+"-next.hkdiff")
	if err != nil {
		return err
	}

	_, err = db.Exec(`insert into patch (plat, cmd, oldver, newver, sha1)
		values ($1, $2, $3, $4, $5)
	`, e.plat, e.cmd, e.oldver, e.newver, hkdiffsha1)

	return err
}

func sha1Hash(h, b []byte) {
	s := sha1.New()
	s.Write(b)
	s.Sum(h[:0])
}

func fetchGzBytes(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("response %s fetching %s", resp.Status, url)
	}
	defer resp.Body.Close()

	var buf bytes.Buffer
	gzr, err := gzip.NewReader(resp.Body)
	if err != nil {
		return nil, err
	}
	if _, err = io.Copy(&buf, gzr); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
