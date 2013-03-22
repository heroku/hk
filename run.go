package main

import (
	"bufio"
	"crypto/tls"
	"github.com/kr/hk/term"
	"io"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strconv"
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

func runRun(cmd *Command, args []string) {
	cols, err := term.Cols()
	if err != nil {
		log.Fatal(err)
	}
	lines, err := term.Lines()
	if err != nil {
		log.Fatal(err)
	}
	data := make(url.Values)
	if !detachedRun {
		data.Add("attach", "true")
		data.Add("ps_env[TERM]", os.Getenv("TERM"))
		data.Add("ps_env[COLUMNS]", strconv.Itoa(cols))
		data.Add("ps_env[LINES]", strconv.Itoa(lines))
	}
	data.Add("command", strings.Join(args, " "))

	resp := struct {
		Url *string `json:"rendezvous_url,omitempty"`
	}{}

	must(Post(&v2{&resp}, "/apps/"+mustApp()+"/ps", data))

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

	if term.IsTerminal(os.Stdin) && term.IsTerminal(os.Stdout) {
		err = term.MakeRaw(os.Stdin)
		if err != nil {
			log.Fatal(err)
		}
		defer term.Restore(os.Stdin)

		sig := make(chan os.Signal)
		signal.Notify(sig, os.Signal(syscall.SIGQUIT), os.Interrupt)
		go func() {
			defer term.Restore(os.Stdin)
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
