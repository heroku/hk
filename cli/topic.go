package cli

type TopicSet map[string]*Topic

type Topic struct {
	Name      string
	ShortHelp string
	Help      string
	Commands  map[string]*Command
}

type Command struct {
	Signature string
	ShortHelp string
	Help      string
	Run       func(args []string, flags map[string]string)
}

func (t *Topic) String() string {
	return t.Name
}

func (c *Command) String() string {
	return c.Signature
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
	for name, cmd := range topic.Commands {
		if dest.Commands[name] == nil {
			dest.Commands[name] = cmd
		}
	}
}
