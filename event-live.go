package retwitch

import (
	"strings"
	"time"
)

type Viewer struct {
	User    string
	Display string
}

type LiveEvent struct {
	Time    time.Time
	Channel string
	Sender  Viewer
	Message string
}

func (v *Viewer) String() string {
	if v.Display == "" {
		return v.User
	} else if v.User == strings.ToLower(v.Display) {
		return v.Display
	}

	return v.Display + "@" + v.User
}

func (e *LiveEvent) String() string {
	sender := e.Sender.String()
	b := &strings.Builder{}
	b.Grow(14 + len(e.Channel) + len(sender) + len(e.Message))
	b.WriteString("[")
	b.WriteString(e.Time.Format("03:04"))
	b.WriteString(" in ")
	b.WriteString(e.Channel)
	b.WriteString("] ")
	b.WriteString(sender)
	b.WriteString(": ")
	b.WriteString(e.Message)
	return b.String()
}
