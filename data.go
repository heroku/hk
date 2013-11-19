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
