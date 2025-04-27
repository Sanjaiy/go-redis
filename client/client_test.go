package client

import (
	"context"
	"fmt"
	"log"
	"sync"
	"testing"
)

func TestNewClients(t *testing.T) {	
	nClients := 10

	var wg sync.WaitGroup
	wg.Add(nClients)

	for i :=0; i < nClients; i++ {
		go func(it int){
			c, err := New("localhost:5001")
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
}