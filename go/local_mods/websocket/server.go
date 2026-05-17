package websocket

import (
	"context"
	"net"

	"git.woa.com/trpc-go/tnet"
	"git.woa.com/trpc-go/tnet/tls"
	"github.com/gobwas/ws"
	"github.com/pkg/errors"
)

// Handler is the websocket connection handler.
type Handler func(c Conn) error

// OnClosed is the websocket connection on closed function.
type OnClosed func(c Conn) error

// NewService creates a new websocket service.
func NewService(ln net.Listener, handler Handler, opts ...ServerOption) (tnet.Service, error) {
	options := defaultServerOptions
	for _, opt := range opts {
		opt(&options)
	}
	if options.tlsConfig != nil {
		return tls.NewService(ln, func(c tls.Conn) error {
			return handleWithOptions(&rawConn{Conn: c}, handler, &options)
		}, tls.WithTCPKeepAlive(options.keepAlive),
			tls.WithServerIdleTimeout(options.idleTimeout),
			tls.WithServerTLSConfig(options.tlsConfig),
			tls.WithOnClosed(onClosedTLS(options.onClosed)),
			tls.WithServerFlushWrite(true), // Enable flush write for websocket.
		)
	}
	return tnet.NewTCPService(ln, func(c tnet.Conn) error {
		return handleWithOptions(c, handler, &options)
	}, tnet.WithTCPKeepAlive(options.keepAlive),
		tnet.WithTCPIdleTimeout(options.idleTimeout),
		tnet.WithOnTCPClosed(onClosed(options.onClosed)),
		tnet.WithFlushWrite(true), // Enable flushwrite for websocket.
	)
}

type graderKey struct{}

// NewContextWithWSUpgrader creates a context with websocket upgrader.
func NewContextWithWSUpgrader(ctx context.Context, grader *ws.Upgrader) context.Context {
	return context.WithValue(ctx, graderKey{}, grader)
}

// UpgraderFromContext extracts a websocket upgrader from context.
func UpgraderFromContext(ctx context.Context) (*ws.Upgrader, bool) {
	g, ok := ctx.Value(graderKey{}).(*ws.Upgrader)
	return g, ok
}

type localAddrKey struct{}

// NewContextWithLocalAddr adds the remote address into context.
func NewContextWithLocalAddr(ctx context.Context, addr net.Addr) context.Context {
	return context.WithValue(ctx, localAddrKey{}, addr)
}

// LocalAddrFromContext extracts the remote address from context.
func LocalAddrFromContext(ctx context.Context) (net.Addr, bool) {
	addr, ok := ctx.Value(localAddrKey{}).(net.Addr)
	return addr, ok
}

type remoteAddrKey struct{}

// NewContextWithRemoteAddr adds the remote address into context.
func NewContextWithRemoteAddr(ctx context.Context, addr net.Addr) context.Context {
	return context.WithValue(ctx, remoteAddrKey{}, addr)
}

// RemoteAddrFromContext extracts the remote address from context.
func RemoteAddrFromContext(ctx context.Context) (net.Addr, bool) {
	addr, ok := ctx.Value(remoteAddrKey{}).(net.Addr)
	return addr, ok
}

func handleWithOptions(c rawConnection, handler Handler, options *serverOptions) error {
	if c.GetMetaData() == nil {
		// Connection has not been upgraded to websocket.
		upgrader := ws.Upgrader{
			Protocol:       options.protocolSelect,
			ProtocolCustom: options.protocolCustom,
		}
		ctx := NewContextWithWSUpgrader(options.newHandshakeContext(), &upgrader)
		ctx = NewContextWithLocalAddr(ctx, c.LocalAddr())
		ctx = NewContextWithRemoteAddr(ctx, c.RemoteAddr())
		ctx, err := options.beforeHandshake(ctx)
		if err != nil {
			return err
		}
		handshake, err := upgrader.Upgrade(c)
		if err != nil {
			return errors.Wrap(err, "websocket upgrade")
		}
		wc := &conn{
			raw:           c,
			role:          ws.StateServerSide,
			subprotocol:   handshake.Protocol,
			pingHandler:   options.pingHandler,
			pongHandler:   options.pongHandler,
			messageType:   options.messageType,
			combineWrites: options.combineWrites,
		}
		c.SetMetaData(wc)
		return options.afterHandshake(ctx, wc)
	}
	wc, ok := c.GetMetaData().(*conn)
	if !ok {
		return errors.New("websocket connection is not stored in metadata")
	}
	return handler(wc)
}

func onClosed(onClosed func(Conn) error) func(tnet.Conn) error {
	if onClosed == nil {
		return func(tnet.Conn) error { return nil }
	}
	return func(c tnet.Conn) error {
		wc, ok := c.GetMetaData().(*conn)
		if !ok {
			return errors.New("on closed: websocket connection is not stored in metadata")
		}
		return onClosed(wc)
	}
}

func onClosedTLS(onClosed func(Conn) error) func(tls.Conn) error {
	if onClosed == nil {
		return func(tls.Conn) error { return nil }
	}
	return func(c tls.Conn) error {
		wc, ok := c.GetMetaData().(*conn)
		if !ok {
			return errors.New("on closed: websocket connection is not stored in metadata")
		}
		return onClosed(wc)
	}
}
