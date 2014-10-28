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
	},
}

var cmdApps = &cli.Command{
	Signature: "apps",
	ShortHelp: "lists your heroku apps",
	NeedsAuth: true,
	Run: func(ctx *cli.Context) {
		client := heroku.Client{Password: ctx.Auth.Password}
		apps, err := client.AppList(nil)
		must(err)
		ownedApps := filterApps(apps, func(app heroku.App) bool {
			return app.Owner.Email == ctx.Auth.Username
		})
		collaboratedApps := filterApps(apps, func(app heroku.App) bool {
			return app.Owner.Email != ctx.Auth.Username
		})
		cli.Println("=== My Apps")
		for _, app := range ownedApps {
			cli.Println(app.Name)
		}
		cli.Println()
		if len(collaboratedApps) > 0 {
			cli.Println("=== Collaborated Apps")
			for _, app := range collaboratedApps {
				cli.Printf("%-30s %s\n", app.Name, app.Owner.Email)
			}
			cli.Println()
		}
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

func must(err error) {
	if err != nil {
		panic(err)
	}
}
