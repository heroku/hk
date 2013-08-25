package main

import (
	"log"
	"sort"
)

type mergedAddon struct {
	Type  string
	Owner string
	ID    string
}

func (m *mergedAddon) String() string {
	return m.Type
}

func getMergedAddons(appname string) []*mergedAddon {
	var addons []*Addon
	app := new(App)
	app.Name = mustApp()
	ch := make(chan error)
	go func() { ch <- Get(&addons, "/apps/"+app.Name+"/addons") }()
	go func() { ch <- Get(app, "/apps/"+app.Name) }()
	if err := <-ch; err != nil {
		log.Fatal(err)
	}
	if err := <-ch; err != nil {
		log.Fatal(err)
	}
	return mergeAddons(app, addons)
}

func mergeAddons(app *App, addons []*Addon) (ms []*mergedAddon) {
	// Type, Name, Owner
	for _, a := range addons {
		m := new(mergedAddon)
		ms = append(ms, m)
		m.Type = a.Plan.Name
		m.Owner = app.Owner.Email
		m.ID = a.ID
	}

	sort.Sort(mergedAddonsByType(ms))
	return ms
}

type mergedAddonsByType []*mergedAddon

func (a mergedAddonsByType) Len() int           { return len(a) }
func (a mergedAddonsByType) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a mergedAddonsByType) Less(i, j int) bool { return a[i].Type < a[j].Type }
