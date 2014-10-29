package cli

type TopicSet map[string]*Topic

type Topic struct {
	Name      string
	ShortHelp string
	Help      string
	Commands  []*Command
}

type Command struct {
	Name      string
	ShortHelp string
	Help      string
	NeedsApp  bool
	NeedsAuth bool
	Run       func(ctx *Context)
}

func (t *Topic) String() string {
	return t.Name
}

func (c *Command) String() string {
	return c.Name
}

func NewTopicSet(topics ...*Topic) TopicSet {
	set := TopicSet{}
	for _, topic := range topics {
		set.AddTopic(topic)
	}
	return set
}

func (topics TopicSet) AddTopic(topic *Topic) {
	if topics[topic.Name] == nil {
		topics[topic.Name] = topic
		return
	}
	dest := topics[topic.Name]
	for _, cmd := range topic.Commands {
		if dest.GetCommand(cmd.Name) == nil {
			dest.Commands = append(dest.Commands, cmd)
		}
	}
}

func (t *Topic) GetCommand(name string) (command *Command) {
	for _, command := range t.Commands {
		if name == command.Name {
			return command
		}
	}
	return nil
}
