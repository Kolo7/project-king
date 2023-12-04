package cache

import (
	"sync"

	redisotel "github.com/go-redis/redis/extra/redisotel/v8"
	"github.com/go-redis/redis/v8"
	"go.opentelemetry.io/otel/attribute"
)

var (
	redisCli *redis.Client
	once     sync.Once
)

func GetDBCache(add, pwd string) *redis.Client {
	once.Do(func() {
		opt := redis.Options{
			Addr:     add,
			Password: pwd,
			DB:       0,
		}

		redisCli = redis.NewClient(&opt)
		redisCli.AddHook(redisotel.NewTracingHook(
			redisotel.WithAttributes(attribute.String("redis addr", add)),
		))
	})
	return redisCli
}
