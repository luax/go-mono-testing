package main

import (
	"log"
)

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan string

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan string),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			log.Println("Registered new client", client.id)
			h.clients[client] = true
		case client := <-h.unregister:
			log.Println("Unregister client", client.id)
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			log.Println("Broadcasting message")
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					log.Println("Not able to broadcast message to client", client.id)
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}
