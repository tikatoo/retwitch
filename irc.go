package retwitch

import (
	"math/rand"
	"strconv"
	"time"

	"github.com/lrstanley/girc"
)

func ircConnect(username string, authcode string) (*girc.Client, error) {
	if username == "" {
		authcode = "BLANK"
		username = makeAnonUser()
	}

	irc := girc.New(girc.Config{
		Server:     "irc.chat.twitch.tv",
		Port:       6667,
		SSL:        false,
		ServerPass: authcode,
		Nick:       username,
		User:       username,
		SupportedCaps: map[string][]string{
			"twitch.tv/commands": nil,
			"twitch.tv/tags":     nil,
		},
	})

	initchan := make(chan error)

	irc.Handlers.Add(girc.INITIALIZED, func(client *girc.Client, event girc.Event) {
		initchan <- nil
	})

	go func() {
		initchan <- irc.Connect()
		close(initchan)
	}()

	err := <-initchan
	if err != nil {
		return nil, err
	}

	return irc, nil
}

func ircToLiveEvent(ircEvent girc.Event) (event LiveEvent) {
	event = LiveEvent{
		Time:    ircEvent.Timestamp,
		Channel: ircEvent.Params[0][1:],
		Sender:  ircToSender(ircEvent),
		Kind:    MessageEvent,
		Message: ircToMessage(ircEvent),
	}

	if ircEvent.IsAction() {
		event.Kind = ActionEvent
	}

	return
}

func ircToSender(ircEvent girc.Event) (sender Viewer) {
	sender = Viewer{
		User:    ircEvent.Source.Name,
		Display: ircEvent.Source.Name,
		Color:   "",
	}

	if display, ok := ircEvent.Tags.Get("display-name"); ok {
		sender.Display = display
	}

	if color, ok := ircEvent.Tags.Get("color"); ok {
		sender.Color = color
	}

	return
}

func ircToMessage(ircEvent girc.Event) (msg string) {
	msg = ircEvent.Last()
	if ok, ctcp := ircEvent.IsCTCP(); ok {
		msg = ctcp.Text
	}

	return
}

func (c *Client) onPrivmsg(ircClient *girc.Client, event girc.Event) {
	if event.Params[0][0] == '#' {
		c.levs <- ircToLiveEvent(event)
	}
}

func (c *Client) waitIRC(channel string, replies ...string) <-chan girc.Event {
	replySet := make(map[string]struct{}, len(replies))
	for _, cmd := range replies {
		replySet[cmd] = struct{}{}
	}

	evch := make(chan girc.Event)
	c.irc.Handlers.AddTmp(girc.ALL_EVENTS, 0, func(_ *girc.Client, event girc.Event) bool {
		if _, ok := replySet[event.Command]; ok && event.Params[0] == channel {
			evch <- event
			return true
		}
		return false
	})

	return evch
}

func makeAnonUser() string {
	return "justinfan" + strconv.Itoa(rand.Intn(999999))
}

func init() {
	rand.Seed(time.Now().UnixMicro())
}
