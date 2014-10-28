package cli

type Context struct {
	App   string            `json:"app"`
	Args  []string          `json:"args"`
	Flags map[string]string `json:"flags"`
	Auth  struct {
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"auth"`
}
