# Websocket examples: byte/text stream

websocket 协议通常在 Message 层级上进行使用（通过 ReadMessage, WriteMessage），假如用户想要将 websocket 协议当作一个用于传输字节流或文本流的协议，可以使用以下选项：

```go
// server side:
opts := []websocket.ServerOption{
    websocket.WithServerMessageType(websocket.Binary), // or websocket.Text
}
s, err := websocket.NewService(ln, handler, opts...)
// client side:
c, err := websocket.Dial(url,
    websocket.WithClientMessageType(websocket.Binary), // or websocket.Text (same with server).
)
```

这些选项分别指定了服务端和客户端的消息类型，这样用户可以直接在连接上使用 Read/Write API：

```go
// server side:
buf := make([]byte, 100)
n, err := c.Read(buf)
if err != nil {
    return err
}
_, err = c.Write(buf[:n])
// client side:
_, err = c.Write([]byte(text))
buf := make([]byte, len(text))
n, err := c.Read(buf)
```
