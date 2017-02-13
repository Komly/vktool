package main

import (
	"flag"
	"github.com/Komly/vktool/vk"
	"log"
	"os"
)

func main() {
	accessToken := flag.String("access_token", "", "VK access token")
	flag.Parse()
	if *accessToken == "" {
		println("You must provide -access_token param")
		os.Exit(1)
	}
	resp, err := vk.ApiCall("messages.getLongPollServer", map[string]string{
		"access_token": *accessToken,
		"use_ssl":      "1",
		"v":            "5.37",
	})
	if err != nil {
		log.Fatal(err)
	}

	server := resp["server"].(string)
	key := resp["key"].(string)
	ts := int(resp["ts"].(float64))

	var lastTs = ts
	for {
		var err error
		var updates []interface{}

		updates, lastTs, err = vk.MakeLongPollRequest(server, key, lastTs)
		if err != nil {
			log.Fatal(err)
		}

		for _, update := range updates {
			switch u := update.(type) {
			default:
				log.Printf("%+v", update)
			case vk.VkLongPollAddNewMessage:
				log.Printf("VkLongPollAddNewMessage: %+v", u)
			}
		}
	}
}
