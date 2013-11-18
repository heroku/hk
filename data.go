package main

import (
	"github.com/bgentry/heroku-go"
	"time"
)

type Release struct {
	heroku.Release

	Commit string // deduced from Description, if possible
	Who    string // who created the release
}

type Addon struct {
	Id   string
	Plan struct {
		Id   string
		Name string
	}
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
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
		Id    string
		Name  string
	}
	Resource struct {
		Name       string
		Type       string
		Id         string
		Value      string
		SSOURL     *NullString `json:"sso_url"`
		BillingApp struct {
			Name  string
			Id    string
			Owner string
		} `json:"billing_app"`
	}
}

type LogSession struct {
	LogplexURL string `json:"logplex_url"`
	CreatedAt  time.Time
}

type NullString string

func (s *NullString) String() string {
	if s == nil {
		return "(null)"
	}
	return string(*s)
}
