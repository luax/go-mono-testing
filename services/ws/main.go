package main

import (
	"github.com/gorilla/websocket"
	"log"
	"mono/lib/print"
	"net/http"
	"os"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {
	log.Println("WS server started")
	print.PrintEnvVariables()

	// Redis
	redisURL := os.Getenv(os.Getenv("REDIS_URL"))
	pool, err := NewRedisPool(redisURL)
	if err != nil {
		log.Fatal("Could not setup pool", err)
		return
	}
	defer pool.Close()

	// WS hub
	hub := NewHub()
	go hub.run()

	// Messaging
	redisPublisher := &RedisPublisher{
		pool: pool,
	}
	redisSubscriber := RedisSubscriber{
		hub:  hub,
		pool: pool,
	}
	go redisSubscriber.listen()

	// Handle Websocket
	handleWebsocket := func(w http.ResponseWriter, r *http.Request) {
		log.Println("Handling WS connection")
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
			return
		}
		client := NewClient(ws, redisPublisher, hub)
		hub.register <- client
		go client.writePump()
		go client.readPump()
	}
	http.HandleFunc("/", handleWebsocket)
	http.HandleFunc("/ws", handleWebsocket)
	log.Fatal(http.ListenAndServe(":"+os.Getenv("PORT"), nil))
}
