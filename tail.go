package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

var (
	follow  bool
	lines   int
	source  string
	process string
)

var cmdTail = &Command{
	Run:   runTail,
	Usage: "tail [-f] [-n lines] [-s source] [-p process]",
	Short: "show the last part of the app log",
	Long:  `Tail prints recent application logs.`,
}

func init() {
	cmdTail.Flag.BoolVar(&follow, "f", false, "do not stop when end of file is reached")
	cmdTail.Flag.IntVar(&lines, "n", -1, "number of log lines to request")
	cmdTail.Flag.StringVar(&source, "s", "", "only display logs from the given source")
	cmdTail.Flag.StringVar(&process, "p", "", "only display logs from the given process")
}

func runTail(cmd *Command, args []string) {
	data := make(url.Values)
	data.Add("logplex", "true")

	if follow {
		data.Add("tail", "1")
	}

	if lines > 0 {
		data.Add("num", strconv.Itoa(lines))
	}

	if source != "" {
		data.Add("source", source)
	}

	if process != "" {
		data.Add("ps", process)
	}

	surl := new(logURL)
	err := APIReq(surl, "GET", "/apps/"+mustApp()+"/logs", data)
	if err != nil {
		log.Fatal(err)
	}
	resp, err := http.Get(surl.String())
	if err != nil {
		log.Fatal(err)
	}
	must(checkResp(resp))
	if _, err = io.Copy(os.Stdout, resp.Body); err != nil {
		log.Fatal(err)
	}
	resp.Body.Close()
}

type logURL struct {
	bytes.Buffer
}

func (logURL) Accept() string {
	return "application/json"
}
