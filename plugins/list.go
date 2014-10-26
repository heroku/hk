package plugins

import (
	"encoding/json"
	"io"
	"os"
	"os/exec"

	"github.com/dickeyxxx/gonpm/cli"
)

func list() {
	cli.Logln("Listing plugins...")
	for _, plugin := range ListPlugins() {
		cli.Stdoutln(plugin.Name, plugin.Version)
	}
}

func ListPlugins() []*Plugin {
	cmd := exec.Command(npmPath, "list", "--json")
	stderr, err := cmd.StderrPipe()
	must(err)
	stdout, err := cmd.StdoutPipe()
	must(err)
	err = cmd.Start()
	go io.Copy(os.Stderr, stderr)
	must(err)
	var doc map[string]map[string]*Plugin
	err = json.NewDecoder(stdout).Decode(&doc)
	must(err)
	err = cmd.Wait()
	must(err)
	var plugins []*Plugin
	for name, p := range doc["dependencies"] {
		p.Topic = &cli.Topic{
			Name:      name,
			ShortHelp: pluginShortHelp(name),
			Run:       pluginRun(name),
			Help:      pluginHelp(name),
		}
		plugins = append(plugins, p)
	}
	return plugins
}
