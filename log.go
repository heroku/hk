package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"

	"github.com/bgentry/heroku-go"
	"github.com/mgutz/ansi"
)

var (
	lines  int
	source string
	dyno   string
)

var cmdLog = &Command{
	Run:      runLog,
	Usage:    "log [-n <lines>] [-s <source>] [-d <dyno>]",
	NeedsApp: true,
	Category: "app",
	Short:    "stream app log lines",
	Long: `
Log prints the streaming application log.

Options:

    -n <N>       print at most N log lines
    -s <source>  filter log source
    -d <dyno>    filter dyno or process type

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
	cmdLog.Flag.IntVarP(&lines, "number", "n", -1, "max number of log lines to request")
	cmdLog.Flag.StringVarP(&source, "source", "s", "", "only display logs from the given source")
	cmdLog.Flag.StringVarP(&dyno, "dyno", "d", "", "only display logs from the given dyno or process type")
}

func runLog(cmd *Command, args []string) {
	if len(args) != 0 {
		cmd.PrintUsage()
		os.Exit(2)
	}

	opts := heroku.LogSessionCreateOpts{}
	if dyno != "" {
		opts.Dyno = &dyno
	}
	if source != "" {
		opts.Source = &source
	}

	if lines != -1 {
		opts.Lines = &lines
	} else {
		tailopt := true
		lineopt := 10
		opts.Tail = &tailopt
		opts.Lines = &lineopt
	}

	session, err := client.LogSessionCreate(mustApp(), &opts)
	if err != nil {
		printFatal(err.Error())
	}
	resp, err := http.Get(session.LogplexURL)
	if err != nil {
		printFatal(err.Error())
	}
	if warning := resp.Header.Get("X-Heroku-Warning"); warning != "" {
		printWarning(warning)
	}
	if resp.StatusCode/100 != 2 {
		if resp.StatusCode/100 == 4 {
			printFatal("Unauthorized")
		} else {
			printFatal("Unexpected error: " + resp.Status)
		}
	}

	// colors are disabled globally in main() depending on term.IsTerminal()
	writer := newColorizer(os.Stdout)

	scanner := bufio.NewScanner(resp.Body)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		_, err = writer.Writeln(scanner.Text())
		must(err)
	}

	resp.Body.Close()
}

type colorizer struct {
	colors      map[string]string
	colorScheme []string
	filter      *regexp.Regexp
	writer      io.Writer
}

func newColorizer(writer io.Writer) *colorizer {
	return &colorizer{
		colors: make(map[string]string),
		colorScheme: []string{
			"cyan",
			"yellow",
			"green",
			"magenta",
			"red",
		},
		filter: regexp.MustCompile(`(?s)^(.*?\[([\w-]+)(?:[\d\.]+)?\]:)(.*)?$`),
		writer: writer,
	}
}

func (c *colorizer) resolve(p string) string {
	if color, ok := c.colors[p]; ok {
		return color
	}

	color := c.colorScheme[len(c.colors)%len(c.colorScheme)]
	c.colors[p] = color
	return color
}

func (c *colorizer) Writeln(p string) (n int, err error) {
	if c.filter.MatchString(p) {
		submatches := c.filter.FindStringSubmatch(p)
		return fmt.Fprintln(c.writer, ansi.Color(submatches[1], c.resolve(submatches[2]))+ansi.ColorCode("reset")+submatches[3])
	}

	return fmt.Fprintln(c.writer, p)
}
