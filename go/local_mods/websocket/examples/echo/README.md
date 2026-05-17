# Websocket examples: echo

## Websocket

* start server:

```shell
$ go run server/main.go 
listen  :9876
```

* start client:

```shell
$ go run client/main.go
receive type: Text, data: hello world!
receive type: Binary, data: helloworld
```

## Websocket over TLS

* start server

```shell
$ go run server/main.go -enabletls
listen  :9876
```

* start client:

```shell
$ go run client/main.go -enabletls
receive type: Text, data: hello world!
receive type: Binary, data: helloworld
```

## WritevMessage

If users want to write multiple byte slices into one message, use `WritevMessage`:

```go
// writev message example:
if err := c.WritevMessage(websocket.Binary, hello, world); err != nil {
    log.Fatal(err)
}
tp, data, err = c.ReadMessage()
if err != nil {
    log.Fatal(err)
}
fmt.Printf("receive type: %s, data: %s\n", tp, data)
```
