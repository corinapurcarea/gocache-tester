package main

import (
	"github.com/bradfitz/gomemcache/memcache"
	"github.com/dgraph-io/ristretto"
	"github.com/eko/gocache/cache"
	"github.com/eko/gocache/marshaler"
	"github.com/eko/gocache/store"
	redis "github.com/go-redis/redis/v7"
	"log"
	"time"
)

type Book struct {
	Title string
	Id    int
}

func main() {

	client, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1000,
		MaxCost:     100,
		BufferItems: 64,
	})
	ristrettoStore := store.NewRistretto(client, &store.Options{Expiration: 2 * time.Second})

	redisStore := store.NewRedis(redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	}), &store.Options{Expiration: 3 * time.Second})

	memcacheStore := store.NewMemcache(
		memcache.New("127.0.0.1:11211"),
		&store.Options{
			Expiration: 5 * time.Second,
		},
	)

	// Initialize chained cache
	cacheManager := cache.NewChain(
		cache.New(ristrettoStore),
		cache.New(redisStore),
		cache.New(memcacheStore),
	)

	// Initializes marshaler
	marshal := marshaler.New(cacheManager)

	key := "my-key"
	value := Book{Title: "Foundation", Id: 1}

	err = marshal.Set(key, &value, nil)
	if err != nil {
		panic(err)
	}
	time.Sleep(4 * time.Second) //let it time to expire

	var retValue Book

	_, _ = marshal.Get(key, &retValue)
	log.Printf("returned value: %v\n", retValue)

	time.Sleep(1 * time.Second) // for ristretto to accept the value

	_, _ = marshal.Get(key, &retValue)
	log.Printf("returned value: %v\n", retValue)
}
