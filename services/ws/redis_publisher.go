package main

import (
	"github.com/gomodule/redigo/redis"
	"log"
)

type WsMessage struct {
	Data string
}

type RedisPublisher struct {
	pool *redis.Pool
}

func (p *RedisPublisher) publish(message WsMessage) {
	log.Println("Publishing message to redis", message.Data)
	c := p.pool.Get()
	defer c.Close()
	c.Send("PUBLISH", "pubsub", message.Data)
	c.Flush()
}
