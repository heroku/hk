package hkclient

import (
	"crypto/tls"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/bgentry/heroku-go"
	"github.com/heroku/hk/postgresql"
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

	tr := &http.Transport{}
	ste.Client.HTTP = &http.Client{Transport: tr}
	ste.PgClient.HTTP = &http.Client{Transport: tr}

	if disableSSLVerify || os.Getenv("HEROKU_SSL_VERIFY") == "disable" {
		tr.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
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

	herokuAgentSocket := os.Getenv("HEROKU_AGENT_SOCK")
	if herokuAgentSocket != "" {
		// expand a tilde (i.e. `~/.heroku-agent.sock`)
		if herokuAgentSocket[0] == '~' {
			herokuAgentSocket = homePath() + herokuAgentSocket[1:]
		}

		tr.Dial = func(_ string, _ string) (net.Conn, error) {
			return net.Dial("unix", herokuAgentSocket)
		}

		ste.Client.HerokuAgentSocket = herokuAgentSocket
		ste.PgClient.HerokuAgentSocket = herokuAgentSocket
	}

	return &ste, nil
}
