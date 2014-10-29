package cli

import "flag"

type Context struct {
	App   string        `json:"app"`
	Args  []string      `json:"args"`
	Flags *flag.FlagSet `json:"flags"`
	Auth  struct {
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"auth"`
}
