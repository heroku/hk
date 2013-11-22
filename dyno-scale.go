package main

import (
	"github.com/bgentry/heroku-go"
	"log"
	"os"
	"strconv"
	"strings"
)

var cmdDynoScale = &Command{
	Run:   runDynoScale,
	Usage: "scale <type>=<count> ...",
	Short: "change dyno counts",
	Long: `
Scale changes the number of dynos for each process type.

Example:

	$ hk dyno scale web=2 worker=5
`,
}

// takes args of the form "web=1", "worker=3", etc
func runDynoScale(cmd *Command, args []string) {
	todo := make(map[string]int)
	for _, arg := range args {
		i := strings.IndexRune(arg, '=')
		if i < 0 {
			cmd.printUsage()
			os.Exit(2)
		}
		val, err := strconv.Atoi(arg[i+1:])
		if err != nil {
			cmd.printUsage()
			os.Exit(2)
		}
		todo[arg[:i]] = val
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

func scale(app, ps string, n int, ch chan error) {
	_, err := client.FormationUpdate(app, ps, heroku.FormationUpdateOpts{Quantity: &n})
	ch <- err
}
