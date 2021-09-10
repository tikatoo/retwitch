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
	flag.BoolVar(&useJSON, "json", false, "show json-formatted messages")
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
		}
	}
}
