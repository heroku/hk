package plugins

import (
	"io"
	"os/exec"

	"github.com/dickeyxxx/gonpm/cli"
)

func install(name string) {
	cli.Stderrf("Installing plugin %s...\n", name)
	cmd := exec.Command(npmPath, "install", name)
	stderr, err := cmd.StderrPipe()
	must(err)
	stdout, err := cmd.StdoutPipe()
	must(err)
	err = cmd.Start()
	go io.Copy(cli.Stderr, stderr)
	go io.Copy(cli.Stdout, stdout)
	must(err)
	err = cmd.Wait()
	must(err)
}
