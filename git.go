package main

import (
	"bufio"
	"bytes"
	"errors"
	"os/exec"
	"strings"
	"syscall"
)

import "fmt"
import "os"

var _ = fmt.Print
var _ = os.Stdout

const (
	gitURLSuf = ".git"
)

func gitHost() string {
	if herokuGitHost := os.Getenv("HEROKU_GIT_HOST"); herokuGitHost != "" {
		return herokuGitHost
	}
	if herokuHost := os.Getenv("HEROKU_HOST"); herokuHost != "" {
		return herokuHost
	}
	return "heroku.com"
}

func gitURLPre() string {
	return "git@" + gitHost() + ":"
}

func gitDescribe(rels []*Release) error {
	args := []string{"name-rev", "--tags", "--no-undefined", "--always", "--"}
	for _, r := range rels {
		if isDeploy(r.Description) {
			r.Commit = r.Description[len(r.Description)-7:]
		}
		if r.Commit != "" {
			args = append(args, r.Commit)
		}
	}
	out, err := exec.Command("git", args...).Output()
	names := mapOutput(out, " ", "\n")
	for _, r := range rels {
		if name, ok := names[r.Commit]; ok {
			if strings.HasPrefix(name, "tags/") {
				name = name[5:]
			}
			if strings.HasSuffix(name, "^0") {
				name = name[:len(name)-2]
			}
			r.Commit = name
		}
	}
	return err
}

func isDeploy(s string) bool {
	return len(s) == len("Deploy 0000000") && strings.HasPrefix(s, "Deploy ")
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

func gitRemotes() (map[string]string, error) {
	b, err := exec.Command("git", "remote", "-v").Output()
	if err != nil {
		return nil, err
	}

	return parseGitRemoteOutput(b)
}

func appNameFromGitURL(remote string) string {
	if !strings.HasPrefix(remote, gitURLPre()) || !strings.HasSuffix(remote, gitURLSuf) {
		return ""
	}
	return remote[len(gitURLPre()) : len(remote)-len(gitURLSuf)]
}

func parseGitRemoteOutput(b []byte) (results map[string]string, err error) {
	s := bufio.NewScanner(bytes.NewBuffer(b))
	s.Split(bufio.ScanLines)

	results = make(map[string]string)

	for s.Scan() {
		by := s.Bytes()
		f := bytes.Fields(by)
		if len(f) != 3 || string(f[2]) != "(push)" {
			// this should have 3 tuples + be a push remote, skip it if not
			continue
		}

		if appName := appNameFromGitURL(string(f[1])); appName != "" {
			results[string(f[0])] = appName
		}
	}
	if err = s.Err(); err != nil {
		return nil, err
	}
	return
}

func remoteFromGitConfig() string {
	b, err := exec.Command("git", "config", "heroku.remote").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

var errMultipleHerokuRemotes = errors.New("multiple apps in git remotes")

func appFromGitRemote(remote string) (string, error) {
	if remote != "" {
		b, err := exec.Command("git", "config", "remote."+remote+".url").Output()
		if err != nil {
			if isNotFound(err) {
				wdir, _ := os.Getwd()
				return "", fmt.Errorf("could not find git remote "+remote+" in %s", wdir)
			}
			return "", err
		}

		out := strings.TrimSpace(string(b))

		appName := appNameFromGitURL(out)
		if appName == "" {
			return "", fmt.Errorf("could not find app name in " + remote + " git remote")
		}
		return appName, nil
	}

	// no remote specified, see if there is a single Heroku app remote
	remotes, err := gitRemotes()
	if err != nil {
		return "", nil // hide this error
	}
	if len(remotes) > 1 {
		return "", errMultipleHerokuRemotes
	}
	for _, v := range remotes {
		return v, nil
	}
	return "", fmt.Errorf("no apps in git remotes")
}

func isNotFound(err error) bool {
	if ee, ok := err.(*exec.ExitError); ok {
		if ws, ok := ee.ProcessState.Sys().(syscall.WaitStatus); ok {
			return ws.ExitStatus() == 1
		}
	}
	return false
}
