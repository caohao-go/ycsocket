# Websocket examples: customized control frame handler (advanced)

Normally, users would not want to utilize this feature, because ping/pong/close frames have their default handling procedures defined in [RFC6455](https://datatracker.ietf.org/doc/rfc6455/).

Options are provided to allow users to implement their own logic of handling the control frames:

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

__Note:__ Once users write their own control handler, the default handling procedures will not be executed.

* start server:

```shell
go run server/main.go 
listen  :9876
```

* start client:

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
