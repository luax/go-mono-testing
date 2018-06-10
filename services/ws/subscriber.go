package main

import (
	"github.com/gomodule/redigo/redis"
	"log"
)

type Subscriber struct {
	pool     *redis.Pool
	messages chan string
}

func (subscriber *Subscriber) listen() {
	c := pool.Get()
	defer c.Close()
	psc := &redis.PubSubConn{Conn: c}
	psc.Subscribe("pubsub")
	for {
		switch v := psc.Receive().(type) {
		case redis.Message:
			log.Printf(
				"Received redis message \"%s\" on channel \"%s\" \n",
				v.Data,
				v.Channel,
			)
			// TODO: Publish directly using publisher or use channel?
			subscriber.messages <- string(v.Data)
		case redis.Subscription:
			log.Printf(
				"Subscribed to redis channel \"%s\", kind \"%s\", count \"%d\"",
				v.Channel,
				v.Kind,
				v.Count,
			)
		case error:
			log.Fatalf(
				"Redis subscriber error: %#v",
				v,
			)
			return
		}
	}
}
