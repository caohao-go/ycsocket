# Websocket examples: customized control frame handler (advanced)

通常情况下，用户不需要对控制帧指定自定义的处理逻辑，因为它们在 [RFC6455](https://datatracker.ietf.org/doc/rfc6455/) 中有默认的处理规则。

假如用户有自定义处理需求，可以使用以下选项：

```go
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
```

__注意:__ 一旦用户自定义了控制帧的处理逻辑，默认的处理逻辑将不会被执行。

示例运行：

* 启动服务端：

```shell
go run server/main.go 
listen  :9876
```

* 启动客户端：

```shell
go run client/main.go
```

```shell
# server output:
receive ping message: hello
enter customized ping handler
receive pong message: world
enter customized pong handler
```
