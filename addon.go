package main

import (
	"log"
	"sort"
)

type mergedAddon struct {
	Type      string
	Name      string
	Owner     string
	ConfigVar string
}

func (m *mergedAddon) String() string {
	if m.ConfigVar == "" {
		return "(" + m.Type + ")"
	}
	return m.ConfigVar
}

func getMergedAddons(appname string) []*mergedAddon {
	var v []*Attachment
	var res []*Resource
	app := new(App)
	app.Name = mustApp()
	ch := make(chan error)
	go func() { ch <- Get(&v2{&res}, "/apps/"+app.Name+"/addons") }()
	go func() { ch <- Get(&v2{&v}, "/apps/"+app.Name+"/attachments") }()
	go func() { ch <- Get(app, "/apps/"+app.Name) }()
	if err := <-ch; err != nil {
		log.Fatal(err)
	}
	if err := <-ch; err != nil {
		log.Fatal(err)
	}
	if err := <-ch; err != nil {
		log.Fatal(err)
	}
	return mergeAddons(app, res, v)
}

func mergeAddons(app *App, res []*Resource, att []*Attachment) (ms []*mergedAddon) {
	for _, a := range att {
		m := new(mergedAddon)
		ms = append(ms, m)
		m.Type = a.Resource.Type
		m.Name = a.Resource.Name
		m.ConfigVar = a.ConfigVar
		m.Owner = a.Resource.BillingApp.Owner
	}

	for _, r := range res {
		var m *mergedAddon
		for _, ex := range ms {
			if ex.Type == r.Name {
				m = ex
				break
			}
		}
		if m == nil {
			m = new(mergedAddon)
			ms = append(ms, m)
		}
		m.Type = r.Name
		if m.Owner == "" {
			m.Owner = app.Owner.Email
		}
	}
	sort.Sort(mergedAddonsByType(ms))
	return ms
}

type mergedAddonsByType []*mergedAddon

func (a mergedAddonsByType) Len() int           { return len(a) }
func (a mergedAddonsByType) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a mergedAddonsByType) Less(i, j int) bool { return a[i].Type < a[j].Type }
