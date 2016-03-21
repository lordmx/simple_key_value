package main

import (
	"flag"
	"log"
	"time"
)

var (
	host = flag.String("host", "0.0.0.0:1234", "Host to listen")
)

func init() {
	flag.Parse()
}

func main() {
	server := NewServer()
	cache := NewCache()

	commands := &commands{
		cache: cache,
	}

	server.OnConnect(func(client *Client) {
		log.Printf("[%v] Client connected. ID: %d", time.Now(), client.ID)
	})

	server.OnDisconnect(func(client *Client, err error) {
		log.Printf("[%v] Client disconnected. ID: %d. Error: %s", time.Now(), client.ID, err)
	})

	server.OnMessage(func(client *Client, data []byte) {
		log.Printf(
			"[%v] Client sent message. ID: %d. Message: %s",
			time.Now(),
			client.ID,
			string(data),
		)
		r := commands.run(data)

		if r.err != nil {
			server.sendToClient(client, []byte(r.err.Error()))
		} else {
			server.sendToClient(client, []byte(r.value))
		}

		server.sendToClient(client, []byte{'\n'})
	})

	server.Listen(*host)
}
