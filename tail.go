package main

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

var cmdTail = &Command{
	Run:   runTail,
	Usage: "tail [-a APP] [-f] [-n LINES] [-s SOURCE] [-p PROCESS]",
	Short: "tail log files",
	Long:  `Tail tails log files.`,
}

var (
	follow  bool
	lines   int
	source  string
	process string
)

func init() {
	cmdTail.Flag.StringVar(&flagApp, "a", "", "app")
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

	req := APIReq("GET", "/apps/"+mustApp()+"/logs")
	req.SetBodyForm(data)
	resp := checkResp(http.DefaultClient.Do((*http.Request)(req)))

	surl, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	resp.Body.Close()

	resp = checkResp(http.Get(string(surl)))
	_, err = io.Copy(os.Stdout, resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	resp.Body.Close()
}
