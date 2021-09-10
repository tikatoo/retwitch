package retwitch

import (
	"github.com/lrstanley/girc"
)

type Client struct {
	irc  *girc.Client
	levs chan LiveEvent
}

func (c *Client) Join(channel string) (err error) {
	channel = "#" + channel
	evch := c.waitIRC(
		channel,
		girc.JOIN,
		girc.ERR_BANNEDFROMCHAN,
		girc.ERR_INVITEONLYCHAN,
		girc.ERR_BADCHANNELKEY,
		girc.ERR_CHANNELISFULL,
		girc.ERR_BADCHANMASK,
		girc.ERR_NOSUCHCHANNEL,
		girc.ERR_TOOMANYCHANNELS,
		girc.ERR_TOOMANYTARGETS,
		girc.ERR_UNAVAILRESOURCE,
	)

	c.irc.Cmd.Join(channel)
	result := <-evch
	if result.Command == girc.JOIN {
		return nil
	}

	return &girc.ErrEvent{Event: &result}
}

func (c *Client) LiveEvents() (events <-chan LiveEvent) {
	return c.levs
}
