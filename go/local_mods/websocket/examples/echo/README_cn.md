# Websocket examples: echo

## Websocket

* 启动服务端：

```shell
$ go run server/main.go 
listen  :9876
```

* 启动客户端：

```shell
$ go run client/main.go
receive type: Text, data: hello world!
receive type: Binary, data: helloworld
```

## Websocket over TLS

* 启动服务端：

```shell
$ go run server/main.go -enabletls
listen  :9876
```

* 启动客户端：

```shell
$ go run client/main.go -enabletls
receive type: Text, data: hello world!
receive type: Binary, data: helloworld
```

## WritevMessage

可以使用 `WritevMessage` 来将多个 byte slices 写到一个消息中：

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
