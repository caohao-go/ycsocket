// Package main is the main package.
package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"

	"git.woa.com/trpc-go/tnet/extensions/websocket"
)

var (
	addr      = flag.String("addr", "127.0.0.1:9876", "dial server address")
	enableTLS = flag.Bool("enabletls", false, "enable TLS")
	hello     = []byte("hello")
	world     = []byte("world")
)

func main() {
	flag.Parse()
	var url string
	var opts []websocket.ClientOption
	if *enableTLS {
		url = fmt.Sprintf("wss://%s", *addr)
		opts = append(opts, websocket.WithClientTLSConfig(&tls.Config{InsecureSkipVerify: true}))
	} else {
		url = fmt.Sprintf("ws://%s", *addr)
	}
	fmt.Printf("dial %s\n", url)
	c, err := websocket.Dial(url, opts...)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()
	text := "hello world!"
	c.WriteMessage(websocket.Text, []byte(text))
	tp, data, err := c.ReadMessage()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("receive type: %s, data: %s\n", tp, data)

	// writev message example:
	if err := c.WritevMessage(websocket.Binary, hello, world); err != nil {
		log.Fatal(err)
	}
	tp, data, err = c.ReadMessage()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("receive type: %s, data: %s\n", tp, data)
}
