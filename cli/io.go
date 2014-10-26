package cli

import (
	"fmt"
	"os"
	"strings"
)

func Exit(code int) {
	exitFn(code)
}

func Stderrf(format string, a ...interface{}) {
	logger.Printf(format, a...)
	fmt.Fprintf(Stderr, format, a...)
}

func Stderrln(a ...interface{}) {
	logger.Println(a...)
	fmt.Fprintln(Stderr, a...)
}

func Stdoutf(format string, a ...interface{}) {
	logger.Printf(format, a...)
	fmt.Fprintf(Stdout, format, a...)
}

func Stdoutln(a ...interface{}) {
	logger.Println(a...)
	fmt.Fprintln(Stdout, a...)
}

func Logln(a ...interface{}) {
	logger.Println(a...)
	if debugging {
		fmt.Fprintln(Stderr, a...)
	}
}

func Logf(format string, a ...interface{}) {
	logger.Printf(format, a...)
	if debugging {
		fmt.Fprintf(Stderr, format, a...)
	}
}

var debugging = isDebugging()

func isDebugging() bool {
	debug := strings.ToUpper(os.Getenv("DEBUG"))
	if debug == "TRUE" || debug == "1" {
		return true
	}
	return false
}
