package main

import (
	"context"
	"fmt"
	"time"

	"github.com/Kolo7/project-king/interval/model"
	cache "github.com/daoshenzzg/jetcache-go"
	"github.com/daoshenzzg/jetcache-go/local"
	"github.com/daoshenzzg/jetcache-go/remote"
	"github.com/go-redis/redis/v8"
)

func main() {
	opt := redis.Options{
		Addr:     "0.0.0.0:6379",
		Password: "gHmNkVBd88sZybj",
		DB:       0,
	}
	redisCli := redis.NewClient(&opt)
	jetCache := cache.New(
		cache.WithRemote(remote.NewGoRedisV8Adaptor(redisCli)),
		cache.WithLocal(local.NewFreeCache(256*local.MB, time.Minute)),
		cache.WithRefreshDuration(time.Minute),
		// cache.WithCodec(json.Name),
	)

	key := "key1"
	res := new(model.User)
	err := jetCache.Once(context.Background(), key, cache.Value(res), cache.Refresh(false), cache.TTL(time.Second),
		cache.Do(func(ctx context.Context) (interface{}, error) {
			return mockGetData(), nil
		}),
	)
	if err != nil {
		fmt.Print(err)
	}
	fmt.Print(res)
}

func mockGetData() *model.User {
	return &model.User{
		ID:          1,
		Username:    "12",
		CreatedAt:   time.Now(),
		UserSecrets: []model.UserSecret{},
	}
}
