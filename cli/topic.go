package cli

type Topic struct {
	Name      string
	ShortHelp string
	Run       func(command string, args ...string)
	Help      func(command string, args ...string)
}
