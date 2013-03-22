package main

import (
	"log"
	"net/url"
	"os"
	"strings"
)

var cmdScale = &Command{
	Run:   runScale,
	Usage: "scale [-a app] type=n...",
	Short: "change dyno counts",
	Long: `
Scale changes the number of dynos for each process type.

Example:

	$ hk scale web=2 worker=5
`,
}

func init() {
	cmdScale.Flag.StringVar(&flagApp, "a", "", "app")
}

// takes args of the form "web=1", "worker=3", etc
func runScale(cmd *Command, args []string) {
	todo := make(map[string]string)
	for _, arg := range args {
		i := strings.IndexRune(arg, '=')
		if i < 0 {
			cmd.printUsage()
			os.Exit(2)
		}
		todo[arg[:i]] = arg[i+1:]
	}

	ch := make(chan error)
	for ps, n := range todo {
		go scale(mustApp(), ps, n, ch)
	}
	for _ = range todo {
		if err := <-ch; err != nil {
			log.Println(err)
		}
	}
}

func scale(app, ps, n string, ch chan error) {
	v := make(url.Values)
	v.Add("type", ps)
	v.Add("qty", n)
	ch <- Post(v2nil, "/apps/"+app+"/ps/scale", v)
}
