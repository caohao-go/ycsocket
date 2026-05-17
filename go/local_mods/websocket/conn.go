package websocket

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"git.woa.com/trpc-go/tnet"
	"git.woa.com/trpc-go/tnet/log"
	"git.woa.com/trpc-go/tnet/tls"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/pkg/errors"
)

// conn implements Conn interface.
type conn struct {
	// mu enables concurrent writes of websocket messages.
	// Although tnet.Conn.Write has its own lock internally, wsutil.WriteMessage needs
	// to call tnet.Conn.Write twice: first to write the header and then to write the body,
	// leading to the need for an additional lock here to ensure that the packet header and
	// the packet body are written to the connection as a whole.
	// Note: This lock does not guarantee concurrent safety when reading message.
	mu            sync.Mutex
	raw           rawConnection
	metaData      interface{}
	role          ws.State    // ws.StateServerSide or ws.StateClientSide.
	source        io.Reader   // It is used as the frame source for reads.
	reader        io.Reader   // connection-wise message type reader for Read.
	messageType   MessageType // connection-wise message type for Read/Write.
	subprotocol   string      // the subprotocol selected during handshake.
	pingHandler   func(c Conn, data []byte) error
	pongHandler   func(c Conn, data []byte) error
	combineWrites bool // Controls if header and payload writes are combined into a single syscall.
}

// rawConnection provides an interface that raw connection inside *websocket.conn
// needs to implement.
type rawConnection interface {
	net.Conn
	// Writev provides multiple data slice write in order.
	Writev(p ...[]byte) (int, error)
	// SetIdleTimeout sets the idle timeout to close connection.
	SetIdleTimeout(d time.Duration) error
	// SetWriteIdleTimeout sets the write idle timeout for closing the connection.
	SetWriteIdleTimeout(d time.Duration) error
	// SetReadIdleTimeout sets the read idle timeout for closing the connection.
	SetReadIdleTimeout(d time.Duration) error
	// SetMetaData sets meta data. Through this method, users can bind some custom data to a connection.
	SetMetaData(interface{})
	// GetMetaData gets meta data.
	GetMetaData() interface{}
}

// rawConn wraps tls.Conn to provide a pseudo Writev implementation.
type rawConn struct {
	tls.Conn
}

// Writev implements websocket.RawConn interface.
func (c *rawConn) Writev(p ...[]byte) (int, error) {
	var num int
	for i := range p {
		n, err := c.Write(p[i])
		if err != nil {
			return num, err
		}
		num += n
	}
	return num, nil
}

// SetReadIdleTimeout implements rawConnection interface.
// tls.Conn does not support read idle timeout, so this is a no-op on non-Linux platforms.
func (c *rawConn) SetReadIdleTimeout(d time.Duration) error {
	return nil
}

// SetWriteIdleTimeout implements rawConnection interface.
// tls.Conn does not support write idle timeout, so this is a no-op on non-Linux platforms.
func (c *rawConn) SetWriteIdleTimeout(d time.Duration) error {
	return nil
}

// Read implements net.Conn. It is used only when message type of this connection
// is set. The connection-wise message type is set by websocket.WithMessageType
// when server is created and cannot be changed thereafter.
// Errors will be returned if message type is not set or the read message type is
// not equal to the set message type.
//
// Not Concurrent safe.
// Do not use this API in multiple goroutines.
func (c *conn) Read(buf []byte) (int, error) {
	if c.messageType != Text && c.messageType != Binary {
		return 0, errors.New("message type is neither Text nor binary for this connection, cannot use Read")
	}
	for {
		if c.reader == nil {
			var (
				err error
				tp  MessageType
			)
			tp, c.reader, err = c.NextMessageReader()
			if err != nil {
				return 0, err
			}
			if tp != c.messageType {
				io.Copy(io.Discard, c.reader) // Discard the mismatch message.
				return 0, fmt.Errorf("inconsistent message type from Read: %s, want %s", tp, c.messageType)
			}
		}
		n, err := c.reader.Read(buf)
		if err == io.EOF {
			c.reader = nil
			if n > 0 {
				return n, nil
			}
			continue
		}
		return n, err
	}
}

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
func (c *conn) ReadAnyMessage() (MessageType, []byte, error) {
	controlHandler := c.newControlFrameHandler()
	rd := c.newReader(controlHandler)
	hdr, err := rd.NextFrame()
	if err != nil {
		return -1, nil, err
	}
	tp := toMessageType[hdr.OpCode]
	bts, err := io.ReadAll(rd)
	return tp, bts, err
}

