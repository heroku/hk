package main

import (
	"bufio"
	"crypto/tls"
	"io"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/heroku/hk/Godeps/_workspace/src/github.com/bgentry/heroku-go"
	"github.com/heroku/hk/term"
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
Run a process on Heroku. Flags such as` + " `-a` " + `may be parsed out of
the command unless the command is quoted or provided after a
double-dash (--).

Options:

    -s <size>  set the size for this dyno (e.g. 2X)
    -d         run in detached mode instead of attached to terminal

Examples:

    $ hk run echo "hello"
    Running ` + "`echo \"hello\"`" + ` on myapp as run.1234:
    "hello"

    $ hk run console
    Running ` + "`console`" + ` on myapp as run.5678:
    Loading production environment (Rails 3.2.14)
    irb(main):001:0> ...

    $ hk run -d -s 2X bin/my_worker
    Ran ` + "`bin/my_worker`" + ` on myapp as run.4321, detached.

    $ hk run -a myapp -- ls -a /
    Running ` + "`ls -a bin /`" + ` on myapp as run.8650:
    /:
    .  ..  app  bin  dev  etc  home  lib  lib64  lost+found  proc  sbin  tmp  usr  var
`,
}

func init() {
	cmdRun.Flag.BoolVarP(&detachedRun, "detached", "d", false, "detached")
	cmdRun.Flag.StringVarP(&dynoSize, "size", "s", "", "dyno size")
}

func runRun(cmd *Command, args []string) {
	if len(args) == 0 {
		cmd.PrintUsage()
		os.Exit(2)
	}
	appname := mustApp()

	cols, err := term.Cols()
	if err != nil {
		printFatal(err.Error())
	}
	lines, err := term.Lines()
	if err != nil {
		printFatal(err.Error())
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
			cmd.PrintUsage()
			os.Exit(2)
		}
		opts.Size = &dynoSize
	}

	command := strings.Join(args, " ")
	dyno, err := client.DynoCreate(appname, command, &opts)
	must(err)

	if detachedRun {
		log.Printf("Ran `%s` on %s as %s, detached.", dyno.Command, appname, dyno.Name)
		return
	}
	log.Printf("Running `%s` on %s as %s:", dyno.Command, appname, dyno.Name)

	u, err := url.Parse(*dyno.AttachURL)
	if err != nil {
		printFatal(err.Error())
	}

	cn, err := tls.Dial("tcp", u.Host, nil)
	if err != nil {
		printFatal(err.Error())
	}
	defer cn.Close()

	br := bufio.NewReader(cn)

	_, err = io.WriteString(cn, u.Path[1:]+"\r\n")
	if err != nil {
		printFatal(err.Error())
	}

	for {
		_, pre, err := br.ReadLine()
		if err != nil {
			printFatal(err.Error())
		}
		if !pre {
			break
		}
	}

	if term.IsTerminal(os.Stdin) && term.IsTerminal(os.Stdout) {
		err = term.MakeRaw(os.Stdin)
		if err != nil {
			printFatal(err.Error())
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
		printFatal(err.Error())
	}
}
