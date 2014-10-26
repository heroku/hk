package plugins

import (
	"os"
	"os/exec"
	"strings"

	"github.com/dickeyxxx/gonpm/cli"
)

type Plugin struct {
	*cli.Topic
	Version  string `json:"version"`
	From     string `json:"from"`
	Resolved string `json:"resolved"`
}

func pluginRun(name string) func(command string, args ...string) {
	return func(command string, args ...string) {
		context := `{
			"app": "dickey-xxx",
			"token": "` + os.Getenv("HEROKU_API_KEY") + `"
		}`
		runNode(`require('` + name + `').run("` + command + `", [], {}, ` + context + `)`)
	}
}

func pluginShortHelp(name string) string {
	script := `console.log(require('` + name + `').shortHelp)`
	output, err := exec.Command(nodePath, "-e", script).Output()
	must(err)
	return strings.TrimSpace(string(output))
}

func pluginHelp(name string) func(command string, args ...string) {
	return func(command string, args ...string) {
		runNode(`require('` + name + `').help()`)
	}
}

func runNode(script string) {
	cmd := exec.Command(nodePath, "-e", script)
	cmd.Stdout = cli.Stdout
	cmd.Stderr = cli.Stderr
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}
