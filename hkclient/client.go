package hkclient

import (
	"crypto/tls"
	"github.com/bgentry/heroku-go"
	"github.com/heroku/hk/postgresql"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type Clients struct {
	ApiURL string
	Client *heroku.Client

	PgClient *postgresql.Client
}

func New(nrc *NetRc, agent string) (*Clients, error) {
	userAgent := agent + " " + heroku.DefaultUserAgent
	ste := Clients{}

	disableSSLVerify := false
	ste.ApiURL = heroku.DefaultAPIURL
	if s := os.Getenv("HEROKU_API_URL"); s != "" {
		ste.ApiURL = s
		disableSSLVerify = true
	}

	apiURL, err := url.Parse(ste.ApiURL)
	if err != nil {
		return nil, err
	}

	user, pass, err := nrc.GetCreds(apiURL)
	if err != nil {
		return nil, err
	}

	debug := os.Getenv("HKDEBUG") != ""
	ste.Client = &heroku.Client{
		URL:       ste.ApiURL,
		Username:  user,
		Password:  pass,
		UserAgent: userAgent,
		Debug:     debug,
	}
	ste.PgClient = &postgresql.Client{
		Username:  user,
		Password:  pass,
		UserAgent: userAgent,
		Debug:     debug,
	}

	if disableSSLVerify || os.Getenv("HEROKU_SSL_VERIFY") == "disable" {
		tr := http.DefaultTransport.(*http.Transport)
		tr.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}

		ste.Client.HTTP = &http.Client{Transport: tr}
		ste.PgClient.HTTP = &http.Client{Transport: tr}
	}
	if s := os.Getenv("HEROKU_POSTGRESQL_HOST"); s != "" {
		ste.PgClient.StarterURL = "https://" + s +
			".herokuapp.com" + postgresql.DefaultAPIPath

		ste.PgClient.URL = "https://" + s + ".herokuapp.com" +
			postgresql.DefaultAPIPath
	}
	if s := os.Getenv("SHOGUN"); s != "" {
		ste.PgClient.URL = "https://shogun-" + s +
			".herokuapp.com" + postgresql.DefaultAPIPath
	}
	ste.Client.AdditionalHeaders = http.Header{}
	ste.PgClient.AdditionalHeaders = http.Header{}
	for _, h := range strings.Split(os.Getenv("HKHEADER"), "\n") {
		if i := strings.Index(h, ":"); i >= 0 {
			ste.Client.AdditionalHeaders.Set(
				strings.TrimSpace(h[:i]),
				strings.TrimSpace(h[i+1:]),
			)
			ste.PgClient.AdditionalHeaders.Set(
				strings.TrimSpace(h[:i]),
				strings.TrimSpace(h[i+1:]),
			)
		}
	}

	return &ste, nil
}
