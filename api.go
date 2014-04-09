package main

import (
	"io"
	"os"
	"strings"
)

var cmdAPI = &Command{
	Run:      runAPI,
	Usage:    "api <method> <path>",
	Category: "hk",
	Short:    "make a single API request" + extra,
	Long: `
The api command is a convenient but low-level way to send requests
to the Heroku API. It sends an HTTP request to the Heroku API
using the given method on the given path, using stdin unmodified
as the request body. It prints the response unmodified on stdout.
Method GET doesn't read or send a request body.

Method name input will be upcased, so both 'hk api GET /apps' and
'hk api get /apps' are valid commands.

As with any hk command, the behavior of hk api is controlled by
various environment variables. See 'hk help environ' for details.

Examples:

    $ hk api GET /apps/myapp | jq .
    {
      "name": "myapp",
      "id": "app123@heroku.com",
      "created_at": "2011-11-11T04:17:13-00:00",
      …
    }

    $ export HKHEADER
    $ HKHEADER='
    Content-Type: application/x-www-form-urlencoded
    Accept: application/json
    '
    $ printf 'type=web&qty=2' | hk api POST /apps/myapp/ps/scale
    2
`,
}

func runAPI(cmd *Command, args []string) {
	if len(args) != 2 {
		cmd.printUsage()
		os.Exit(2)
	}
	method := strings.ToUpper(args[0])
	var body io.Reader
	if method == "PATCH" || method == "PUT" || method == "POST" {
		body = os.Stdin
	}
	if err := client.APIReq(os.Stdout, method, args[1], body); err != nil {
		printFatal(err.Error())
	}
}
