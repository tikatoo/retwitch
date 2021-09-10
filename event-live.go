package retwitch

import (
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
	User    string
	Display string
	Color   string
}

type LiveEvent struct {
	Time    time.Time
	Channel string
	Sender  Viewer
	Kind    LiveEventKind
	Message string
}

func (v *Viewer) String() string {
	if v.Display == "" {
		return v.User + v.Color
	} else if v.User == strings.ToLower(v.Display) {
		return v.Display + v.Color
	}

	return v.Display + v.Color + "@" + v.User
}

func (e *LiveEvent) String() string {
	sender := e.Sender.String()
	b := &strings.Builder{}
	b.Grow(16 + len(e.Channel) + len(sender) + len(e.Message))
	b.WriteString("[")
	b.WriteString(e.Time.Format("03:04"))
	b.WriteString(" in ")
	b.WriteString(e.Channel)
	b.WriteString("] ")

	switch e.Kind {
	case MessageEvent:
		b.WriteString(sender)
		b.WriteString(": ")
		b.WriteString(e.Message)
	case ActionEvent:
		b.WriteString("* ")
		b.WriteString(sender)
		b.WriteString(" ")
		b.WriteString(e.Message)
	default:
		b.WriteString("<unknown event ")
		b.WriteString(strconv.Itoa(int(e.Kind)))
		b.WriteString(" from ")
		b.WriteString(sender)
		b.WriteString(">")
	}

	return b.String()
}
