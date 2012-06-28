package main

import (
	"bufio"
	"crypto/tls"
	"io"
	"log"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
)

var (
	detachedRun bool
)

var cmdRun = &Command{
	Run:   runRun,
	Usage: "run [-a app] command [arguments]",
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
	return c
}

func tput(what string) string {
	c := exec.Command("tput", what)
	c.Stdin = os.Stdin
	out, err := c.Output()
	if err != nil {
		log.Fatal(err)
	}
	return strings.TrimSpace(string(out))
}

func runRun(cmd *Command, args []string) {
	data := make(url.Values)
	if !detachedRun {
		data.Add("attach", "true")
		data.Add("ps_env[TERM]", os.Getenv("TERM"))
		data.Add("ps_env[COLUMNS]", tput("cols"))
		data.Add("ps_env[LINES]", tput("lines"))
	}
	data.Add("command", strings.Join(args, " "))

	resp := struct {
		Url *string `json:"rendezvous_url,omitempty"`
	}{}

	r := APIReq("POST", "/apps/"+mustApp()+"/ps")
	r.SetBodyForm(data)
	r.Do(&resp)

	if detachedRun {
		return
	}

	u, err := url.Parse(*resp.Url)
	if err != nil {
		log.Fatal(err)
	}

	cn, err := tls.Dial("tcp", u.Host, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer cn.Close()

	br := bufio.NewReader(cn)

	_, err = io.WriteString(cn, u.Path[1:]+"\r\n")
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

	if isTerminal(os.Stdin) && isTerminal(os.Stdout) {
		err = stty("-icanon", "-echo").Run()
		if err != nil {
			log.Fatal(err)
		}
		defer stty("icanon", "echo").Run()

		sig := make(chan os.Signal)
		signal.Notify(sig, os.Signal(syscall.SIGQUIT), os.Interrupt)
		go func() {
			defer stty("icanon", "echo").Run()
			for sg := range sig {
				switch sg {
				case os.Interrupt:
					cn.Write([]byte{3})
				case os.Signal(syscall.SIGQUIT):
					cn.Write([]byte{28})
				default:
					panic("not reached")
				}
			}
		}()
	}

	errc := make(chan error)
	cp := func(a io.Writer, b io.Reader) {
		_, err := io.Copy(a, b)
		errc <- err
	}

	go cp(os.Stdout, br)
	go cp(cn, os.Stdin)
	if err = <-errc; err != nil {
		log.Fatal(err)
	}
}
