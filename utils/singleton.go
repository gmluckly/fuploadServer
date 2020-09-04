package utils

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"time"
)

type Singleton struct{}

var singletonIntance *Singleton
var once sync.Once

func GetSingletonObj() *Singleton {
	once.Do(func() {
		fmt.Println("Create obj")
		singletonIntance = &Singleton{}
	})
	return singletonIntance
}

func main() {
	c := sync.NewCond(&sync.Mutex{})
	for i := 0; i < 10; i++ {
		go listen(c)
	}
	time.Sleep(1 * time.Second)
	go broadcast(c)
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	<-ch
}

func broadcast(c *sync.Cond) {
	c.L.Lock()
	c.Broadcast()
	c.L.Unlock()
}

func listen(c *sync.Cond) {
	c.L.Lock()
	c.Wait()
	fmt.Println("listen")
	c.L.Unlock()
}
