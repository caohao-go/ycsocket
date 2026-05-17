// Package websocket provides websocket connection interface.
package websocket

import (
	"io"
	"net"
	"time"
)

// MessageType specifies message types.
type MessageType int

// Message types.
const (
	Text MessageType = iota + 1
	Binary
	Ping
	Pong
	Close
)

// String implements fmt.Stringer.
func (m MessageType) String() string {
	switch m {
	case Text:
		return "Text"
	case Binary:
		return "Binary"
	case Ping:
		return "Ping"
	case Pong:
		return "Pong"
	case Close:
		return "Close"
	default:
		return "Invalid"
	}
}

// Conn provides websocket connection interface.
type Conn interface {
	net.Conn
	// ReadAnyMessage reads a data message of any kinds.
	//
	// This API is now in favor of the ReadMessage method, which may lead to
	// blocking goroutines in the scenario of interleaved control and data frames.
	//
	// Note: The control frame returned by this method will not be automatically
	// handled by the default or customized control frame handlers.
	//
	// Not Concurrent safe.
	// Do not use this API in multiple goroutines.
	ReadAnyMessage() (MessageType, []byte, error)

	// ReadMessage reads a complete text or binary data message.
	//
	// Note: ReadAnyMessage is now in favor of this method.
	//
	// The returned type can only be text or binary, (i.e. websocket.Text or
	// websocket.Binary), because control frames are automatically handled by
	// control handlers.
	//
	// For control frame types like websocket.Ping or websocket.Pong, please use
	// conn.SetPingHandler or conn.SetPongHandler instead.
	// Caveat: Please read carefully through https://tools.ietf.org/html/rfc6455
	// and make sure you understand the default operations required by the rfc
	// for control frames. Only after that should you consider customizing your
	// own control frame handling logic.
	//
	// Not Concurrent safe.
	// Do not use this API in multiple goroutines.
	ReadMessage() (MessageType, []byte, error)
	// NextMessageReader returns a reader to read the next message.
	//
	// The returned type can only be text or binary, (i.e. websocket.Text or
	// websocket.Binary), because control frames are automatically handled by
	// control handlers.
	//
	// For control frame types like websocket.Ping or websocket.Pong, please use
	// conn.SetPingHandler or conn.SetPongHandler instead.
	// Caveat: Please read carefully through https://tools.ietf.org/html/rfc6455
	// and make sure you understand the default operations required by the rfc
	// for control frames. Only after that should you consider customizing your
	// own control frame handling logic.
	//
	// If you want to use websocket as a plain byte stream protocol,
	// try to wrap it with a customized connection and re-create a
	// new reader in a loop to retrieve the payload bytes inside
	// multiple messages:
	//
	//	type customizedConn struct {
	//		c      *websocket.conn
	//		reader io.Reader
	//	}
	//
	//	// Read implements io.Reader, treats websocket protocol as a plain
	//	// binary data stream protocol. The returned data type must be of type binary.
	//	// Payloads from multiple messages can be read to fill the given buffer p.
	//	func (c *customizedConn) Read(p []byte) (int, error) {
	//		var (
	//			tp  websocket.MessageType
	//			err error
	//		)
	//		tp, c.reader, err = c.c.NextMessageReader()
	//		if err != nil {
	//			return 0, err
	//		}
	//		if tp != Binary {
	//			return 0, errors.New("must read binary data type from Read")
	//		}
	//		for {
	//			if c.reader == nil {
	//				var err error
	//				_, c.reader, err = c.c.NextMessageReader()
	//				if err != nil {
	//					return 0, err
	//				}
	//			}
	//			n, err := c.reader.Read(p)
	//			if err == io.EOF {
	//				c.reader = nil
	//				if n > 0 {
	//					return n, nil
	//				}
	//				continue
	//			}
	//			return n, err
	//		}
	//	}
	//
	// Not Concurrent safe.
	// Do not use this API in multiple goroutines.
	NextMessageReader() (MessageType, io.Reader, error)
	// WriteMessage writes a message in a single frame.
	//
	// The MessageType is recommended to be websocket.Text or websocket.Binary.
	// For control frame types like websocket.Ping or websocket.Pong, please use
	// conn.SetPingHandler or conn.SetPongHandler instead.
	// Caveat: Please read carefully through https://tools.ietf.org/html/rfc6455
	// and make sure you understand the default operations required by the rfc
	// for control frames. Only after that should you consider customizing your
	// own control frame handling logic.
	//
	// Concurrent safe.
	// You can use this API in multiple goroutines.
	WriteMessage(MessageType, []byte) error
	// WritevMessage writes multiple messages in a single frame.
	// Note that client side needs to mask the data to form payload,
	// therefore writev does not actually work with client side writing.
	//
	// The MessageType is recommended to be websocket.Text or websocket.Binary.
	// For control frame types like websocket.Ping or websocket.Pong, please use
	// conn.SetPingHandler or conn.SetPongHandler instead.
	// Caveat: Please read carefully through https://tools.ietf.org/html/rfc6455
	// and make sure you understand the default operations required by the rfc
	// for control frames. Only after that should you consider customizing your
	// own control frame handling logic.
	//
	// Concurrent safe.
	// You can use this API in multiple goroutines.
	WritevMessage(MessageType, ...[]byte) error
	// NextMessageWriter return a writer to write the next message.
	// A finished message write should end with writer.Close().
	//
	// The MessageType is recommended to be websocket.Text or websocket.Binary.
	// For control frame types like websocket.Ping or websocket.Pong, please use
	// conn.SetPingHandler or conn.SetPongHandler instead.
	// Caveat: Please read carefully through https://tools.ietf.org/html/rfc6455
	// and make sure you understand the default operations required by the rfc
	// for control frames. Only after that should you consider customizing your
	// own control frame handling logic.
	//
	// Not Concurrent safe.
	// Do not use this API in multiple goroutines.
	NextMessageWriter(MessageType) (io.WriteCloser, error)
	// SetMetaData sets metadata. Through this method, users can bind some custom data to a connection.
	SetMetaData(interface{})
	// GetMetaData gets meta data.
	GetMetaData() interface{}
	// Subprotocol returns the negotiated protocol for the connection.
	Subprotocol() string
	// SetPingHandler sets customized Ping frame handler.
	//
	// Caveat: Please read carefully through https://tools.ietf.org/html/rfc6455
	// and make sure you understand the default operations required by the rfc
	// for control frames. Only after that should you consider customizing your
	// own control frame handling logic.
	SetPingHandler(handler func(Conn, []byte) error)
	// SetPongHandler sets customized Pong frame handler.
	//
	// Caveat: Please read carefully through https://tools.ietf.org/html/rfc6455
	// and make sure you understand the default operations required by the rfc
	// for control frames. Only after that should you consider customizing your
	// own control frame handling logic.
	SetPongHandler(handler func(Conn, []byte) error)
	// SetAsyncPingHandler sets customized asynchronous Ping frame handler.
	//
	// Caveat: Please read carefully through https://tools.ietf.org/html/rfc6455
	// and make sure you understand the default operations required by the rfc
	// for control frames. Only after that should you consider customizing your
	// own control frame handling logic.
	SetAsyncPingHandler(handler func(Conn, []byte) error)
	// SetAsyncPongHandler sets customized asynchronous Pong frame handler.
	//
	// Caveat: Please read carefully through https://tools.ietf.org/html/rfc6455
	// and make sure you understand the default operations required by the rfc
	// for control frames. Only after that should you consider customizing your
	// own control frame handling logic.
	SetAsyncPongHandler(handler func(Conn, []byte) error)
	// SetIdleTimeout sets connection level idle timeout.
	SetIdleTimeout(time.Duration) error
	// SetWriteIdleTimeout sets the write idle timeout for closing the connection.
	SetWriteIdleTimeout(d time.Duration) error
	// SetReadIdleTimeout sets the read idle timeout for closing the connection.
	SetReadIdleTimeout(d time.Duration) error
	// SetOnRequest sets request handler for websocket connection.
	// Typically used by websocket client.
	SetOnRequest(handle Handler) error
	// SetOnClosed set on closed function for websocket connection.
	SetOnClosed(handle OnClosed) error
}

// HandshakeHeader writes additional HTTP headers during the WebSocket handshake.
type HandshakeHeader interface {
	io.WriterTo
}
