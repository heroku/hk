package main

import (
	"encoding/json"
	"github.com/bmizerany/pat"
	"log"
	"net/http"
	"os"
	"strings"
)

var curver = os.Getenv("CURVER")

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

func main() {
	log.SetFlags(0)
	m := pat.New()
	m.Get("/:plat/:oldver/next.hkdiff", http.HandlerFunc(patch))
	m.Get("/:plat/:ver/hk.gz", http.HandlerFunc(full))
	m.Get("/:plat/hk.gz", http.HandlerFunc(full))
	m.Get("/hk.gz", http.HandlerFunc(full))
	m.Get("/", http.FileServer(http.Dir("dist/public")))
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

func patch(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	plat := q.Get(":plat")
	switch plat {
	case "darwin-amd64", "linux-amd64":
	default:
		http.NotFound(w, r)
		return
	}
	oldver := q.Get(":oldver")
	to := "https://github.com/downloads/kr/hk/" + plat + "-" + oldver + "-next.hkdiff"
	reqlogj(r, J{
		"updatereq": "patch",
		"plat":      plat,
		"oldver":    oldver,
		"to":        to,
	})
	http.Redirect(w, r, to, 307)
}

func full(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	var guess bool
	plat := q.Get(":plat")
	switch plat {
	case "":
		plat = guessPlat(r.UserAgent())
		guess = true
	case "darwin-amd64", "linux-amd64":
	default:
		http.NotFound(w, r)
		return
	}
	ver := q.Get(":ver")
	if ver == "" {
		ver = curver
	}
	to := "https://github.com/downloads/kr/hk/" + plat + "-hk-" + ver + ".gz"
	reqlogj(r, J{
		"updatereq": "full",
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
	b, err := json.Marshal(j)
	if err != nil {
		return
	}
	log.Println(string(b))
}
