package main

import (
	"github.com/gorilla/websocket"
	"log"
	"math/rand"
	"time"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 45 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

// Client is an middleman between the websocket connection and the hub.
type Client struct {
	// The websocket connection.
	ws *websocket.Conn

	// Buffered channel of outbound messages.
	send chan string

	redisPublisher *RedisPublisher

	hub *Hub

	id int
}

func NewClient(w *websocket.Conn, r *RedisPublisher, h *Hub) *Client {
	return &Client{
		ws:             w,
		redisPublisher: r,
		hub:            h,
		send:           make(chan string, 256),
		id:             rand.Intn(9999999),
	}
}

// readPump pumps messages from the websocket connection to the hub.
func (c *Client) readPump() {
	defer func() {
		log.Println("Closing read pump", c.id)
		c.hub.unregister <- c
		c.ws.Close()
	}()
	c.ws.SetReadLimit(maxMessageSize)
	c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(func(string) error {
		c.ws.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		var msg WsMessage
		log.Println("Client waiting for msg", c.id)
		err := c.ws.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Println("Error waiting for message", err, c.id)
			}
			break
		}
		log.Println("Client received message", msg, c.id)
		c.redisPublisher.publish(msg)
	}
}

// write writes a message with the given message type and payload.
func (c *Client) write(mt int, payload []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteMessage(mt, payload)
}

// writePump pumps messages from the hub to the websocket connection.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		log.Println("Closing write pump", c.id)
		ticker.Stop()
		c.ws.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				// The hub closed the channel.
				log.Println("Hub closed channel", c.id)
				c.write(websocket.CloseMessage, []byte{})
				return
			}
			log.Println("Writing message", c.id)
			c.ws.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.ws.WriteJSON(message); err != nil {
				log.Println("Error writing message", c.id, err)
				return
			}
		case <-ticker.C:
			log.Println("Tick", c.id)
			c.ws.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.write(websocket.PingMessage, []byte{}); err != nil {
				log.Println("Could not ping client", c.id)
				return
			}
		}
	}
}
