package main

import (
	"github.com/gomodule/redigo/redis"
	"log"
)

type RedisSubscriber struct {
	pool *redis.Pool
	hub  *Hub
}

func (s *RedisSubscriber) listen() {
	c := s.pool.Get()
	defer c.Close()
	psc := &redis.PubSubConn{Conn: c}
	psc.Subscribe("pubsub")
	for {
		switch v := psc.Receive().(type) {
		case redis.Message:
			log.Printf(
				"Read redis message \"%s\" on channel \"%s\" \n",
				v.Data,
				v.Channel,
			)
			s.hub.broadcast <- string(v.Data)
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
