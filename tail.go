package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

var cmdTail = &Command{
	Run:   runTail,
	Usage: "tail",
	Short: "tail log files",
	Long:  `Tail tails log files.`,
}

func runTail(cmd *Command, args []string) {
	data := make(url.Values)
	data.Add("logplex", "true")

	for _, a := range args {
		if a == "-f" {
			data.Add("tail", "1")
		}
	}

	req, err := http.NewRequest("GET", apiURL+"/apps/"+app()+"/logs", strings.NewReader(data.Encode()))
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	req.SetBasicAuth(getCreds(req.URL))
	req.Header.Add("User-Agent", "hk/"+Version)
	req.Header.Add("Accept", "application/json")

	resp := checkResp(http.DefaultClient.Do(req))

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

func checkResp(resp *http.Response, err error) *http.Response {
	if err != nil {
		log.Fatal(err)
	}
	if resp.StatusCode == 401 {
		log.Fatal("Unauthorized")
	}
	if resp.StatusCode == 403 {
		log.Fatal("Unauthorized")
	}
	if resp.StatusCode/100 != 2 { // 200, 201, 202, etc
		log.Fatal("Unexpected error: ", resp.Status)
	}

	if msg := resp.Header.Get("X-Heroku-Warning"); msg != "" {
		fmt.Fprintln(os.Stderr, strings.TrimSpace(msg))
	}

	return resp
}
