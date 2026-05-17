package websocket

import (
	"bytes"
	stdtls "crypto/tls"
	"fmt"
	"io"
	neturl "net/url"
	"strings"

	"git.woa.com/trpc-go/tnet"
	"git.woa.com/trpc-go/tnet/tls"
	"github.com/gobwas/ws"
)

// Dial creates a client side connection of websocket.
// url is like: "ws://hostname:port/xxx".
func Dial(url string, opts ...ClientOption) (Conn, error) {
	u, err := neturl.ParseRequestURI(url)
	if err != nil {
		return nil, err
	}
	var options clientOptions
	options.setDefaults()
	for _, opt := range opts {
		opt(&options)
	}
	c, err := dial(u, &options)
	if err != nil {
		return nil, err
	}
	dialer := ws.Dialer{
		Protocols:     options.subprotocols,
		Header:        options.handshakeHeader,
		OnHeader:      options.onHandshakeHeader,
		OnStatusError: options.onHandshakeStatusError,
	}
	br, handshake, err := dialer.Upgrade(c, u)
	if err != nil {
		return nil, err
	}
	var source io.Reader
	if br != nil {
		n := br.Buffered()
		if n > 0 {
			prefix := make([]byte, n)
			if _, err := io.ReadFull(br, prefix); err != nil {
				ws.PutReader(br)
				return nil, err
			}
			source = io.MultiReader(bytes.NewReader(prefix), c)
		}
		ws.PutReader(br)
	}
	wc := &conn{
		raw:           c,
		role:          ws.StateClientSide,
		source:        source,
		subprotocol:   handshake.Protocol,
		messageType:   options.messageType,
		combineWrites: options.combineWrites,
	}
	c.SetMetaData(wc)
	return wc, nil
}

func dial(u *neturl.URL, options *clientOptions) (rawConnection, error) {
	const (
		httpPort  = ":80"
		httpsPort = ":443"
	)
	switch u.Scheme {
	case "ws":
		_, addr := createHostPortAddress(u.Host, httpPort)
		c, err := tnet.DialTCP("tcp", addr, options.timeout)
		if err != nil {
			return nil, err
		}
		c.SetFlushWrite(true) // Enable flush write for websocket.
		return c, nil
	case "wss":
		hostname, addr := createHostPortAddress(u.Host, httpsPort)
		if options.tlsConfig == nil {
			options.tlsConfig = &stdtls.Config{}
		}
		if options.tlsConfig.ServerName == "" {
			options.tlsConfig.ServerName = hostname
		}
		c, err := tls.Dial("tcp", addr, tls.WithClientTLSConfig(options.tlsConfig))
		if err != nil {
			return nil, err
		}
		c.SetFlushWrite(true) // Enable flush write for websocket.
		return &rawConn{Conn: c}, nil
	default:
		return nil, fmt.Errorf("unexpected websocket scheme: %q", u.Scheme)
	}
}

func createHostPortAddress(host string, defaultPort string) (hostname string, addr string) {
	// Handle ipv6 case like:
	// ldap://[2001:db8::7]:89/
	var (
		colon   = strings.LastIndexByte(host, ':')
		bracket = strings.IndexByte(host, ']')
	)
	if colon > bracket {
		return host[:colon], host
	}
	return host, host + defaultPort
}
