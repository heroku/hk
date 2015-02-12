package postgresql

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/heroku/hk/Godeps/_workspace/src/github.com/bgentry/testnet"
)

// Tests

func TestNewRequestURL(t *testing.T) {
	c := &Client{}
	req, err := c.NewRequest(true, "GET", "/")
	if err != nil {
		t.Error(err)
	} else {
		if req.Host != "postgres-starter-api.heroku.com" {
			t.Errorf("expected starter req.Host=%s, got %s", "postgres-starter-api.heroku.com", req.Host)
		}
		if req.URL.Path != DefaultAPIPath+"/" {
			t.Errorf("expected starter req.Path=%s, got %s", DefaultAPIPath+"/", req.URL.Path)
		}
	}
	req, err = c.NewRequest(false, "GET", "/")
	if err != nil {
		t.Error(err)
	} else {
		if req.Host != "postgres-api.heroku.com" {
			t.Errorf("expected non-starter req.Host=%s, got %s", "postgres-api.heroku.com", req.Host)
		}
		if req.URL.Path != DefaultAPIPath+"/" {
			t.Errorf("expected non-starter req.Path=%s, got %s", DefaultAPIPath+"/", req.URL.Path)
		}
	}

	// Test with an overridden Client.URL
	c.URL = "https://myfakeurl.com/omg"
	c.StarterURL = ""
	req, err = c.NewRequest(false, "GET", "/")
	if err != nil {
		t.Error(err)
	} else if req.URL.String() != c.URL+"/" {
		t.Errorf("expected overridden non-starter req.URL=%s, got %s", c.URL+"/", req.URL.String())
	}
	req, err = c.NewRequest(true, "GET", "/")
	if err != nil {
		t.Error(err)
	} else if req.URL.String() != DefaultStarterAPIURL+"/" {
		t.Errorf("expected default starter req.URL=%s, got %s", DefaultStarterAPIURL+"/", req.URL.String())
	}

	// Test with an overridden Client.StarterURL
	c.URL = ""
	c.StarterURL = "https://myfakeurl.com/omg"
	req, err = c.NewRequest(false, "GET", "/")
	if err != nil {
		t.Error(err)
	} else if req.URL.String() != DefaultAPIURL+"/" {
		t.Errorf("expected default non-starter req.URL=%s, got %s", DefaultAPIURL+"/", req.URL.String())
	}
	req, err = c.NewRequest(true, "GET", "/")
	if err != nil {
		t.Error(err)
	} else if req.URL.String() != c.StarterURL+"/" {
		t.Errorf("expected overridden starter req.URL=%s, got %s", c.StarterURL+"/", req.URL.String())
	}
}

func TestAdditionalHeaders(t *testing.T) {
	multival := []string{"awesome", "multival"}
	c := &Client{AdditionalHeaders: http.Header{
		"Fake-Header":     []string{"value"},
		"X-Heroku-Header": multival,
	}}
	req, err := c.NewRequest(false, "GET", "/")
	if err != nil {
		t.Fatal(err)
	}
	if val := req.Header.Get("Fake-Header"); val != "value" {
		t.Errorf("Fake-Header expected %q, got %q", "value", val)
	}
	val := req.Header["X-Heroku-Header"]
	if len(val) != len(multival) {
		t.Errorf("X-Heroku-Header len expected %d, got %d", len(multival), len(val))
	}
	for i, v := range val {
		if v != multival[i] {
			t.Errorf("X-Heroku-Header value[%d] expected %q, got %q", i, multival[i], v)
		}
	}
}

func TestHerokuAgent(t *testing.T) {
	c := &Client{
		HerokuAgentSocket: "~/.heroku-agent.sock",
	}
	req, err := c.NewRequest(false, "GET", "/")
	if err != nil {
		t.Fatal(err)
	}
	if req.URL.Scheme != "http" {
		t.Error("Expected http scheme, got %s", req.URL.Scheme)
	}
}

