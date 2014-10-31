package cli

type Topic struct {
	Name      string
	ShortHelp string
	Help      string
	Hidden    bool
	Commands  []*Command
}

type Command struct {
	Name      string
	ShortHelp string
	Help      string
	Hidden    bool
	NeedsApp  bool
	NeedsAuth bool
	Args      []*Arg
	Flags     []*Flag
	Run       func(ctx *Context) `json:"-"`
}

type Arg struct {
	Name     string
	Optional bool
}

type Flag struct {
	Name    string
	Char    rune
	Default string
}

func (t *Topic) String() string {
	return t.Name
}

func (c *Command) String() string {
	return c.Name
}

func (t *Topic) GetCommand(name string) (command *Command) {
	for _, command := range t.Commands {
		if name == command.Name {
			return command
		}
	}
	return nil
}
