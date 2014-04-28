package speakeasy

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Ask the user to enter a password with input hidden. prompt is a string to
// display before the user's input. Returns the provided password, or an error
// if the command failed.
func Ask(prompt string) (password string, err error) {
	if prompt != "" {
		fmt.Fprint(os.Stdout, prompt) // Display the prompt.
	}
	return getPassword()
}

func readline() (value string, err error) {
	var pw []byte
	b := make([]byte, 1)
	stdin := bufio.NewReader(os.Stdin)
	for {
		// read one byte at a time so we don't accidentally read extra bytes
		_, err = stdin.Read(b)
		if err != nil {
			return
		}
		if b[0] == '\n' {
			break
		}
		pw = append(pw, b[0])
	}

	// Carriage return after the user input.
	fmt.Println("")
	return strings.TrimSuffix(string(pw), "\r"), nil
}
