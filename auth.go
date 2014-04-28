package main

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/bgentry/heroku-go"
	"github.com/bgentry/speakeasy"
	"github.com/heroku/hk/hkclient"
	"github.com/heroku/hk/term"
)

var cmdCreds = &Command{
	Run:      runCreds,
	Usage:    "creds",
	Category: "hk",
	Short:    "show credentials" + extra,
	Long:     `Creds shows credentials that will be used for API calls.`,
}

func runCreds(cmd *Command, args []string) {
	var err error

	nrc, err = hkclient.LoadNetRc()
	if err != nil {
		printFatal(err.Error())
	}

	u, err := url.Parse(apiURL)
	if err != nil {
		printFatal("could not parse API url: " + err.Error())
	}

	user, pass, err := nrc.GetCreds(u)
	if err != nil {
		printFatal("could not get credentials: " + err.Error())
	}

	fmt.Println(user, pass)
}

var cmdLogin = &Command{
	Run:      runLogin,
	Usage:    "login",
	Category: "hk",
	Short:    "log in to your Heroku account" + extra,
	Long: `
Log in with your Heroku credentials. Input is accepted by typing
on the terminal. On unix machines, you can also pipe a password
on standard input.

Example:

    $ hk login
    Enter email: user@test.com
    Enter password: 
    Login successful.
`,
}

func runLogin(cmd *Command, args []string) {
	if len(args) != 0 {
		cmd.PrintUsage()
		os.Exit(2)
	}

	oldEmail := client.Username
	var email string
	if oldEmail == "" {
		fmt.Printf("Enter email: ")
	} else {
		fmt.Printf("Enter email [%s]: ", oldEmail)
	}
	_, err := fmt.Scanln(&email)
	switch {
	case err != nil && err.Error() != "unexpected newline":
		printFatal(err.Error())
	case email == "" && oldEmail == "":
		printFatal("email is required.")
	case email == "":
		email = oldEmail
	}

	// NOTE: gopass doesn't support multi-byte chars on Windows
	password, err := readPassword("Enter password: ")
	switch {
	case err == nil:
	case err.Error() == "unexpected newline":
		printFatal("password is required.")
	default:
		printFatal(err.Error())
	}

	hostname, token, err := attemptLogin(email, password, "")
	if err != nil {
		if herror, ok := err.(heroku.Error); ok && herror.Id == "two_factor" {
			// 2FA requested, attempt 2FA login
			var twoFactorCode string
			fmt.Printf("Enter two-factor auth code: ")
			if _, err := fmt.Scanln(&twoFactorCode); err != nil {
				printFatal("reading two-factor auth code: " + err.Error())
			}
			hostname, token, err = attemptLogin(email, password, twoFactorCode)
			must(err)
		} else {
			must(err)
		}
	}

	nrc, err = hkclient.LoadNetRc()
	if err != nil {
		printFatal("loading netrc: " + err.Error())
	}

	err = nrc.SaveCreds(hostname, email, token)
	if err != nil {
		printFatal("saving new token: " + err.Error())
	}
	fmt.Println("Logged in.")
}

func readPassword(prompt string) (password string, err error) {
	if acceptPasswordFromStdin && !term.IsTerminal(os.Stdin) {
		_, err = fmt.Scanln(&password)
		return
	}
	// NOTE: speakeasy may not support multi-byte chars on Windows
	return speakeasy.Ask("Enter password: ")
}

func attemptLogin(username, password, twoFactorCode string) (hostname, token string, err error) {
	description := "hk login from " + time.Now().UTC().Format(time.RFC3339)
	expires := 2592000 // 30 days
	opts := heroku.OAuthAuthorizationCreateOpts{
		Description: &description,
		ExpiresIn:   &expires,
	}

	req, err := client.NewRequest("POST", "/oauth/authorizations", &opts)
	if err != nil {
		return "", "", fmt.Errorf("unknown error when creating login request: %s", err.Error())
	}
	req.SetBasicAuth(username, password)

	if twoFactorCode != "" {
		req.Header.Set("Heroku-Two-Factor-Code", twoFactorCode)
	}

	var auth heroku.OAuthAuthorization
	if err = client.DoReq(req, &auth); err != nil {
		return
	}
	if auth.AccessToken == nil {
		return "", "", fmt.Errorf("access token missing from Heroku API login response")
	}
	return strings.Split(req.Host, ":")[0], auth.AccessToken.Token, nil
}

var cmdLogout = &Command{
	Run:      runLogout,
	Usage:    "logout",
	Category: "hk",
	Short:    "log out of your Heroku account" + extra,
	Long: `
Log out of your Heroku account and remove credentials from
this machine.

Example:

    $ hk logout
    Logged out.
`,
}

func runLogout(cmd *Command, args []string) {
	if len(args) != 0 {
		cmd.PrintUsage()
		os.Exit(2)
	}
	u, err := url.Parse(client.URL)
	if err != nil {
		printFatal("couldn't parse client URL: " + err.Error())
	}

	nrc, err = hkclient.LoadNetRc()
	if err != nil {
		printError(err.Error())
	}

	err = removeCreds(strings.Split(u.Host, ":")[0])
	if err != nil {
		printFatal("saving new netrc: " + err.Error())
	}
	fmt.Println("Logged out.")
}
