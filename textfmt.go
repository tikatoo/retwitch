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
	Bits      int
	BitsColor string
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
			if segment.Bits != 0 {
				b.WriteString("*")
				b.WriteString(strconv.Itoa(segment.Bits))
			}
			if segment.BitsColor != "" {
				b.WriteString(segment.BitsColor)
			}
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
				Bits      int    `json:"bits,omitempty"`
				BitsColor string `json:"bits_color,omitempty"`
			}

			enc, err := json.Marshal(emoteSegment{
				EmoteID:   segment.EmoteID,
				EmoteText: segment.EmoteText,
				Bits:      segment.Bits,
				BitsColor: segment.BitsColor,
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
	CheerValue int
	CheerColor string
	StartAt    int
	EndAt      int
}

func (c *ChannelInfo) parseIRCText(msgtext string, emotespec string) (Text, error) {
	var emotelist []emoteLocation
	if emotespec != "" {
		speclist := strings.Split(emotespec, "/")
		emotelist = make([]emoteLocation, 0,
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
	}

	emotelist = append(emotelist, c.parseIRCCheer(msgtext)...)

	if len(emotelist) == 0 {
		return Text{{Text: msgtext}}, nil
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

		if entry.CheerValue != 0 {
			segment.Bits = entry.CheerValue
			segment.BitsColor = entry.CheerColor
		}
	}

	if off < len(msgtext) {
		segments = append(segments, TextSegment{Text: msgtext[off:]})
	}

	return Text(segments), nil
}

func (c *ChannelInfo) parseIRCCheer(msgtext string) (locs []emoteLocation) {
	if c == nil {
		return
	}

	if err := c.resolveCheermotes(); err != nil {
		return
	}

	m := c.cheerMatch.FindAllStringSubmatchIndex(msgtext, -1)
	locs = make([]emoteLocation, 0, len(m))

	for _, match := range m {
		cmPrefix := msgtext[match[2]:match[3]]
		cmValueString := msgtext[match[4]:match[5]]
		cmValue, err := strconv.Atoi(cmValueString)
		if err != nil {
			continue
		}

		cheerInfo := c.cheerInfo[cmPrefix+"1"]
		for cheerInfo.NextTierID != "" {
			nextCheerInfo := c.cheerInfo[cheerInfo.NextTierID]
			if cmValue < nextCheerInfo.CheerValue {
				break
			}

			cheerInfo = nextCheerInfo
		}

		locs = append(locs, emoteLocation{
			EmoteID:    cheerInfo.CheerID,
			CheerValue: cmValue,
			CheerColor: cheerInfo.CheerColor,
			StartAt:    match[2],
			EndAt:      match[5] - 1,
		})
	}

	return
}
