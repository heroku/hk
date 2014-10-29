package apps

import (
	"github.com/bgentry/heroku-go"
	"github.com/heroku/hk/cli"
	"github.com/pivotal-golang/bytefmt"
)

var Info = &cli.Topic{
	Name:      "info",
	ShortHelp: "show detailed app information",
	Hidden:    true,
	Commands: []*cli.Command{
		{
			ShortHelp: "show detailed app information",
			NeedsApp:  true,
			NeedsAuth: true,
			Run:       info,
		},
	},
}

var cmdInfo = &cli.Command{
	Name:      "info",
	ShortHelp: "show detailed app information",
	NeedsApp:  true,
	NeedsAuth: true,
	Run:       info,
}

func info(ctx *cli.Context) {
	client := heroku.Client{Password: ctx.Auth.Password}
	cli.Printf("=== %s\n", ctx.App)
	addons := getAddons(client, ctx.App)
	collaborators := getCollaborators(client, ctx.App)
	app := getAppInfo(client, ctx.App)
	printAddons(<-addons)
	printCollaborators(<-collaborators)
	printApp(<-app)
}

func getAppInfo(client heroku.Client, app string) <-chan heroku.App {
	out := make(chan heroku.App)
	go func() {
		info, err := client.AppInfo(app)
		must(err)
		out <- *info
	}()
	return out
}

func getAddons(client heroku.Client, app string) <-chan []heroku.Addon {
	out := make(chan []heroku.Addon)
	go func() {
		addons, err := client.AddonList(app, nil)
		must(err)
		out <- addons
	}()
	return out
}

func getCollaborators(client heroku.Client, app string) <-chan []heroku.Collaborator {
	out := make(chan []heroku.Collaborator)
	go func() {
		collaborators, err := client.CollaboratorList(app, nil)
		must(err)
		out <- collaborators
	}()
	return out
}

func printAddons(addons []heroku.Addon) {
	if len(addons) == 0 {
		return
	}
	printItem("Addons:", addons[0].Plan.Name)
	for _, addon := range addons[1:] {
		printItem("", addon.Plan.Name)
	}
	cli.Println()
}

func printCollaborators(collaborators []heroku.Collaborator) {
	if len(collaborators) == 0 {
		return
	}
	printItem("Collaborators:", collaborators[0].User.Email)
	for _, collaborator := range collaborators[1:] {
		printItem("", collaborator.User.Email)
	}
	cli.Println()
}

func printApp(app heroku.App) {
	printItem("Git URL:", app.GitURL)
	printItem("Owner Email:", app.Owner.Email)
	printItem("Region:", app.Region.Name)
	if app.RepoSize != nil {
		printItem("Repo Size:", bytefmt.ByteSize(uint64(*app.RepoSize)))
	}
	if app.SlugSize != nil {
		printItem("Slug Size:", bytefmt.ByteSize(uint64(*app.SlugSize)))
	}
	printItem("Stack:", app.Stack.Name)
	printItem("Web URL:", app.WebURL)
}

func printItem(label, value string) {
	cli.Printf("%-14s %s\n", label, value)
}
