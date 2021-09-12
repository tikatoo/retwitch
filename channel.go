package retwitch

import (
	"regexp"
	"strings"
)

type ChannelInfo struct {
	Client     *Client
	Name       string
	id         string
	cheerMatch *regexp.Regexp
	cheerInfo  map[string]HelixCheermote
	badges     map[string]HelixChatBadge
}

func (c *Client) GetChannel(name string) (ch *ChannelInfo, err error) {
	if cch, cached := c.channels[name]; cached {
		return cch, nil
	}

	helix, err := c.Helix()
	if err != nil {
		return
	}

	chid, err := helix.GetUserID(name)
	if err != nil {
		return
	}

	ch = &ChannelInfo{
		Client: c,
		Name:   name,
		id:     chid,
	}

	c.channels[name] = ch
	return
}

func (c *ChannelInfo) resolveCheermotes() (err error) {
	if c.cheerInfo != nil {
		return nil
	}

	helix, err := c.Client.Helix()
	if err != nil {
		return
	}

	prefixes, infos, err := helix.GetCheermotes(c.id)
	if err != nil {
		return
	}

	cmPattern, err := regexp.Compile("\\b(" + strings.Join(prefixes, "|") + ")([0-9]+)\\b")
	if err != nil {
		return
	}

	c.cheerMatch = cmPattern
	c.cheerInfo = infos
	return
}

func (c *ChannelInfo) GetEmoteURL(emoteID string) (emoteURL string, err error) {
	if c.cheerInfo != nil {
		if cminfo, iscm := c.cheerInfo[emoteID]; iscm {
			emoteURL = cminfo.ImageURL
			return
		}
	}

	emoteURL = "https://static-cdn.jtvnw.net/emoticons/v2/" +
		emoteID + "/default/dark/1.0"
	return
}

func (c *ChannelInfo) GetBadgeURL(badgeID string) (badgeURL string, err error) {
	var helix *HelixAPI
	if c.badges == nil || c.Client.badges == nil {
		helix, err = c.Client.Helix()
		if err != nil {
			return
		}
	}

	if c.badges == nil {
		c.badges, err = helix.GetChannelChatBadges(c.id)
		if err != nil {
			return
		}
	}

	if c.Client.badges == nil {
		c.Client.badges, err = helix.GetGlobalChatBadges()
		if err != nil {
			return
		}
	}

	if badgeInfo, ok := c.badges[badgeID]; ok {
		return badgeInfo.ImageURL, nil
	}

	if badgeInfo, ok := c.Client.badges[badgeID]; ok {
		return badgeInfo.ImageURL, nil
	}

	err = ErrNoSuchBadge
	return
}
