package cache

import (
	"github.com/go-redis/redis"
)

type (
	// Engine struct ..
	Engine struct {
		*redis.Client
	}
)

// New function return engine struct with setuped redis client
func New(redisHost string) Engine {
	c := redis.NewClient(&redis.Options{
		Addr: redisHost,
	})
	return Engine{c}
}
