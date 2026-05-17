# Websocket examples: byte/text stream

Typically websocket is used to work on message level (ReadMessage or WriteMessage). If users want to use websocket as a protocol to transmit byte stream or text stream. Certain options are provided:

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

Then users can use Read/Write API directly on server/client connections without specifying the message type:

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