func TestUserAgent(t *testing.T) {
	c := &Client{}
	req, err := c.NewRequest(false, "GET", "/")
	if err != nil {
		t.Fatal(err)
	}
	if ua := req.Header.Get("User-Agent"); ua != DefaultUserAgent {
		t.Errorf("Default User-Agent expected %q, got %q", DefaultUserAgent, ua)
	}

	// try a custom User-Agent
	customAgent := "custom-client 2.1 " + DefaultUserAgent
	c.UserAgent = customAgent
	req, err = c.NewRequest(false, "GET", "/")
	if err != nil {
		t.Fatal(err)
	}
	if ua := req.Header.Get("User-Agent"); ua != customAgent {
		t.Errorf("User-Agent expected %q, got %q", customAgent, ua)
	}
}

func TestGet(t *testing.T) {
	resp := testnet.TestResponse{
		Status: http.StatusOK,
		Body:   `{"omg": "wtf"}`,
	}
	req := newTestRequest("GET", "/", resp)

	ts, _, c := newTestServerAndClient(t, req)
	defer ts.Close()

	var respBody struct {
		Omg string
	}
	err := c.Get(false, "/", &respBody)
	if err != nil {
		t.Fatal(err)
	}
	if respBody.Omg != "wtf" {
		t.Errorf("expected %q, got %q", "wtf", respBody.Omg)
	}
}

func TestPost(t *testing.T) {
	resp := testnet.TestResponse{
		Status: http.StatusOK,
		Body:   `{"omg": "wtf"}`,
	}
	req := newTestRequest("POST", "/", resp)

	ts, _, c := newTestServerAndClient(t, req)
	defer ts.Close()

	var reqBody struct {
		Wtf string
	}
	reqBody.Wtf = "bbq"
	var respBody struct {
		Omg string
	}
	err := c.Post(false, "/", &respBody)
	if err != nil {
		t.Fatal(err)
	}
	if respBody.Omg != "wtf" {
		t.Errorf("expected %q, got %q", "wtf", respBody.Omg)
	}
}

type respTest struct {
	Response http.Response
	Expected error
}

func newTestResponse(statuscode int, body string) http.Response {
	return http.Response{
		StatusCode:    statuscode,
		Status:        http.StatusText(statuscode),
		ContentLength: int64(len(body)),
		Body:          ioutil.NopCloser(bytes.NewBufferString(body)),
	}
}

var respTests = []respTest{
	{newTestResponse(200, `{"code": "OK"}`), nil},
	{newTestResponse(201, `{"code": "OK"}`), nil},
	{
		newTestResponse(403, `Access denied`),
		errors.New("unexpected status code=403 message=\"Access denied\""),
	},
	{
		newTestResponse(401, `Unauthorized`),
		errors.New("unexpected status code=401 message=\"Unauthorized\""),
	},
}

func TestCheckResp(t *testing.T) {
	for i, rt := range respTests {
		resp := checkResp(&rt.Response)
		if !reflect.DeepEqual(rt.Expected, resp) {
			t.Errorf("checkResp respTests[%d] expected %v, got %v", i, rt.Expected, resp)
		}
	}
}

func TestNewDB(t *testing.T) {
	testId := "resource123@heroku.com"
	c := &Client{}
	db := c.NewDB(testId, "heroku-postgresql:dev")
	if db.client != c {
		t.Error("DB.client not set")
	}
	if db.Id != testId {
		t.Errorf("DB.Id expected %s, got %s", testId, db.Id)
	}
	if db.Plan != "dev" {
		t.Errorf("DB.Plan expected %s, got %s", "dev", db.Plan)
	}
}

// test helpers

func newTestRequest(method, path string, resp testnet.TestResponse) testnet.TestRequest {
	headers := http.Header{}
	headers.Set("Accept", "application/json")
	req := testnet.TestRequest{
		Method:   method,
		Path:     path,
		Response: resp,
		Header:   headers,
	}
	return req
}

func newTestServerAndClient(t *testing.T, requests ...testnet.TestRequest) (*httptest.Server, *testnet.Handler, *Client) {
	ts, handler := testnet.NewServer(t, requests)
	c := &Client{}
	c.URL = ts.URL

	return ts, handler, c
}

func testStringsEqual(t *testing.T, fieldName, expected, actual string) {
	if actual != expected {
		t.Errorf("%s expected %s, got %s", fieldName, expected, actual)
	}
}
