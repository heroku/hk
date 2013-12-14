package main

import (
	"bufio"
	"crypto/tls"
	"github.com/bgentry/heroku-go"
	"github.com/heroku/hk/term"
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
	dynoSize    string
)

var cmdRun = &Command{
	Run:      runRun,
	Usage:    "run [-s <size>] [-d] <command> [<argument>...]",
	NeedsApp: true,
	Category: "dyno",
	Short:    "run a process in a dyno",
	Long: `
Run a process on Heroku

Options:

    -s <size>  set the size for this dyno (e.g. 2X)
    -d         run in detached mode instead of attached to terminal

Examples:

    $ hk run echo "hello"
    "hello"

    $ hk run -s 2X console
    Loading production environment (Rails 3.2.14)
    irb(main):001:0> ...

    $ hk run -d bin/my_worker
`,
}

func init() {
	cmdRun.Flag.BoolVar(&detachedRun, "d", false, "detached")
	cmdRun.Flag.StringVar(&dynoSize, "s", "", "dyno size")
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

	attached := !detachedRun
	opts := heroku.DynoCreateOpts{Attach: &attached}
	if attached {
		env := map[string]string{
			"COLUMNS": strconv.Itoa(cols),
			"LINES":   strconv.Itoa(lines),
			"TERM":    os.Getenv("TERM"),
		}
		opts.Env = &env
	}
	if dynoSize != "" {
		if !strings.HasSuffix(dynoSize, "X") {
			cmd.printUsage()
			os.Exit(2)
		}
		opts.Size = &dynoSize
	}

	dyno, err := client.DynoCreate(mustApp(), strings.Join(args, " "), opts)
	must(err)

	if detachedRun {
		return
	}

	u, err := url.Parse(*dyno.AttachURL)
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
