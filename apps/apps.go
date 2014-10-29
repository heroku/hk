package apps

import (
	"github.com/bgentry/heroku-go"
	"github.com/heroku/hk/cli"
)

var Apps = &cli.Topic{
	Name:      "apps",
	ShortHelp: "manage your heroku apps",
	Commands: []*cli.Command{
		cmdApps,
		cmdInfo,
	},
}

var cmdApps = &cli.Command{
	ShortHelp: "lists your heroku apps",
	NeedsAuth: true,
	Run: func(ctx *cli.Context) {
		client := heroku.Client{Password: ctx.Auth.Password}
		apps, err := client.AppList(nil)
		must(err)
		owned := filterApps(apps, func(app heroku.App) bool {
			return app.Owner.Email == ctx.Auth.Username
		})
		collaborated := filterApps(apps, func(app heroku.App) bool {
			return app.Owner.Email != ctx.Auth.Username
		})
		printApps(owned, collaborated)
	},
}

func filterApps(from []heroku.App, fn func(heroku.App) bool) []heroku.App {
	to := make([]heroku.App, 0, len(from))
	for _, app := range from {
		if fn(app) {
			to = append(to, app)
		}
	}
	return to
}

func printApps(owned []heroku.App, collaborated []heroku.App) {
	cli.Println("=== My Apps")
	for _, app := range owned {
		cli.Println(app.Name)
	}
	cli.Println()
	if len(collaborated) > 0 {
		cli.Println("=== Collaborated Apps")
		for _, app := range collaborated {
			cli.Printf("%-30s %s\n", app.Name, app.Owner.Email)
		}
		cli.Println()
	}
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
