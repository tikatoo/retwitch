package retwitch

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"
)

type LiveEventKind int

const (
	MessageEvent LiveEventKind = iota
	ActionEvent
)

type Viewer struct {
	User    string   `json:"user"`
	Display string   `json:"display,omitempty"`
	Color   string   `json:"color,omitempty"`
	Badges  []string `json:"badges,omitempty"`
}

type LiveEvent struct {
	Time    time.Time     `json:"time"`
	Channel string        `json:"channel"`
	Sender  Viewer        `json:"sender"`
	Kind    LiveEventKind `json:"kind"`
	Message Text          `json:"message"`
}

func (v *Viewer) String() string {
	prefix := ""
	if len(v.Badges) > 0 {
		fmtbadges := make([]string, len(v.Badges))
		copy(fmtbadges, v.Badges)
		for i, badge := range fmtbadges {
			fmtbadges[i] = strings.TrimSuffix(badge, "/1")
		}

		prefix = "<" + strings.Join(fmtbadges, ", ") + "> "
	}

	if v.Display == "" {
		return prefix + v.User + v.Color
	} else if v.User == strings.ToLower(v.Display) {
		return prefix + v.Display + v.Color
	}

	return prefix + v.Display + v.Color + "@" + v.User
}

func (e *LiveEvent) String() string {
	sender := e.Sender.String()
	message := e.Message.String()
	b := &strings.Builder{}
	b.Grow(16 + len(e.Channel) + len(sender) + len(message))
	b.WriteString("[")
	b.WriteString(e.Time.Format("15:04"))
	b.WriteString(" in ")
	b.WriteString(e.Channel)
	b.WriteString("] ")

	switch e.Kind {
	case MessageEvent:
		b.WriteString(sender)
		b.WriteString(": ")
		b.WriteString(message)
	case ActionEvent:
		b.WriteString("* ")
		b.WriteString(sender)
		b.WriteString(" ")
		b.WriteString(message)
	default:
		b.WriteString("<unknown event ")
		b.WriteString(strconv.Itoa(int(e.Kind)))
		b.WriteString(" from ")
		b.WriteString(sender)
		b.WriteString(">")
	}

	return b.String()
}

func (k LiveEventKind) MarshalJSON() (result []byte, err error) {
	var word string

	switch k {
	case MessageEvent:
		word = "message"
	case ActionEvent:
		word = "action"
	default:
		return nil, errEventKind
	}

	return json.Marshal(word)
}

func (k *LiveEventKind) UnmarshalJSON(data []byte) (err error) {
	var word string
	if err = json.Unmarshal(data, &word); err != nil {
		return
	}

	switch word {
	case "message":
		*k = MessageEvent
	case "action":
		*k = ActionEvent
	default:
		err = errEventKind
	}

	return
}

var errEventKind = errors.New("invalid event kind")
