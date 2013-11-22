package main

var nsDyno = &Namespace{
	Name: "dyno",
	Commands: []*Command{
		cmdDynoList,
		cmdDynoRun,
		cmdDynoScale,
		cmdDynoRestart,
	},
	Short: "manage dynos",
}
