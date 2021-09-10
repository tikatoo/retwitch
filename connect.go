package retwitch

import "github.com/lrstanley/girc"

type ClientConfig struct {
}

func NewClient(config ClientConfig) (c *Client, err error) {
	c = &Client{}

	c.irc, err = ircConnect("", "")
	if err != nil {
		c = nil
		return
	}

	c.levs = make(chan LiveEvent, 24)
	c.irc.Handlers.Add(girc.PRIVMSG, c.onPrivmsg)

	c.channels = map[string]*ChannelInfo{}

	return
}

func NewAnonymousClient() (client *Client, err error) {
	return NewClient(ClientConfig{})
}
