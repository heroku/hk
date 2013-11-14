package main

import (
	"io"
	"log"
	"net/http"
	"os"
)

var (
	lines  int
	source string
	dyno   string
)

var cmdLog = &Command{
	Run:   runLog,
	Usage: "log [-n lines] [-s source] [-d dyno]",
	Short: "stream app log lines",
	Long: `
Log prints the streaming application log.

Options:

    -n N       print at most N log lines
    -s source  filter log source
    -d dyno    filter dyno or process type

Examples:

    $ hk log
    2013-10-17T00:17:35.066089+00:00 app[web.1]: Completed 302 Found in 0ms
    2013-10-17T00:17:35.079095+00:00 heroku[router]: at=info method=GET path=/ host=www.heroku.com fwd="1.2.3.4" dyno=web.1 connect=1ms service=6ms status=302 bytes=95
    2013-10-17T00:17:35.505389+00:00 heroku[nginx]: 1.2.3.4 - - [17/Oct/2013:00:17:35 +0000] "GET / HTTP/1.1" 301 5 "-" "Amazon Route 53 Health Check Service" www.heroku.com
		...

    $ hk log -n 2 -s app -d web
		2013-10-17T00:17:34.288521+00:00 app[web.1]: Completed 200 OK in 10ms (Views: 10.0ms)
    2013-10-17T00:17:33.918946+00:00 heroku[web.5]: Started GET "/" for 1.2.3.4 at 2013-10-17 00:17:32 +0000
		2013-10-17T00:17:34.667654+00:00 heroku[router]: at=info method=GET path=/ host=www.heroku.com fwd="1.2.3.4" dyno=web.5 connect=3ms service=8ms status=301 bytes=0
    2013-10-17T00:17:35.079095+00:00 heroku[router]: at=info method=GET path=/ host=www.heroku.com fwd="1.2.3.4" dyno=web.1 connect=1ms service=6ms status=302 bytes=95
		...

    $ hk log -d web.5
    2013-10-17T00:17:33.918946+00:00 app[web.5]: Started GET "/" for 1.2.3.4 at 2013-10-17 00:17:32 +0000
    2013-10-17T00:17:33.918658+00:00 app[web.5]: Processing by PagesController#root as HTML
		...
`,
}

func init() {
	cmdLog.Flag.IntVar(&lines, "n", -1, "max number of log lines to request")
	cmdLog.Flag.StringVar(&source, "s", "", "only display logs from the given source")
	cmdLog.Flag.StringVar(&dyno, "d", "", "only display logs from the given dyno or process type")
}

func runLog(cmd *Command, args []string) {
	var v struct {
		Dyno   string `json:"dyno,omitempty"`
		Lines  int    `json:"lines,omitempty"`
		Source string `json:"source,omitempty"`
		Tail   bool   `json:"tail,omitempty"`
	}

	v.Dyno = dyno
	v.Source = source

	if lines == -1 {
		v.Tail = true
		v.Lines = 10
	} else {
		v.Lines = lines
	}

	var session struct {
		Id         string `json:"id"`
		LogplexURL string `json:"logplex_url"`
	}
	err := APIReq(&session, "POST", "/apps/"+mustApp()+"/log-sessions", v)
	if err != nil {
		log.Fatal(err)
	}
	resp, err := http.Get(session.LogplexURL)
	if err != nil {
		log.Fatal(err)
	}
	must(checkResp(resp))
	if _, err = io.Copy(os.Stdout, resp.Body); err != nil {
		log.Fatal(err)
	}
	resp.Body.Close()
}
