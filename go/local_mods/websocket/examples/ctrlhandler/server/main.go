// Package main is the main package.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"git.woa.com/trpc-go/tnet"
	"git.woa.com/trpc-go/tnet/extensions/websocket"
)

var (
	addr = flag.String("addr", ":9876", "websocket server listen address")
)

func main() {
	flag.Parse()
	ln, err := tnet.Listen("tcp", *addr)
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()
	fmt.Println("listen ", *addr)
	opts := []websocket.ServerOption{
		websocket.WithPingHandler(func(c websocket.Conn, b []byte) error {
			fmt.Printf("receive ping message: %s\n", string(b))
			fmt.Printf("enter customized ping handler\n")
			return nil
		}),
		websocket.WithPongHandler(func(c websocket.Conn, b []byte) error {
			fmt.Printf("receive pong message: %s\n", string(b))
			fmt.Printf("enter customized pong handler\n")
			return nil
		}),
	}
	s, err := websocket.NewService(ln, handler, opts...)
	if err != nil {
		log.Fatal(err)
	}
	log.Print(s.Serve(context.Background()))
}

func handler(c websocket.Conn) error {
	tp, msg, err := c.ReadMessage()
	if err != nil {
		return err
	}
	return c.WriteMessage(tp, msg)
}
