// Package main is the main package.
package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"git.woa.com/trpc-go/tnet/extensions/websocket"
)

var (
	addr  = flag.String("addr", "127.0.0.1:9876", "dial server address")
	hello = []byte("hello")
	world = []byte("world")
)

func main() {
	flag.Parse()
	url := fmt.Sprintf("ws://%s", *addr)
	fmt.Printf("dial %s\n", url)
	c, err := websocket.Dial(url)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()
	if err := c.WriteMessage(websocket.Ping, []byte(hello)); err != nil {
		log.Fatal(err)
	}
	if err := c.WriteMessage(websocket.Pong, []byte(world)); err != nil {
		log.Fatal(err)
	}
	time.Sleep(time.Second) // wait till the server runs the control handler.
}
