package main

import (
	"github.com/gomodule/redigo/redis"
	"github.com/gorilla/websocket"
	"log"
	"mono/lib/env"
	"net/http"
	"os"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var pool *redis.Pool
var publisher *Publisher

func handleWebsocket(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling WS connection")
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
		return
	}
	publisher.add <- conn
	// TODO: Move to publisher?
	for {
		var msg WsMessage
		log.Println("Conn waiting for msg")
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("Error waiting for message: %v", err)
			publisher.remove <- conn
			break
		}
		log.Println("Conn received message", msg)
		publisher.publish(msg)
	}
}

func main() {
	log.Println("WS server started")
	env.PrintEnvVariables()

	// Redis
	redisURL := os.Getenv(os.Getenv("REDIS_URL"))
	var err error
	pool, err = CreateRedisPool(redisURL)
	if err != nil {
		log.Fatal("Could not setup pool", err)
		return
	}
	defer pool.Close()

	// Messaging
	messages := make(chan string)
	publisher = &Publisher{
		pool:        pool,
		add:         make(chan *websocket.Conn),
		remove:      make(chan *websocket.Conn),
		connections: make(map[*websocket.Conn]bool),
		messages:    messages,
	}
	go publisher.listen()
	subscriber := Subscriber{
		pool:     pool,
		messages: messages,
	}
	go subscriber.listen()

	// Handle Websocket
	http.HandleFunc("/", handleWebsocket)
	log.Fatal(http.ListenAndServe(":"+os.Getenv("PORT"), nil))

}
