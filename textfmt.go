package retwitch

import (
	"bytes"
	"encoding/json"
	"sort"
	"strconv"
	"strings"
)

type Text []TextSegment

type TextSegment struct {
	Text      string
	EmoteID   string
	EmoteText string
}

func (t Text) String() string {
	b := &strings.Builder{}

	for _, segment := range t {
		b.WriteString(segment.Text)

		if segment.EmoteID != "" {
			b.WriteString("<")
			if segment.EmoteText != "" {
				b.WriteString(segment.EmoteText)
				b.WriteString(":")
			}
			b.WriteString(segment.EmoteID)
			b.WriteString(">")
		}
	}

	return b.String()
}

func (t Text) MarshalJSON() ([]byte, error) {
	buf := &bytes.Buffer{}
	buf.WriteRune('[')

	for _, segment := range t {
		if segment.Text != "" {
			enc, err := json.Marshal(segment.Text)
			if err != nil {
				return nil, err
			}

			buf.Write(enc)
			buf.WriteRune(',')
		}

		if segment.EmoteID != "" {
			type emoteSegment struct {
				EmoteID   string `json:"emote"`
				EmoteText string `json:"name,omitempty"`
			}

			enc, err := json.Marshal(emoteSegment{
				EmoteID:   segment.EmoteID,
				EmoteText: segment.EmoteText,
			})
			if err != nil {
				return nil, err
			}

			buf.Write(enc)
			buf.WriteRune(',')
		}
	}

	if buf.Len() > 1 {
		buf.Bytes()[buf.Len()-1] = ']'
	} else {
		buf.WriteRune(']')
	}

	return buf.Bytes(), nil
}

type emoteLocation struct {
	EmoteID    string
	CheerValue uint
	StartAt    int
	EndAt      int
}

func parseIRCText(msgtext string, emotespec string) (Text, error) {
	if emotespec == "" {
		return Text{{Text: msgtext}}, nil
	}

	speclist := strings.Split(emotespec, "/")
	emotelist := make([]emoteLocation, 0,
		strings.Count(emotespec, "/")+strings.Count(emotespec, ","))

	for _, specentry := range speclist {
		splitspec := strings.Split(specentry, ":")
		emoteid := splitspec[0]

		splitspec = strings.Split(splitspec[1], ",")
		for _, specentry := range splitspec {
			offsets := strings.Split(specentry, "-")
			startat, err := strconv.ParseInt(offsets[0], 10, 32)
			if err != nil {
				return nil, err
			}

			endat, err := strconv.ParseInt(offsets[1], 10, 32)
			if err != nil {
				return nil, err
			}

			emotelist = append(emotelist, emoteLocation{
				EmoteID: emoteid,
				StartAt: int(startat),
				EndAt:   int(endat),
			})
		}
	}

	sort.Slice(emotelist, func(i int, j int) bool {
		return emotelist[i].StartAt < emotelist[j].StartAt
	})

	segments := []TextSegment{}
	off := 0
	for _, entry := range emotelist {
		segments = append(segments, TextSegment{})
		segment := &segments[len(segments)-1]
		if entry.StartAt > off {
			segment.Text = msgtext[off:entry.StartAt]
		}

		off = entry.EndAt + 1
		segment.EmoteID = entry.EmoteID
		segment.EmoteText = msgtext[entry.StartAt:off]
	}

	if off < len(msgtext) {
		segments = append(segments, TextSegment{Text: msgtext[off:]})
	}

	return Text(segments), nil
}
