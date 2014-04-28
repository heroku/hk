package speakeasy

import (
	"bufio"
	"fmt"
	"os"
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
	s := bufio.NewScanner(os.Stdin)
	s.Scan()
	// Carriage return after the user input.
	fmt.Println("")
	return s.Text(), s.Err()
}
