package main

import (
	"bufio"
	"crypto/tls"
	"io"
	"log"
	"net"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"strings"
)

var (
	detachedRun bool
)

var cmdRun = &Command{
	Run:   runRun,
	Usage: "run [-a APP] UTILITY [ARG ...]",
	Short: "run a process",
	Long:  `Run a process on Heroku`,
}

func init() {
	cmdRun.Flag.StringVar(&flagApp, "a", "", "app")
	cmdRun.Flag.BoolVar(&detachedRun, "d", false, "detached")
}

func stty(args ...string) *exec.Cmd {
	c := exec.Command("stty", args...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c
}

func runRun(cmd *Command, args []string) {
	data := make(url.Values)
	if !detachedRun {
		data.Add("attach", "true")
	}
	data.Add("command", strings.Join(args, " "))

	resp := struct {
		Url *string `json:"rendezvous_url,omitempty"`
	}{}

	r := APIReq("POST", "/apps/"+app()+"/ps")
	r.SetBodyForm(data)
	r.Do(&resp)

	if detachedRun {
		return
	}

	u, err := url.Parse(*resp.Url)
	if err != nil {
		log.Fatal(err)
	}

	cn, err := net.Dial("tcp", u.Host)
	if err != nil {
		log.Fatal(err)
	}
	tcn := tls.Client(cn, nil)
	defer tcn.Close()

	br := bufio.NewReader(tcn)

	_, err = io.WriteString(tcn, u.Path[1:]+"\r\n")
	if err != nil {
		log.Fatal(err)
	}

	for {
		_, pre, err := br.ReadLine()
		if err != nil {
			log.Fatal(err)
		}
		if !pre {
			break
		}
	}

	err = stty("-icanon", "-echo").Run()
	if err != nil {
		log.Fatal(err)
	}
	defer stty("icanon", "echo").Run()

	sig := make(chan os.Signal)
	signal.Notify(sig)
	go func() {
		<-sig
		// This will get called after the first os.Exit()
		stty("icanon", "echo").Run()
		os.Exit(1)
	}()

	done := make(chan error)
	go func() {
		_, err := io.Copy(os.Stdout, br)
		done <- err
	}()
	go func() {
		_, err := io.Copy(tcn, os.Stdin)
		done <- err
	}()

	err = <-done
	if err != nil {
		log.Fatal(err)
	}

	stty("icanon", "echo").Run()
}
