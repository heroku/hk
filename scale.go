package main

import (
	"net/url"
	"os"
	"strings"
	"sync"
)

var cmdScale = &Command{
	Run:   runScale,
	Usage: "scale [-a APP] type=n...",
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

	var wg sync.WaitGroup
	for ps, n := range todo {
		wg.Add(1)
		go scale(mustApp(), ps, n, &wg)
	}
	wg.Wait()
}

func scale(app, ps, n string, wg *sync.WaitGroup) {
	v := make(url.Values)
	v.Add("type", ps)
	v.Add("qty", n)
	req := APIReq("POST", "/apps/"+app+"/ps/scale")
	req.SetBodyForm(v)
	req.Do(nil) // TODO make non-2xx response non-fatal
	wg.Done()
}
