package main

import (
	"github.com/bgentry/heroku-go"
)

var rangeex = "Range: name; max=2, order=desc"

type ListRange struct {
	Field      string
	Max        int
	Descending bool
}

func (r *Range) ToHeaderValue() {

	var rangeex = "Range: name; max=2, order=desc"
}

func main() {
	appId := "example"
	newname := "example-renamed"

	var a heroku.App

	c := &heroku.Client{}
	opts := AppCreateOpts{Name: &newname}

	c.Post(&a, "/apps/"+appId, opts)
	c.Post("/apps/"+appId, opts, &a)

	c.CreateApp(newname)

	r := ListRange{Field: "name", Max: 10, Descending: true}
	c.AppList(&r)

	app := heroku.App{Name: "example"}
	a.Update(opts)
	a.Delete()
}
