package main

import (
	"github.com/gomodule/redigo/redis"
	"github.com/gorilla/websocket"
	"log"
)

type Publisher struct {
	pool        *redis.Pool
	connections map[*websocket.Conn]bool
	add         chan *websocket.Conn
	remove      chan *websocket.Conn
	messages    chan string
}

func (publisher *Publisher) broadcast(message string) {
	log.Println("Broadcasting message", message)
	for conn := range publisher.connections {
		err := conn.WriteJSON(message)
		if err != nil {
			log.Printf("Error writing to conn: %v", err)
			conn.Close()
			delete(publisher.connections, conn)
		}
	}
}

func (publisher *Publisher) listen() {
	for {
		select {
		case message := <-publisher.messages:
			publisher.broadcast(message)
		case conn := <-publisher.remove:
			log.Println("Removing connection")
			_, exist := publisher.connections[conn]
			if exist {
				delete(publisher.connections, conn)
				conn.Close()
			}
		case conn := <-publisher.add:
			log.Println("Adding connection")
			publisher.connections[conn] = true
		}
	}
}

func (publisher *Publisher) publish(message WsMessage) {
	log.Println("Publishing message to redis", message.Data)
	c := publisher.pool.Get()
	defer c.Close()
	c.Send("PUBLISH", "pubsub", message.Data)
	c.Flush()
}
