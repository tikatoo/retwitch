package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/tikatoo/retwitch"
)

func main() {
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

	for event := range client.LiveEvents() {
		fmt.Println(&event)
	}
}
