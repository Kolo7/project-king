package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/allegro/bigcache/v3"
	"github.com/google/gops/agent"
)

var (
	BigCache *bigcache.BigCache
)

func main() {
	config := bigcache.Config{
		Shards:             1,
		LifeWindow:         5 * time.Minute,
		CleanWindow:        1 * time.Minute,
		MaxEntriesInWindow: 10,
		MaxEntrySize:       1024 * 10,
		//HardMaxCacheSize:   8192, // 8GB
		//OnRemove:           nil,  // 驱逐时的回调函数,nil的话不会回调，也不会发生驱逐
		//LifeWindow:         5 * time.Second,
		//MaxEntrySize:       1024 * 1024,
		//Verbose:            true,

		OnRemoveWithReason: func(key string, entry []byte, reason bigcache.RemoveReason) {
			switch reason {
			case bigcache.Deleted:
				log.Printf("remove entry, reason(deleted), key(%s)\n", key)
			case bigcache.Expired:
				log.Printf("remove entry, reason(expired), key(%s)\n", key)
			case bigcache.NoSpace:
				log.Printf("remove entry, reason(no space), key(%s)\n", key)
			}
		},
		Logger: log.Default(),
	}

	var initErr error
	BigCache, initErr = bigcache.New(context.Background(), config)
	if initErr != nil {
		panic(initErr)
	}

	if err := agent.Listen(agent.Options{}); err != nil {
		log.Fatalf("agent.Listen err: %v", err)
	}
	//go doReset()
	do()
}

func do() {
	fmt.Println("start cache")
	for {
		buf := make([]byte, 1024*1024)
		err := BigCache.Set(rangeKey(), buf)
		if err != nil {
			fmt.Println(err)
		}
		time.Sleep(time.Minute * 1)
	}
	fmt.Println("end cache")
}

func doReset() {
	fmt.Println("start reset")
	tick := time.NewTicker(time.Second * 10)

	defer tick.Stop()
	for {
		<-tick.C
		err := BigCache.Reset()
		if err != nil {
			log.Println(err)
		}
		log.Printf("reset cache success, capacity(%d)\n", BigCache.Capacity()/(1024*1024))
	}
}

func rangeKey() string {
	//return "test"
	return "key:" + strconv.Itoa(rand.Intn(5))
}
