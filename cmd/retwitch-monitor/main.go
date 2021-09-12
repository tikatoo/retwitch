package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/tikatoo/retwitch"
)

func main() {
	var useJSON bool
	var showURLs bool
	flag.BoolVar(&useJSON, "json", false, "show json-formatted messages")
	flag.BoolVar(&showURLs, "urls", false, "print image urls (text mode only)")
	flag.Parse()
	channels := flag.Args()

	client, err := retwitch.NewAnonymousClient()
	if err != nil {
		panic(err)
	}

	for _, channel := range channels {
		time.Sleep(500)
		fmt.Println("Joining #" + channel)
		err = client.Join(channel)
		if err != nil {
			panic(err)
		}
	}

	if useJSON {
		enc := json.NewEncoder(os.Stdout)
		for event := range client.LiveEvents() {
			if err := enc.Encode(&event); err != nil {
				fmt.Println(err)
			}
		}
	} else {
		for event := range client.LiveEvents() {
			fmt.Println(&event)

			if showURLs {
				printURLs(client, event)
			}
		}
	}
}

func printURLs(client *retwitch.Client, event retwitch.LiveEvent) {
	channel, err := client.GetChannel(event.Channel)
	if err != nil {
		fmt.Println("    error:", err)
		return
	}

	for _, badgeID := range event.Sender.Badges {
		badgeURL, err := channel.GetBadgeURL(badgeID)
		if err != nil {
			fmt.Printf("    badge %s error: %s\n", badgeID, err)
		} else {
			fmt.Printf("    badge %s: %q\n", badgeID, badgeURL)
		}
	}

	showedURLFor := map[string]struct{}{}

	for _, segment := range event.Message {
		if segment.EmoteID == "" {
			continue
		}

		if _, exists := showedURLFor[segment.EmoteID]; !exists {
			showedURLFor[segment.EmoteID] = struct{}{}
			emoteURL, err := channel.GetEmoteURL(segment.EmoteID)
			if err != nil {
				fmt.Printf("    emote %s error: %s\n", segment.EmoteText, err)
			} else {
				fmt.Printf("    emote %s: %q\n", segment.EmoteText, emoteURL)
			}
		}
	}
}
