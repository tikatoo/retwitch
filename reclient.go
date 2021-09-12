package retwitch

import (
	"github.com/lrstanley/girc"
)

type Client struct {
	appAuth  *twitchauth
	helix    *HelixAPI
	irc      *girc.Client
	levs     chan LiveEvent
	channels map[string]*ChannelInfo // TODO: Memory leak
	badges   map[string]HelixChatBadge
}

func (c *Client) Helix() (*HelixAPI, error) {
	var err error

	if c.appAuth == nil {
		c.appAuth, err = getDefaultAuth()
		if err != nil {
			return nil, err
		}
	}

	if c.helix == nil {
		c.helix = getHelixAPI(c.appAuth)
	}

	return c.helix, nil
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
		// c.getChannelInfo(channel)
		return nil
	}

	return &girc.ErrEvent{Event: &result}
}

func (c *Client) LiveEvents() (events <-chan LiveEvent) {
	return c.levs
}
