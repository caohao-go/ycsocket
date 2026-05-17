// Package main is the main package.
package main

import (
	"flag"
	"fmt"
	"log"

	"git.woa.com/trpc-go/tnet/extensions/websocket"
)

var (
	addr = flag.String("addr", "127.0.0.1:9876", "dial server address")
)

func main() {
	flag.Parse()
	url := fmt.Sprintf("ws://%s", *addr)
	fmt.Printf("dial %s\n", url)
	c, err := websocket.Dial(url,
		websocket.WithClientMessageType(websocket.Binary), // or websocket.Text (same with server).
	)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()
	text := "hello world!"
	_, err = c.Write([]byte(text))
	if err != nil {
		log.Fatal(err)
	}
	buf := make([]byte, len(text))
	n, err := c.Read(buf)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("receive data: %s\n", string(buf[:n]))
}
