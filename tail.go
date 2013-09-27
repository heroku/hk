package main

import (
	"io"
	"log"
	"net/http"
	"os"
)

var (
	follow bool
	lines  int
	source string
	dyno   string
)

var cmdTail = &Command{
	Run:   runTail,
	Usage: "tail [-f] [-n lines] [-s source] [-d dyno]",
	Short: "show the last part of the app log",
	Long:  `Tail prints recent application logs.`,
}

func init() {
	cmdTail.Flag.BoolVar(&follow, "f", false, "do not stop when end of file is reached")
	cmdTail.Flag.IntVar(&lines, "n", -1, "number of log lines to request")
	cmdTail.Flag.StringVar(&source, "s", "", "only display logs from the given source")
	cmdTail.Flag.StringVar(&dyno, "d", "", "only display logs from the given dyno or process type")
}

func runTail(cmd *Command, args []string) {
	var v struct {
		Dyno   string `json:"dyno,omitempty"`
		Lines  int    `json:"lines,omitempty"`
		Source string `json:"source,omitempty"`
		Tail   bool   `json:"tail,omitempty"`
	}

	v.Dyno = dyno
	v.Lines = lines
	v.Source = source
	v.Tail = follow

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
