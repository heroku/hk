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
	"text/tabwriter"
)

var (
	sshPubKeyPath string
)

var cmdKeys = &Command{
	Run:      runKeys,
	Usage:    "keys",
	Category: "account",
	Short:    "list ssh public keys" + extra,
	Long: `
Keys lists SSH public keys associated with your Heroku account.

Examples:

    $ hk keys
    5e:67:40:b6:79:db:56:47:cd:3a:a7:65:ab:ed:12:34  user@test.com
`,
}

func runKeys(cmd *Command, args []string) {
	if len(args) != 0 {
		cmd.printUsage()
		os.Exit(2)
	}

	keys, err := client.KeyList(nil)
	must(err)

	w := tabwriter.NewWriter(os.Stdout, 1, 2, 2, ' ', 0)
	defer w.Flush()

	for i := range keys {
		listRec(w,
			keys[i].Fingerprint,
			keys[i].Email,
		)
	}
}

var cmdKeyAdd = &Command{
	Run:      runKeyAdd,
	Usage:    "key-add [<public-key-file>]",
	Category: "account",
	Short:    "add ssh public key" + extra,
	Long: `
Command key-add adds an ssh public key to your Heroku account.

It tries these sources for keys, in order:

1. public-key-file argument, if present
2. output of ssh-add -L, if any
3. file $HOME/.ssh/id_rsa.pub
`,
}

func runKeyAdd(cmd *Command, args []string) {
	if len(args) > 1 {
		cmd.printUsage()
		os.Exit(2)
	}
	if len(args) == 1 {
		sshPubKeyPath = args[0]
	}
	keys, err := findKeys()
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

func findKeys() ([]byte, error) {
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

var cmdKeyRemove = &Command{
	Run:      runKeyRemove,
	Usage:    "key-remove <fingerprint>",
	Category: "account",
	Short:    "remove an ssh public key" + extra,
	Long: `
Command key-remove removes an ssh public key from your Heroku account.

Examples:

    $ hk key-remove 5e:67:40:b6:79:db:56:47:cd:3a:a7:65:ab:ed:12:34
    Key 5e:67:40:b6:79:db… removed.
`,
}

func runKeyRemove(cmd *Command, args []string) {
	if len(args) != 1 {
		cmd.printUsage()
		os.Exit(2)
	}
	fingerprint := args[0]

	err := client.KeyDelete(fingerprint)
	must(err)
	log.Printf("Key %s removed.", abbrev(fingerprint, 18))
}
