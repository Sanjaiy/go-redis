package main

import (
	"context"
	"fmt"
	"log"
	"practise/Learnings/go-redis-clone/client"
	"sync"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func TestServerWithClient(t *testing.T) {
	server := NewServer(Config{})
	go func() {
		log.Fatal(server.Start())
	}()
	
	time.Sleep(1 * time.Second)

	nClients := 10

	var wg sync.WaitGroup
	wg.Add(nClients)

	for i :=0; i < nClients; i++ {
		go func(it int){
			c, err := client.New("localhost:5001")
			if err != nil {
				log.Fatal(err)
			}
			defer c.Close()

			key := fmt.Sprintf("client_foo_%d", i)
			val := fmt.Sprintf("client_bar_%d", i)
			if err := c.Set(context.TODO(), key, val); err != nil {
				log.Fatal(err)
			}

			v, err := c.Get(context.TODO(), key)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Printf("client %d got this val back => %s\n", i, v)
			wg.Done()
		}(i)
	}

	wg.Wait()

	time.Sleep(1 * time.Second)

	if len(server.peers) != 0 {
		t.Fatalf("expected 0 peers but got %d", len(server.peers))
	}
}

func TestRespWriteMap(t *testing.T) {
	in := map[string]string{
		"first": "1",
		"second": "2",
	}

	out := respWriteMap(in)
	fmt.Println(out)
}


func TestOfficialRedisClient(t *testing.T) {
	server := NewServer(Config{})
	go func(){
		log.Fatal(server.Start())
	}()

	time.Sleep(1 * time.Second)

	rdb := redis.NewClient(&redis.Options{
        Addr:     "localhost:5001",
        Password: "",
        DB:       0,
    })

	if err := rdb.Set(context.Background(), "key", "value", 0).Err(); err != nil {
		t.Fatal(err)
	}

	v, err := rdb.Get(context.Background(), "key").Result()
	if err != nil {
		t.Fatal(err)
	}
	
	if v != "value" {
		t.Fatalf("expected %s but got %s", "value", v)
	}
}