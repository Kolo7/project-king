package main

import (
	"errors"
	"log"
	"time"

	"github.com/dtm-labs/rockscache"
	"github.com/redis/go-redis/v9"
)

var (
	ErrNotFound = errors.New("not found")
)

func main() {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "0.0.0.0:6379",
		Password: "gHmNkVBd88sZybj",
		DB:       0,
	})
	ds := NewMockDataSource()
	key := "key1"
	ds.set(key, "value1")
	// new a client for rockscache using the default options
	rc := rockscache.NewClient(redisClient, NewDefaultOptions())

	v, err := wrapGetCache(rc, key, ds.get)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("缓存值: %s", v)

	ds.set(key, "value2")
	err = rc.TagAsDeleted(key)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("删除缓存key: %s", key)
	v, err = wrapGetCache(rc, key, ds.get)
	if err != nil {
		log.Fatal(err)
	}
	v2, ok := ds.get(key)
	if ok {
		log.Printf("真实值: %s", v2)
	}
	log.Printf("缓存值: %s", v)
}

func wrapGetCache(rc *rockscache.Client, key string, get func(k string) (string, bool)) (string, error) {
	// use Fetch to fetch data
	// 1. the first parameter is the key of the data
	// 2. the second parameter is the data expiration time
	// 3. the third parameter is the data fetch function which is called when the cache does not exist
	v, err := rc.Fetch(key, 300*time.Second, func() (string, error) {
		// fetch data from database or other sources
		v, ok := get(key)
		if ok {
			return v, nil
		} else {
			return "", ErrNotFound
		}
	})
	return v, err
}

func NewDefaultOptions() rockscache.Options {
	默认值 := rockscache.NewDefaultOptions()
	默认值.Delay = 1 * time.Second
	rockscache.SetVerbose(true)
	return 默认值
}

type MockDataSource struct {
	data map[string]string
}

func NewMockDataSource() *MockDataSource {
	return &MockDataSource{
		data: map[string]string{},
	}
}

func (m *MockDataSource) get(k string) (string, bool) {
	v, ok := m.data[k]
	log.Printf("数据源get,k: %s", k)
	return v, ok
}

func (m *MockDataSource) set(k, v string) {
	log.Printf("数据源set,k: %s, v: %s", k, v)
	m.data[k] = v
}
