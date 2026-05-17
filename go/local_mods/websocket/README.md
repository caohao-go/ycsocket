# extension: websocket

This is an extension combining [tnet](https://git.woa.com/trpc-go/tnet) and [gobwas/ws](https://github.com/gobwas/ws) to improve memory usage and performance under millions of connections, while still providing idiomatic usage for [websocket](https://datatracker.ietf.org/doc/rfc6455/) protocol.

Features:

* Based upon tnet, reduce memory usage and improve CPU utilization.
* Read/Write for a full message.
* NextMessageReader/NextMessageWriter for user customized read/write.
* Writev for multiple byte slices at server side.
* SetMetadata/GetMetadata to store/retrieve user's private data.
* Customized control frame handler for Ping/Pong/Close.
* Set the message type of the connection to use Read/Write API directly.
* Combined writes optimization to merge header and payload into a single syscall.
* Client handshake header injection and response header hooks.

## Handshake Headers

You can write additional HTTP headers to the client handshake request and inspect
non-WebSocket headers from the handshake response.

Note that the key/value bytes received by callback are only valid until the
callback returns.

To inspect a non-101 handshake response, use `WithClientOnHandshakeResponseStatusError`.

```go
reqHeader := http.Header{}
reqHeader.Set("X-Token", "abc")

opts := []websocket.ClientOption{
    websocket.WithClientHandshakeRequestHeaderHTTP(reqHeader),
    websocket.WithClientOnHandshakeResponseHeader(func(key, value []byte) error {
        // Copy key/value if you need to keep them after this callback returns.
        return nil
    }),
}
conn, err := websocket.Dial(url, opts...)
```

## Combined Writes Optimization

**Note: Requires version >= v0.0.10**

By default, writing a WebSocket message requires two syscalls: one for the frame header and one for the payload data. For small messages, performance can be improved by combining these writes.

With combined writes optimization enabled, the frame header and payload data are merged into a single syscall write, improving performance.

### Usage Recommendations

- Recommended for small message scenarios
- Combined writes mode is disabled by default

### Server Side

```go
opts := []websocket.ServerOption{
    websocket.WithServerCombinedWrites(true),
}
s, err := websocket.NewService(ln, handler, opts...)
```

### Client Side

```go
opts := []websocket.ClientOption{
    websocket.WithClientCombinedWrites(true),
}
conn, err := websocket.Dial(url, opts...)
```
