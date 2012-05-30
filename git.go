package main

import (
	"bytes"
	"log"
	"os/exec"
)

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
		log.Fatal(err)
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
