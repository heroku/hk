package main

import (
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"crypto/sha1"
	"database/sql"
	"encoding/base32"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/bmizerany/pat"
	"github.com/kr/s3"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	bucket = os.Getenv("BUCKET")
	s3keys = s3.Keys{
		os.Getenv("S3_ACCESS_KEY"),
		os.Getenv("S3_SECRET_KEY"),
	}
)

var db *sql.DB

type J map[string]interface{}

type httpsOnly struct {
	http.Handler
}

func (x httpsOnly) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("X-Forwarded-Proto") != "https" {
		http.Error(w, "use https", http.StatusForbidden)
		return
	}
	w.Header().Set("Strict-Transport-Security", "max-age=31536000")
	x.Handler.ServeHTTP(w, r)
}

type authenticate struct {
	http.Handler
}

func (x authenticate) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	hr, _ := http.NewRequest("GET", "https://api.heroku.com/user", nil)
	hr.Header.Set("Accept", "application/json")
	hr.Header.Set("Authorization", r.Header.Get("Authorization"))
	res, err := http.DefaultClient.Do(hr)
	if err != nil {
		log.Println(err)
		http.Error(w, "internal error", 500)
		return
	}
	if res.StatusCode == 401 {
		http.Error(w, "unauthorized", 401)
		return
	}
	if res.StatusCode != 200 {
		log.Println("unexpected status from heroku api:", res.StatusCode)
		http.Error(w, "internal error", 500)
		return
	}

	var info struct {
		Email string
	}
	err = json.NewDecoder(res.Body).Decode(&info)
	res.Body.Close()
	if err != nil {
		log.Println(err)
		http.Error(w, "internal error", 500)
		return
	}

	r.Header.Set(":email", info.Email)
	x.Handler.ServeHTTP(w, r)
}

type herokaiOnly struct {
	http.Handler
}

func (x herokaiOnly) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasSuffix(r.Header.Get(":email"), "@heroku.com") {
		http.Error(w, "unauthorized", 401)
		return
	}
	x.Handler.ServeHTTP(w, r)
}

func main() {
	log.SetFlags(log.Lshortfile)
	db = initdb()
	if len(os.Args) == 2 && os.Args[1] == "gen" {
		gen()
		return
	}
	m := pat.New()
	m.Put("/:cmd-:ver-:plat", authenticate{herokaiOnly{http.HandlerFunc(put)}})
	m.Post("/:cmd-:plat", authenticate{herokaiOnly{http.HandlerFunc(setcur)}})
	m.Get("/:cmd-:ver-:plat.gz", http.HandlerFunc(full))
	m.Get("/:cmd-:oldver-next-:plat.hkdiff", http.HandlerFunc(patch))
	m.Get("/:plat/:oldver/next.hkdiff", http.HandlerFunc(patch)) // for compat
	m.Get("/:plat/:cmd.gz", http.HandlerFunc(full))
	m.Get("/hk.gz", http.HandlerFunc(full)) // for the html page
	m.Get("/release.txt", http.HandlerFunc(lsrelease))
	m.Get("/cur.txt", http.HandlerFunc(lscur))
	m.Get("/patch.txt", http.HandlerFunc(lspatch))
	m.Get("/next.txt", http.HandlerFunc(lsnext))
	m.Get("/", http.FileServer(http.Dir("hkdist/public")))
	var root http.Handler = m
	if os.Getenv("HTTPSONLY") != "" {
		root = httpsOnly{m}
	}
	http.Handle("/", root)
	err := http.ListenAndServe(":"+os.Getenv("PORT"), nil)
	if err != nil {
		log.Fatalf(`{"func":"ListenAndServe", "error":%q}`, err)
	}
}

