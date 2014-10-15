package postgresql

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime"
	"strings"

	"code.google.com/p/go-uuid/uuid"
)

const (
	Version              = "0.0.1"
	DefaultAPIPath       = "/client/v11/databases"
	DefaultAPIURL        = "https://postgres-api.heroku.com" + DefaultAPIPath
	DefaultStarterAPIURL = "https://postgres-starter-api.heroku.com" + DefaultAPIPath
	DefaultUserAgent     = "heroku-postgres-go/" + Version + " (" + runtime.GOOS + "; " + runtime.GOARCH + ")"
)

// A Client is a Heroku Postgres API client. Its zero value is a usable client
// that uses default settings for the Heroku Postgres API. The Client has an
// internal HTTP client (HTTP) which defaults to http.DefaultClient.
//
// As with all http.Clients, this Client's Transport has internal state (cached
// HTTP connections), so Clients should be reused instead of created as needed.
// Clients are safe for use by multiple goroutines.
type Client struct {
	// HTTP is the Client's internal http.Client, handling HTTP requests to the
	// Heroku Postgres API.
	HTTP *http.Client

	// The URL of the Heroku Postgres API to communicate with. Defaults to
	// DefaultAPIURL.
	URL string

	// The URL of the Heroku Postgres Starter API to communicate with. Defaults
	// to DefaultStarterAPIURL.
	StarterURL string

	// Username is the HTTP basic auth username for API calls made by this Client.
	Username string

	// Password is the HTTP basic auth password for API calls made by this Client.
	Password string

	// UserAgent to be provided in API requests. Set to DefaultUserAgent if not
	// specified.
	UserAgent string

	// Debug mode can be used to dump the full request and response to stdout.
	Debug bool

	// AdditionalHeaders are extra headers to add to each HTTP request sent by
	// this Client.
	AdditionalHeaders http.Header

	// Path to the Unix domain socket or a running heroku-agent.
	HerokuAgentSocket string
}

func (c *Client) Get(isStarterPlan bool, path string, v interface{}) error {
	return c.APIReq(isStarterPlan, "GET", path, v)
}

func (c *Client) Post(isStarterPlan bool, path string, v interface{}) error {
	return c.APIReq(isStarterPlan, "POST", path, v)
}

func (c *Client) Put(isStarterPlan bool, path string, v interface{}) error {
	return c.APIReq(isStarterPlan, "PUT", path, v)
}

// Creates a new DB struct initialized with this Client.
func (c *Client) NewDB(id, plan string) DB {
	return DB{
		Id:     id,
		Plan:   strings.TrimLeft(plan, "heroku-postgresql:"),
		client: c,
	}
}

// Generates an HTTP request for the Heroku Postgres API, but does not
// perform the request. The request's Accept header field will be
// set to:
//
//   Accept: application/json
//
// The Request-Id header will be set to a random UUID. The User-Agent header
// will be set to the Client's UserAgent, or DefaultUserAgent if UserAgent is
// not set.
//
// isStarterPlan should be set to true if the target database is a starter plan,
// and false otherwise (as defined in DB.IsStarterPlan() ). Method is the HTTP
// method of this request, and path is the HTTP path.
func (c *Client) NewRequest(isStarterPlan bool, method, path string) (*http.Request, error) {
	var rbody io.Reader

	apiURL := strings.TrimRight(c.URL, "/")
	if isStarterPlan {
		apiURL = strings.TrimRight(c.StarterURL, "/")
	}
	if apiURL == "" {
		if isStarterPlan {
			apiURL = DefaultStarterAPIURL
		} else {
			apiURL = DefaultAPIURL
		}
	}
	req, err := http.NewRequest(method, apiURL+path, rbody)
	if err != nil {
		return nil, err
	}
	// If we're talking to heroku-agent over a local Unix socket, downgrade to
	// HTTP; heroku-agent will establish a secure connection between itself and
	// the Heorku API.
	if c.HerokuAgentSocket != "" {
		req.URL.Scheme = "http"
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Request-Id", uuid.New())
	useragent := c.UserAgent
	if useragent == "" {
		useragent = DefaultUserAgent
	}
	req.Header.Set("User-Agent", useragent)
	req.SetBasicAuth(c.Username, c.Password)
	for k, v := range c.AdditionalHeaders {
		req.Header[k] = v
	}
	return req, nil
}

// Sends a Heroku Postgres API request and decodes the response into v. As
// described in DoReq(), the type of v determines how to handle the response
// body.
func (c *Client) APIReq(isStarterPlan bool, meth, path string, v interface{}) error {
	req, err := c.NewRequest(isStarterPlan, meth, path)
	if err != nil {
		return err
	}
	return c.DoReq(req, v)
}

// Submits an HTTP request, checks its response, and deserializes
// the response into v. The type of v determines how to handle
// the response body:
//
//   nil        body is discarded
//   io.Writer  body is copied directly into v
//   else       body is decoded into v as json
//
func (c *Client) DoReq(req *http.Request, v interface{}) error {
	if c.Debug {
		dump, err := httputil.DumpRequestOut(req, true)
		if err != nil {
			log.Println(err)
		} else {
			os.Stderr.Write(dump)
			os.Stderr.Write([]byte{'\n', '\n'})
		}
	}

	httpClient := c.HTTP
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	res, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if c.Debug {
		dump, err := httputil.DumpResponse(res, true)
		if err != nil {
			log.Println(err)
		} else {
			os.Stderr.Write(dump)
			os.Stderr.Write([]byte{'\n'})
		}
	}
	if err = checkResp(res); err != nil {
		return err
	}
	switch t := v.(type) {
	case nil:
	case io.Writer:
		_, err = io.Copy(t, res.Body)
	default:
		err = json.NewDecoder(res.Body).Decode(v)
	}
	return err
}

func checkResp(res *http.Response) error {
	if res.StatusCode/100 != 2 { // 200, 201, 202, etc
		errb, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return fmt.Errorf("unexpected error code=%d", res.StatusCode)
		}
		return fmt.Errorf("unexpected status code=%d message=%q", res.StatusCode, string(errb))
	}
	return nil
}

//   @headers = { :x_heroku_gem_version  => Heroku::Client.version }
