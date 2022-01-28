package main

import (
	"encoding/json"
	"os"
)

func main() {
	loc := os.Getenv("NETRELAY_CONFIG")
	if loc == "" {
		loc = "deploy/relay.json"
	}
	fp, err := os.Open(loc)
	if err != nil {
		panic(err)
	}
	decoder := json.NewDecoder(fp)

	var relays []Relay
	if err := decoder.Decode(&relays); err != nil {
		panic(err)
	}

	for _, relay := range relays {
		go relay.Start()
	}

	select {}
}
