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

var cmdSSHKeyAdd = &Command{
	Run:      runSSHKeyAdd,
	Usage:    "sshkey-add [<public-key-file>]",
	Category: "account",
	Short:    "add ssh public key",
	Long: `
Command sshkey-add adds an ssh public key to your Heroku account.

It tries these sources for keys, in order:

1. public-key-file argument, if present
2. output of ssh-add -L, if any
3. file $HOME/.ssh/id_rsa.pub
`,
}

func runSSHKeyAdd(cmd *Command, args []string) {
	if len(args) > 1 {
		cmd.printUsage()
		os.Exit(2)
	}
	if len(args) == 1 {
		sshPubKeyPath = args[0]
	}
	keys, err := findSSHKeys()
	if err != nil {
		if _, ok := err.(privKeyError); ok {
			log.Println("refusing to upload")
		}
		printError(err.Error())
	}

	key, err := client.KeyCreate(string(keys))
	must(err)
	log.Printf("Key %s for %s added.", abbrev(key.Fingerprint, 15), key.Email)
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

	key, err := sshReadPubKey(filepath.Join(homePath(), ".ssh", "id_rsa.pub"))
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
