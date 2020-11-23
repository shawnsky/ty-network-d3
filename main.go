package main

import s "github.com/shawnsky/ty-network-d3/server"

func main() {
	server := s.NewServer(":7341")

	// Define websocket connect url, default "/ws"
	server.WSPath = "/ws"
	// Define produce message url, default "/produce"
	server.PushPath = "/produce"

	// Run server
	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}
