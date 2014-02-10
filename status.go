package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

var cmdStatus = &Command{
	Run:      runStatus,
	Usage:    "status",
	Category: "misc",
	Short:    "display heroku platform status" + extra,
	Long: `
Displays the current status of the Heroku platform.

Examples:

    $ hk status
    Production:   No known issues at this time.
    Development:  No known issues at this time.
`,
}

type statusResponse struct {
	Status struct {
		Production  string
		Development string
	} `json:"status"`
	Issues []statusIssue
}

type statusIssue struct {
	Resolved   bool   `json:"resolved"`
	StatusDev  string `json:"status_dev"`
	StatusProd string `json:"status_prod"`
	Title      string `json:"title"`
	Upcoming   bool   `json:"upcoming"`
	Href       string `json:"href"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func runStatus(cmd *Command, args []string) {
	if len(args) != 0 {
		cmd.printUsage()
		os.Exit(2)
	}
	herokuStatusHost := "status.heroku.com"
	if e := os.Getenv("HEROKU_STATUS_HOST"); e != "" {
		herokuStatusHost = e
	}
	res, err := http.Get("https://" + herokuStatusHost + "/api/v3/current-status.json")
	must(err)
	if res.StatusCode/100 != 2 { // 200, 201, 202, etc
		printFatal("unexpected HTTP status: %d", res.StatusCode)
	}

	var sr statusResponse
	err = json.NewDecoder(res.Body).Decode(&sr)
	must(err)

	fmt.Println("Production:  ", statusValueFromColor(sr.Status.Production))
	fmt.Println("Development: ", statusValueFromColor(sr.Status.Development))
}

func statusValueFromColor(color string) string {
	if color == "green" {
		return "No known issues at this time."
	}
	return color
}
