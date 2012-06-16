package main

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
)

var cmdTail = &Command{
	Run:   runTail,
	Usage: "tail [-a APP] [-f]",
	Short: "tail log files",
	Long:  `Tail tails log files.`,
}

var follow bool

func init() {
	cmdTail.Flag.StringVar(&flagApp, "a", "", "app")
	cmdTail.Flag.BoolVar(&follow, "f", false, "do not stop when end of file is reached")
}

func runTail(cmd *Command, args []string) {
	data := make(url.Values)
	data.Add("logplex", "true")

	if follow {
		data.Add("tail", "1")
	}

	req := APIReq("GET", "/apps/"+app()+"/logs")
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
