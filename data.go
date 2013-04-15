package main

import (
	"encoding/json"
	"time"
)

// See https://github.com/heroku/api-doc#apps
type App struct {
	ID     string
	Name   string
	Stack  string
	WebURL string `json:"web_url"`
	GitURL string `json:"git_url"`
	Owner  struct {
		Id    string
		Email string
	}
	RepoSize    *int       `json:"repo_size"`
	SlugSize    *int       `json:"slug_size"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	ReleasedAt  *time.Time `json:"released_at"`
	Maintenance bool

	BuildpackProvidedDescription *NullString `json:"buildpack_provided_description"`
}

type Dyno struct {
	Name           string `json:"process"`
	ID             string
	UPID           *string
	Type           string
	Command        string
	AppName        string `json:"app_name"`
	Slug           string
	Action         string
	State          string
	PrettyState    string `json:"pretty_state"`
	Elapsed        int
	RendezvousURL  *string `json:"rendezvous_url"`
	Attached       *bool
	TransisionedAt V2Time `json:"transitioned_at"`
}

func (d *Dyno) Age() time.Duration {
	return time.Now().Sub(d.TransisionedAt.Time)
}

type Release struct {
	ID   string
	Name string
	User struct {
		ID    string
		Email string
	}
	Description string
	CreatedAt   time.Time `json:"created_at"`

	Who    string // same as User.Email or abbreviated
	Commit string // deduced from Description, if possible
}

type Resource struct {
	URL   *NullString
	Price struct {
		Unit  string
		Cents int
	}
	TermsOfService    bool   `json:"terms_of_service"`
	GroupDescription  string `json:"group_description"`
	Configured        bool
	State             string
	SSOURL            *NullString `json:"sso_url"`
	Attachable        bool
	Description       string
	ConsumesDynoHours bool
	Selective         bool
	Beta              bool
	Name              string
	Slug              string
}

type Attachment struct {
	ConfigVar string `json:"config_var"`
	App       struct {
		Owner string
		ID    string
		Name  string
	}
	Resource struct {
		Name       string
		Type       string
		ID         string
		Value      string
		SSOURL     *NullString `json:"sso_url"`
		BillingApp struct {
			Name  string
			ID    string
			Owner string
		} `json:"billing_app"`
	}
}

type LogSession struct {
	LogplexURL string `json:"logplex_url"`
	CreatedAt  time.Time
}

type v2 struct {
	v interface{}
}

// Called by APIReq. Causes Heroku to use its "v2" API.
func (v2) Accept() string {
	return "application/json"
}

func (v v2) UnmarshalJSON(p []byte) error {
	return json.Unmarshal(p, v.v)
}

var v2nil = &v2{new(interface{})}

type V2Time struct {
	time.Time
}

const V2TimeFormat = "2006/01/02 15:04:05 -0700"

func (t *V2Time) UnmarshalJSON(data []byte) (err error) {
	// Fractional seconds are handled implicitly by Parse.
	t.Time, err = time.Parse(`"`+V2TimeFormat+`"`, string(data))
	return
}

type NullString string

func (s *NullString) String() string {
	if s == nil {
		return "(null)"
	}
	return string(*s)
}
