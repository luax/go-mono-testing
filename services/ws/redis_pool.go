package main

import (
	"github.com/gomodule/redigo/redis"
	"log"
	"net/url"
)

func parseURL(us string) (string, string, error) {
	u, err := url.Parse(us)
	if err != nil {
		return "", "", err
	}
	var password string
	if u.User != nil {
		password, _ = u.User.Password()
	}
	var host string
	if u.Host == "" {
		host = "localhost"
	} else {
		host = u.Host
	}
	return host, password, nil
}

func NewRedisPool(url string) (*redis.Pool, error) {
	host, password, err := parseURL(url)
	if err != nil {
		log.Println("Error parsing redis URL")
		return nil, err
	}
	return &redis.Pool{
		MaxActive: 20,
		MaxIdle:   20,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", host)
			if err != nil {
				log.Println("Error connecting to redis", host)
				return nil, err
			}
			if password != "" {
				if _, err := c.Do("AUTH", password); err != nil {
					log.Println("Redis password error", password)
					c.Close()
					return nil, err
				}
			}
			log.Println("Connected to redis")
			return c, nil
		},
	}, nil
}
