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
	buf  = make([]byte, 100)
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
		websocket.WithServerMessageType(websocket.Binary), // or websocket.Text
	}
	s, err := websocket.NewService(ln, handler, opts...)
	if err != nil {
		log.Fatal(err)
	}
	log.Print(s.Serve(context.Background()))
}

func handler(c websocket.Conn) error {
	n, err := c.Read(buf)
	if err != nil {
		return err
	}
	_, err = c.Write(buf[:n])
	return err
}
