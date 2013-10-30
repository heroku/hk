package main

import (
	"bytes"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"github.com/bmizerany/pat"
	"github.com/bmizerany/pq"
	"github.com/kr/secureheader"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	pgUniqueViolation = "23505"
)

var db *sql.DB

// Examples:
//
//   PUT /hk-1-linux-386.json
//   PUT /hk-linux-386.json
//
//   GET /hk-current-linux-386.json
//   GET /hk-1-linux-386.json
//   GET /hk.gz
func web(args []string) {
	mustHaveEnv("DATABASE_URL")
	initwebdb()
	m := pat.New()
	m.Get("/:cmd.gz", http.HandlerFunc(initial))
	m.Get("/:cmd-current-:plat.json", http.HandlerFunc(curInfo))
	m.Get("/:cmd-:ver-:plat.json", http.HandlerFunc(getHash))
	m.Get("/release.json", http.HandlerFunc(listReleases))
	m.Put("/:cmd-:ver-:os-:arch.json", authenticate{herokaiOnly{http.HandlerFunc(putVer)}})
	m.Put("/:cmd-:os-:arch.json", authenticate{herokaiOnly{http.HandlerFunc(setCur)}})
	m.Get("/", http.FileServer(http.Dir("hkdist/public")))
	http.Handle("/", m)
	secureheader.DefaultConfig.PermitClearLoopback = true
	err := http.ListenAndServe(":"+os.Getenv("PORT"), secureheader.DefaultConfig)
	if err != nil {
		log.Fatalf(`{"func":"ListenAndServe", "error":%q}`, err)
	}
}

func setCur(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	q := r.URL.Query()
	plat := q.Get(":os") + "-" + q.Get(":arch")
	cmd := q.Get(":cmd")
	if strings.IndexFunc(plat, badIdentRune) >= 0 ||
		strings.IndexFunc(cmd, badIdentRune) >= 0 {
		http.Error(w, "bad character in path", 400)
		return
	}

	var info struct{ Version string }
	if !readReqJSON(w, r, 1000, &info) {
		return
	}
	_, err := db.Exec(`
		update cur set curver=$1 where plat=$2 and cmd=$3
	`, info.Version, plat, cmd)
	if err != nil {
		log.Println(err)
		http.Error(w, "internal error", 500)
		return
	}
	_, err = db.Exec(`
		insert into cur (plat, cmd, curver)
		select $1, $2, $3
		where not exists (select 1 from cur where plat=$1 and cmd=$2)
	`, plat, cmd, info.Version)
	if err != nil {
		log.Println(err)
		http.Error(w, "internal error", 500)
		return
	}
	if _, err = db.Exec(`update mod set t=now()`); err != nil {
		log.Println(err)
		http.Error(w, "internal error", 500)
		return
	}
	io.WriteString(w, "ok\n")
}

func scan(w http.ResponseWriter, r *http.Request, q *sql.Row, v ...interface{}) bool {
	switch err := q.Scan(v...); err {
	case nil:
	case sql.ErrNoRows:
		http.NotFound(w, r)
		return false
	default:
		log.Println(err)
		w.WriteHeader(500)
		return false
	}
	return true
}

func lookupCurInfo(w http.ResponseWriter, r *http.Request, plat, cmd string) (v struct {
	Version string
	Sha256  []byte
}, ok bool) {
	const s = `select c.curver, r.sha256 from cur c, release r
				where c.plat=$1 and c.cmd=$2
				and c.plat = r.plat and c.cmd = r.cmd and c.curver = r.ver`
	ok = scan(w, r, db.QueryRow(s, plat, cmd), &v.Version, &v.Sha256)
	return
}

func initial(w http.ResponseWriter, r *http.Request) {
	cmd := r.URL.Query().Get(":cmd")
	plat := guessPlat(r.UserAgent())
	if info, ok := lookupCurInfo(w, r, plat, cmd); ok {
		url := s3DistURL + cmd + "-" + info.Version + "-" + plat + ".gz"
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	}
}

func curInfo(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	if info, ok := lookupCurInfo(w, r, q.Get(":plat"), q.Get(":cmd")); ok {
		logErr(json.NewEncoder(w).Encode(info))
	}
}

func getHash(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	var info jsonsha
	const s = `select sha256 from release where plat=$1 and cmd=$2 and ver=$3`
	if scan(w, r, db.QueryRow(s, q.Get(":plat"), q.Get(":cmd"), q.Get(":ver")), &info.Sha256) {
		logErr(json.NewEncoder(w).Encode(info))
	}
}

