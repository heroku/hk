package main

import (
	"bytes"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

var (
	sshPubKeyPath string
)

var cmdSSHAuth = &Command{
	Run:   runSSHAuth,
	Usage: "sshauth [-i identity-file]",
	Short: "authorize ssh public keys",
	Long: `
Command sshauth installs your ssh public keys for authorized use on Heroku.

It tries these sources for keys, in order:

1. flag -i, if present
2. output of ssh-add -L, if any
3. file $HOME/.ssh/id_rsa.pub
`,
}

func init() {
	cmdSSHAuth.Flag.StringVar(&sshPubKeyPath, "i", "", "ssh public key file")
}

func runSSHAuth(cmd *Command, args []string) {
	keys, err := findSSHKeys()
	if err != nil {
		if _, ok := err.(privKeyError); ok {
			log.Println("refusing to upload")
		}
		log.Fatal(err)
	}

	_, err = client.KeyCreate(string(keys))
	must(err)
}

func findSSHKeys() ([]byte, error) {
	if sshPubKeyPath != "" {
		return sshReadPubKey(sshPubKeyPath)
	}

	out, err := exec.Command("ssh-add", "-L").Output()
	if err != nil {
		return nil, err
	}
	if len(out) != 0 {
		print(string(out))
		return out, nil
	}

	key, err := sshReadPubKey(homePath + "/.ssh/id_rsa.pub")
	switch err {
	case syscall.ENOENT:
		return nil, errors.New("No SSH keys found")
	case nil:
		return key, nil
	}
	return nil, err
}

func sshReadPubKey(s string) ([]byte, error) {
	f, err := os.Open(filepath.FromSlash(s))
	if err != nil {
		return nil, err
	}

	key, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	if bytes.Contains(key, []byte("PRIVATE")) {
		return nil, privKeyError(s)
	}

	return key, nil
}

type privKeyError string

func (e privKeyError) Error() string {
	return "appears to be a private key: " + string(e)
}