func put(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	q := r.URL.Query()
	plat := q.Get(":plat")
	cmd := q.Get(":cmd")
	ver := q.Get(":ver")
	if strings.IndexFunc(plat, badIdentRune) >= 0 ||
		strings.IndexFunc(cmd, badIdentRune) >= 0 ||
		strings.IndexFunc(ver, badVersionRune) >= 0 {
		http.Error(w, "bad character in path", 400)
		return
	}

	body, err := ioutil.ReadAll(http.MaxBytesReader(w, r.Body, 10e6))
	if err != nil && err.Error() == "http: request body too large" {
		http.Error(w, "too big", 413)
		return
	}
	if err != nil {
		log.Println(err)
		http.Error(w, "internal error", 500)
		return
	}

	var buf bytes.Buffer
	gz, _ := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	gz.Name = cmd + "-" + ver
	gz.Write(body)
	gz.Close()
	sha1, err := s3put(buf.Bytes(), gz.Name+".gz")
	if err != nil {
		log.Println(err)
		w.WriteHeader(500)
		return
	}

	_, err = db.Exec(`
		insert into release (plat, cmd, ver, sha1)
		values ($1, $2, $3, $4)
	`, plat, cmd, ver, sha1)
	if err != nil {
		log.Println(err)
		http.Error(w, "internal error", 500)
		return
	}
	w.WriteHeader(201)
	w.Write([]byte("created\n"))
}

func setcur(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	q := r.URL.Query()
	plat := q.Get(":plat")
	cmd := q.Get(":cmd")
	if strings.IndexFunc(plat, badIdentRune) >= 0 ||
		strings.IndexFunc(cmd, badIdentRune) >= 0 {
		http.Error(w, "bad character in path", 400)
		return
	}

	body, err := ioutil.ReadAll(http.MaxBytesReader(w, r.Body, 100))
	if err != nil && err.Error() == "http: request body too large" {
		http.Error(w, "too big", 413)
		return
	}
	if err != nil {
		log.Println(err)
		http.Error(w, "internal error", 500)
		return
	}
	ver := string(body)

	_, err = db.Exec(`
		update cur set curver=$1 where plat=$2 and cmd=$3
	`, ver, plat, cmd)
	if err != nil {
		log.Println(err)
		http.Error(w, "internal error", 500)
		return
	}
	_, err = db.Exec(`
		insert into cur (plat, cmd, curver)
		select $1, $2, $3
		where not exists (select 1 from cur where plat=$1 and cmd=$2)
	`, plat, cmd, ver)
	if err != nil {
		log.Println(err)
		http.Error(w, "internal error", 500)
		return
	}
	w.WriteHeader(200)
	w.Write([]byte("ok\n"))
}

func s3url(sha1 string) string {
	return "https://" + bucket + ".s3.amazonaws.com/" + sha1
}

func s3put(data []byte, filename string) (sha1 string, err error) {
	sha1 = string(b32sha1(data))
	url := s3url(sha1)
	r, _ := http.NewRequest("PUT", url, bytes.NewReader(data))
	r.ContentLength = int64(len(data))
	r.Header.Set("Date", time.Now().UTC().Format(http.TimeFormat))
	r.Header.Set("Content-Disposition", "attachment; filename="+filename)
	r.Header.Set("X-Amz-Acl", "public-read")
	r.Header.Set("Content-Md5", string(b16md5(data)))
	s3.Sign(r, s3keys)
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("s3 status %v putting %q: %q", resp.Status, url, string(body))
	}
	return sha1, nil
}

func lsrelease(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(`select plat, cmd, ver, sha1 from release`)
	if err != nil {
		log.Print(err)
		w.WriteHeader(500)
		return
	}

	for rows.Next() {
		var plat, cmd, ver, sha1 string
		err = rows.Scan(&plat, &cmd, &ver, &sha1)
		if err != nil {
			log.Print(err)
		} else {
			fmt.Fprintln(w, plat, cmd, ver, sha1)
		}
	}
}

func lscur(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(`select plat, cmd, curver from cur`)
	if err != nil {
		log.Print(err)
		w.WriteHeader(500)
		return
	}

	for rows.Next() {
		var plat, cmd, ver string
		err = rows.Scan(&plat, &cmd, &ver)
		if err != nil {
			log.Print(err)
		} else {
			fmt.Fprintln(w, plat, cmd, ver)
		}
	}
}

func lspatch(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(`select plat, cmd, oldver, newver, sha1 from patch`)
	if err != nil {
		log.Print(err)
		w.WriteHeader(500)
		return
	}

	for rows.Next() {
		var plat, cmd, oldver, newver, sha1 string
		err = rows.Scan(&plat, &cmd, &oldver, &newver, &sha1)
		if err != nil {
			log.Print(err)
		} else {
			fmt.Fprintln(w, plat, cmd, oldver, newver, sha1)
		}
	}
}