func listReleases(w http.ResponseWriter, r *http.Request) {
	rels := make([]release, 0)
	rows, err := db.Query(`select plat, cmd, ver, sha256 from release`)
	if err != nil {
		log.Println(err)
		http.Error(w, "internal error", 500)
		return
	}
	for rows.Next() {
		var rel release
		err := rows.Scan(&rel.Plat, &rel.Cmd, &rel.Ver, &rel.Sha256)
		if err != nil {
			log.Println(err)
		} else {
			rels = append(rels, rel)
		}
	}
	if err := rows.Err(); err != nil {
		log.Println(err)
		http.Error(w, "internal error", 500)
		return
	}
	b := new(bytes.Buffer)
	if err = json.NewEncoder(b).Encode(rels); err != nil {
		log.Println(err)
		http.Error(w, "internal error", 500)
		return
	}
	var mod time.Time
	db.QueryRow(`select t from mod`).Scan(&mod)
	http.ServeContent(w, r, "", mod, bytes.NewReader(b.Bytes()))
}

func logErr(err error) error {
	if err != nil {
		log.Println(err)
	}
	return err
}

func guessArch(ua string) string {
	if strings.Contains(ua, "amd64") || strings.Contains(ua, "x86_64") {
		return "amd64"
	}
	return "386"
}

func guessOS(ua string) string {
	ua = strings.ToLower(ua)
	if strings.Contains(ua, "windows") {
		return "windows"
	}
	if strings.Contains(ua, "mac os x") || strings.Contains(ua, "darwin") {
		return "darwin"
	}
	return "linux"
}

func guessPlat(ua string) string {
	return guessOS(ua) + "-" + guessArch(ua)
}

func putVer(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	q := r.URL.Query()
	plat := q.Get(":os") + "-" + q.Get(":arch")
	cmd := q.Get(":cmd")
	ver := q.Get(":ver")
	if strings.IndexFunc(plat, badIdentRune) >= 0 ||
		strings.IndexFunc(cmd, badIdentRune) >= 0 ||
		strings.IndexFunc(ver, badVersionRune) >= 0 {
		http.Error(w, "bad character in path", 400)
		return
	}

	var info jsonsha
	if !readReqJSON(w, r, 1000, &info) {
		return
	}
	if len(info.Sha256) != sha256.Size {
		log.Printf("bad hash length %d != %d", len(info.Sha256), sha256.Size)
		http.Error(w, "unprocessable entity", 422)
		return
	}

	_, err := db.Exec(`
		insert into release (plat, cmd, ver, sha256)
		values ($1, $2, $3, $4)
	`, plat, cmd, ver, info.Sha256)
	if pe, ok := err.(pq.PGError); ok && pe.Get('C') == pgUniqueViolation {
		http.Error(w, "conflict", http.StatusConflict)
		return
	} else if err != nil {
		log.Println(err)
		http.Error(w, "internal error", 500)
		return
	}
	if _, err = db.Exec(`update mod set t=now()`); err != nil {
		log.Println(err)
		http.Error(w, "internal error", 500)
		return
	}
	w.WriteHeader(http.StatusCreated)
	io.WriteString(w, "created\n")
}

func readReqJSON(w http.ResponseWriter, r *http.Request, n int64, v interface{}) bool {
	err := json.NewDecoder(http.MaxBytesReader(w, r.Body, n)).Decode(v)
	if err != nil {
		http.Error(w, "unprocessable entity", 422)
	}
	return err == nil
}

type authenticate struct {
	http.Handler
}

func (x authenticate) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	hr, _ := http.NewRequest("GET", "https://api.heroku.com/account", nil)
	hr.Header.Set("Accept", "application/vnd.heroku+json; version=3")
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

func mustExec(q string) {
	if _, err := db.Exec(q); err != nil {
		log.Fatal(err)
	}
}

func initwebdb() {
	connstr, err := pq.ParseURL(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("pq.ParseURL", err)
	}
	db, err = sql.Open("postgres", connstr+" sslmode=disable")
	if err != nil {
		log.Fatal("sql.Open", err)
	}
	mustExec(`SET bytea_output = 'hex'`) // work around https://github.com/bmizerany/pq/issues/76
	mustExec(`create table if not exists release (
		plat text not null,
		cmd text not null,
		ver text not null,
		sha256 bytea not null,
		primary key (plat, cmd, ver)
	)`)
	mustExec(`create table if not exists cur (
		plat text not null,
		cmd text not null,
		curver text not null,
		foreign key (plat, cmd, curver) references release (plat, cmd, ver),
		primary key (plat, cmd)
	)`)
	mustExec(`create table if not exists mod (
		t timestamptz not null
	)`)
	mustExec(`insert into mod (t)
		select now()
		where not exists (select 1 from mod)
	`)
}

func badIdentRune(r rune) bool {
	return !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-')
}

func badVersionRune(r rune) bool {
	return !(r >= '0' && r <= '9' || r == '.')
}
