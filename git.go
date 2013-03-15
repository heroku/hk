package main

import (
	"bytes"
	"os/exec"
	"strings"
)

import "fmt"
import "os"

var _ = fmt.Print
var _ = os.Stdout

const (
	gitURLPre = "git@heroku.com:"
	gitURLSuf = ".git"
)

func gitURL(app string) string {
	return gitURLPre + app + gitURLSuf
}

func gitRemotes(url string) (names []string) {
	out, err := exec.Command("git", "remote", "-v").Output()
	if err != nil {
		return nil
	}
	for _, line := range bytes.Split(out, []byte{'\n'}) {
		if i := bytes.IndexByte(line, '\t'); i >= 0 {
			if bytes.HasPrefix(line[i+1:], []byte(url+" ")) {
				names = append(names, string(line[:i]))
			}
		}
	}
	return names
}

func gitDescribe(rels []*Release) error {
	args := []string{"name-rev", "--tags", "--no-undefined", "--always", "--"}
	for _, r := range rels {
		if r.Commit != nil {
			args = append(args, string(*r.Commit))
		}
	}
	out, err := exec.Command("git", args...).Output()
	names := mapOutput(out, " ", "\n")
	for _, r := range rels {
		if name, ok := names[GitRef(r.Commit)]; ok {
			if strings.HasPrefix(name, "tags/") {
				name = name[5:]
			}
			if strings.HasSuffix(name, "^0") {
				name = name[:len(name)-2]
			}
			*r.Commit = name
		}
	}
	return err
}

func mapOutput(out []byte, sep, term string) map[string]string {
	m := make(map[string]string)
	lines := strings.Split(string(out), term)
	for _, line := range lines[:len(lines)-1] { // omit trailing ""
		parts := strings.SplitN(line, sep, 2)
		if len(parts) == 2 {
			m[parts[0]] = parts[1]
		}
	}
	return m
}