func lsnext(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(`select plat, cmd, oldver, newver from next`)
	if err != nil {
		log.Print(err)
		w.WriteHeader(500)
		return
	}

	for rows.Next() {
		var plat, cmd, oldver, newver string
		err = rows.Scan(&plat, &cmd, &oldver, &newver)
		if err != nil {
			log.Print(err)
		} else {
			fmt.Fprintln(w, plat, cmd, oldver, newver)
		}
	}
}

func patch(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	plat := q.Get(":plat")
	cmd := q.Get(":cmd")
	if cmd == "" {
		cmd = "hk"
	}
	oldver := q.Get(":oldver")

	var newver, sha1 string
	err := db.QueryRow(`
		select n.newver, sha1
		from patch p join next n
		on (p.plat=n.plat and
			p.cmd=n.cmd and
			p.oldver=n.oldver and
			p.newver=n.newver)
		where n.plat=$1 and n.cmd=$2 and n.oldver=$3
	`, plat, cmd, oldver).Scan(&newver, &sha1)
	switch err {
	case nil:
	case sql.ErrNoRows:
		http.NotFound(w, r)
		return
	default:
		log.Println(err)
		w.WriteHeader(500)
		return
	}

	to := "https://" + bucket + ".s3.amazonaws.com/" + sha1
	reqlogj(r, J{
		"event":  "patchreq",
		"plat":   plat,
		"oldver": oldver,
		"newver": newver,
		"to":     to,
	})
	http.Redirect(w, r, to, 307)
}

func full(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	var guess bool
	plat := q.Get(":plat")
	if plat == "" {
		plat = guessPlat(r.UserAgent())
		guess = true
	}
	cmd := q.Get(":cmd")
	if cmd == "" {
		cmd = "hk"
	}
	ver := q.Get(":ver")
	if ver == "" {
		const s = `select curver from cur where plat=$1 and cmd=$2`
		switch err := db.QueryRow(s, plat, cmd).Scan(&ver); err {
		case nil:
		case sql.ErrNoRows:
			http.NotFound(w, r)
			return
		default:
			log.Println(err)
			w.WriteHeader(500)
			return
		}

	}

	var sha1 string
	const s = `select sha1 from release where plat=$1 and cmd=$2 and ver=$3`
	switch err := db.QueryRow(s, plat, cmd, ver).Scan(&sha1); err {
	case nil:
	case sql.ErrNoRows:
		http.NotFound(w, r)
		return
	default:
		log.Println(err)
		w.WriteHeader(500)
		return
	}

	to := "https://" + bucket + ".s3.amazonaws.com/" + sha1
	reqlogj(r, J{
		"event":     "fullreq",
		"plat":      plat,
		"guessplat": guess,
		"ver":       ver,
		"to":        to,
	})
	http.Redirect(w, r, to, 307)
}

func guessPlat(ua string) string {
	ua = strings.ToLower(ua)
	if strings.Contains(ua, "mac os x") || strings.Contains(ua, "darwin") {
		return "darwin-amd64"
	}
	return "linux-amd64"
}

func reqlogj(r *http.Request, j J) {
	j["ua"] = r.UserAgent()
	j["reqpath"] = r.URL.Path
	logj(j)
}

func logj(j J) {
	b, err := json.Marshal(j)
	if err != nil {
		return
	}
	log.Println(string(b))
}

// returns the base32-encoded sha1 of b
func b32sha1(b []byte) []byte {
	h := sha1.New()
	h.Write(b)
	s := make([]byte, base32.StdEncoding.EncodedLen(sha1.Size))
	base32.StdEncoding.Encode(s, h.Sum(nil))
	return s
}

// returns the base16-encoded md5 of b
func b16md5(b []byte) []byte {
	h := md5.New()
	h.Write(b)
	s := make([]byte, base64.StdEncoding.EncodedLen(md5.Size))
	base64.StdEncoding.Encode(s, h.Sum(nil))
	return s
}

func badIdentRune(r rune) bool {
	return !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-')
}

func badVersionRune(r rune) bool {
	return !(r >= '0' && r <= '9' || r == '.')
}