// ReadMessage reads a complete text or binary data message.
// The returned DataType specifies that it is text or binary.
// The returned type can only be text or binary, because control
// frame are automatically handled by control handlers.
//
// Not Concurrent safe.
// Do not use this API in multiple goroutines.
func (c *conn) ReadMessage() (MessageType, []byte, error) {
	tp, rd, err := c.NextMessageReader()
	if err != nil {
		return tp, nil, err
	}
	bts, err := io.ReadAll(rd)
	return tp, bts, err
}

// NextMessageReader returns a reader to read the next message.
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
func (c *conn) NextMessageReader() (MessageType, io.Reader, error) {
	controlHandler := c.newControlFrameHandler()
	rd := c.newReader(controlHandler)
	for {
		hdr, err := rd.NextFrame()
		if err != nil {
			return toMessageType[hdr.OpCode], nil, err
		}
		// Process control frames and get a next frame.
		if hdr.OpCode.IsControl() {
			if err := controlHandler(hdr, rd); err != nil {
				return toMessageType[hdr.OpCode], nil, err
			}
			continue
		}
		return toMessageType[hdr.OpCode], rd, err
	}
}

// Write implements net.Conn. Connection-wise message type should be set using
// func websocket.WithMessageType in order to use this API. It will write a message
// of connection-wise message type.
// Error will be returned if message type is not set.
//
// Concurrent safe.
// You can use this API in multiple goroutines.
func (c *conn) Write(buf []byte) (int, error) {
	if c.messageType != Text && c.messageType != Binary {
		return 0, errors.New("message type is neither Text nor Binary for this connection")
	}
	if c.combineWrites {
		if err := c.writeMessageCombined(c.messageType, buf); err != nil {
			return 0, err
		}
		return len(buf), nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := c.writeMessage(c.messageType, buf); err != nil {
		return 0, err
	}
	return len(buf), nil
}

// WriteMessage writes a message in a single frame.
//
// Concurrent safe.
// You can use this API in multiple goroutines.
func (c *conn) WriteMessage(tp MessageType, buf []byte) error {
	if c.combineWrites {
		return c.writeMessageCombined(tp, buf)
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.writeMessage(tp, buf)
}

// WritevMessage writes multiple messages in a single frame.
// Note that client side needs to mask the data to form payload,
// therefore writev does not actually work with client side writing.
//
// Concurrent safe.
// You can use this API in multiple goroutines.
func (c *conn) WritevMessage(tp MessageType, p ...[]byte) error {
	if c.role.ClientSide() {
		// Client side does not use writev, because masking needs to be done for client side.
		var payload []byte
		for i := range p {
			payload = append(payload, p[i]...)
		}
		if c.combineWrites {
			return c.writeMessageCombined(tp, payload)
		}
		c.mu.Lock()
		defer c.mu.Unlock()
		return c.writeMessage(tp, payload)
	}
	// Server side can use writev directly
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.serverWritev(tp, p...)
}

func (c *conn) serverWritev(tp MessageType, p ...[]byte) error {
	var length int64
	for i := range p {
		length += int64(len(p[i]))
	}
	h := ws.Header{
		Fin:    true,
		OpCode: toOpCode[tp],
		Length: length,
	}
	if err := ws.WriteHeader(c.raw, h); err != nil {
		return err
	}
	_, err := c.raw.Writev(p...)
	return err
}

func (c *conn) writeMessage(tp MessageType, buf []byte) error {
	return wsutil.WriteMessage(c.raw, c.role, toOpCode[tp], buf)
}

// writeCombined writes message using combined writes optimization.
func (c *conn) writeMessageCombined(tp MessageType, buf []byte) error {
	bw := &bufWriter{
		buf: make([]byte, 0, len(buf)+14), // Pre-allocate space for header (max 14 bytes) and payload.
		w:   c.raw,
	}
	if err := wsutil.WriteMessage(bw, c.role, toOpCode[tp], buf); err != nil {
		return err
	}
	return bw.Flush()
}

// buffer writer
type bufWriter struct {
	buf []byte
	w   io.Writer
}

// Write implements io.Writer interface.
func (bw *bufWriter) Write(p []byte) (n int, err error) {
	bw.buf = append(bw.buf, p...)
	return len(p), nil
}

// Flush writes all buffered data to underlying writer.
func (bw *bufWriter) Flush() error {
	if bw.w == nil {
		return errors.New("writer is nil")
	}
	if len(bw.buf) == 0 {
		return nil
	}
	n, err := bw.w.Write(bw.buf)
	if err != nil {
		return err
	}
	if n < len(bw.buf) {
		return io.ErrShortWrite
	}
	bw.buf = bw.buf[:0]
	return nil
}

// NextMessageWriter return a writer to write the next message.
// A finished message write should end with writer.Close().
//
// Not Concurrent safe.
// Do not use this API in multiple goroutines.
func (c *conn) NextMessageWriter(tp MessageType) (io.WriteCloser, error) {
	return &writeCloser{wsutil.NewWriter(c.raw, c.role, toOpCode[tp])}, nil
}

type writeCloser struct {
	*wsutil.Writer
}

// Close implements io.Closer.
func (w *writeCloser) Close() error {
	return w.Flush()
}

// SetMetaData sets meta data.
func (c *conn) SetMetaData(m interface{}) {
	c.metaData = m
}

// GetMetaData gets meta data.
func (c *conn) GetMetaData() interface{} {
	return c.metaData
}

// Close closes the websocket connection with error code and reason.
func (c *conn) Close() error {
	// Not necessary to call the options.onClose,
	// since options.onClose has already been registered
	// into tnet's onClose.
	return c.raw.Close()
}

// Subprotocol returns the negotiated protocol for the connection.
func (c *conn) Subprotocol() string {
	return c.subprotocol
}

// SetPingHandler sets customized Ping frame handler.
func (c *conn) SetPingHandler(handler func(Conn, []byte) error) {
	c.pingHandler = handler
}

// SetPongHandler sets customized Pong frame handler.
func (c *conn) SetPongHandler(handler func(Conn, []byte) error) {
	c.pongHandler = handler
}

// SetAsyncPingHandler sets customized asynchronous Ping frame handler.
func (c *conn) SetAsyncPingHandler(handler func(Conn, []byte) error) {
	c.pingHandler = func(c Conn, data []byte) error {
		tnet.Submit(func() {
			if err := handler(c, data); err != nil {
				log.Errorf("ping handler handle error: %+v", err)
			}
		})
		return nil
	}
}

// SetAsyncPongHandler sets customized asynchronous Pong frame handler.
func (c *conn) SetAsyncPongHandler(handler func(Conn, []byte) error) {
	c.pongHandler = func(c Conn, data []byte) error {
		tnet.Submit(func() {
			if err := handler(c, data); err != nil {
				log.Errorf("pong handler handle error: %+v", err)
			}
		})
		return nil
	}
}

func (c *conn) newReader(handler wsutil.FrameHandlerFunc) *wsutil.Reader {
	return &wsutil.Reader{
		Source:         c.readSource(),
		State:          c.role,
		CheckUTF8:      true,
		OnIntermediate: handler,
	}
}

func (c *conn) readSource() io.Reader {
	if c.source != nil {
		return c.source
	}
	return c.raw
}

func (c *conn) newControlFrameHandler() wsutil.FrameHandlerFunc {
	return func(h ws.Header, r io.Reader) error {
		return newControlHandler(c, r).handle(h)
	}
}

// SetIdleTimeout sets connection level idle timeout.
func (c *conn) SetIdleTimeout(d time.Duration) error {
	return c.raw.SetIdleTimeout(d)
}

// SetWriteIdleTimeout sets the write idle timeout for closing the connection.
// If d is less than or equal to 0, the idle timeout is disabled.
func (c *conn) SetWriteIdleTimeout(d time.Duration) error {
	return c.raw.SetWriteIdleTimeout(d)
}

// SetReadIdleTimeout sets the read idle timeout for closing the connection.
// If d is less than or equal to 0, the idle timeout is disabled.
func (c *conn) SetReadIdleTimeout(d time.Duration) error {
	return c.raw.SetReadIdleTimeout(d)
}

// LocalAddr returns the local network address, if known.
func (c *conn) LocalAddr() net.Addr {
	return c.raw.LocalAddr()
}

// RemoteAddr returns the remote network address, if known.
func (c *conn) RemoteAddr() net.Addr {
	return c.raw.RemoteAddr()
}

// SetDeadline sets the read and write deadlines associated
// with the connection. It is equivalent to calling both
// SetReadDeadline and SetWriteDeadline.
//
// A deadline is an absolute time after which I/O operations
// fail instead of blocking. The deadline applies to all future
// and pending I/O, not just the immediately following call to
// Read or Write. After a deadline has been exceeded, the
// connection can be refreshed by setting a deadline in the future.
//
// If the deadline is exceeded a call to Read or Write or to other
// I/O methods will return an error that wraps os.ErrDeadlineExceeded.
// This can be tested using errors.Is(err, os.ErrDeadlineExceeded).
// The error's Timeout method will return true, but note that there
// are other possible errors for which the Timeout method will
// return true even if the deadline has not been exceeded.
//
// An idle timeout can be implemented by repeatedly extending
// the deadline after successful Read or Write calls.
//
// A zero value for t means I/O operations will not time out.
func (c *conn) SetDeadline(t time.Time) error {
	return c.raw.SetDeadline(t)
}

// SetReadDeadline sets the deadline for future Read calls
// and any currently-blocked Read call.
// A zero value for t means Read will not time out.
func (c *conn) SetReadDeadline(t time.Time) error {
	return c.raw.SetReadDeadline(t)
}

// SetWriteDeadline sets the deadline for future Write calls
// and any currently-blocked Write call.
// Even if write times out, it may return n > 0, indicating that
// some of the data was successfully written.
// A zero value for t means Write will not time out.
func (c *conn) SetWriteDeadline(t time.Time) error {
	return c.raw.SetWriteDeadline(t)
}

// SetOnRequest sets request handler for websocket connection.
// Typically used by websocket client.
func (c *conn) SetOnRequest(handler Handler) error {
	if tc, ok := c.raw.(tnet.Conn); ok {
		return tc.SetOnRequest(func(conn tnet.Conn) error {
			o := defaultServerOptions
			return handleWithOptions(conn, handler, &o)
		})
	}
	return errors.New("websocket.conn is expected to have raw to be tnet.Conn in SetOnRequest")
}

// SetOnClosed set on closed function for websocket connection.
func (c *conn) SetOnClosed(oc OnClosed) error {
	if tc, ok := c.raw.(tnet.Conn); ok {
		return tc.SetOnClosed(onClosed(oc))
	}
	return errors.New("websocket.conn is expected to have raw to be tnet.Conn in SetOnClosed")
}

var toMessageType = map[ws.OpCode]MessageType{
	ws.OpText:   Text,
	ws.OpBinary: Binary,
	ws.OpPing:   Ping,
	ws.OpPong:   Pong,
	ws.OpClose:  Close,
}

var toOpCode = map[MessageType]ws.OpCode{
	Text:   ws.OpText,
	Binary: ws.OpBinary,
	Ping:   ws.OpPing,
	Pong:   ws.OpPong,
	Close:  ws.OpClose,
}
