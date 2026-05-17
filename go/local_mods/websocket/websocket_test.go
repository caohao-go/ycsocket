package websocket_test

import (
	"bytes"
	"context"
	stdtls "crypto/tls"
	"errors"
	"io"
	"log"
	"sync"
	"testing"
	"time"

	"git.woa.com/trpc-go/tnet"
	"git.woa.com/trpc-go/tnet/extensions/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	wsURLPrefix  = "ws://"
	wssURLPrefix = "wss://"
	hello        = []byte("hello")
	world        = []byte("world!")
)

func TestClientHandle(t *testing.T) {
	var conns []websocket.Conn
	done := make(chan struct{})
	addr, cancel := runServer(t, func(conn websocket.Conn) error {
		tp, buf, err := conn.ReadMessage()
		require.Nil(t, err)
		require.Equal(t, websocket.Text, tp)
		require.Equal(t, world, buf)
		return nil
	}, done, websocket.WithHookAfterHandshake(func(ctx context.Context, c websocket.Conn) error {
		conns = append(conns, c)
		return nil
	}))
	clientConn, err := websocket.Dial(wsURLPrefix + addr)
	require.Nil(t, err)
	clientHandle := func(conn websocket.Conn) error {
		tp, buf, err := conn.ReadMessage()
		if err != nil {
			return err
		}
		require.Equal(t, websocket.Text, tp)
		return conn.WriteMessage(websocket.Text, buf)
	}
	clientConn.SetOnRequest(clientHandle)
	clientConn.SetOnClosed(func(c websocket.Conn) error { return nil })
	require.True(t, len(conns) != 0)
	// Push data from server connections.
	for i := range conns {
		require.Nil(t, conns[i].WriteMessage(websocket.Text, world))
	}
	require.Nil(t, clientConn.Close())
	cancel()
	<-done // Wait for server closing.
}

func TestServer(t *testing.T) {
	runTestWithHandles(t, func(conn websocket.Conn) error {
		tp, buf, err := conn.ReadMessage()
		if err != nil {
			return err
		}
		require.Equal(t, websocket.Text, tp)
		require.Nil(t, conn.WriteMessage(websocket.Text, buf))
		return nil
	}, func(conn websocket.Conn) error {
		require.Nil(t, conn.WriteMessage(websocket.Text, hello))
		_, data, err := conn.ReadMessage()
		require.Nil(t, err)
		require.Equal(t, string(hello), string(data))
		return nil
	})
}

func TestServerBinary(t *testing.T) {
	runTestWithHandles(t, func(conn websocket.Conn) error {
		tp, buf, err := conn.ReadMessage()
		if err != nil {
			return err
		}
		require.Equal(t, websocket.Binary, tp)
		require.Nil(t, conn.WriteMessage(websocket.Binary, buf))
		return nil
	}, func(conn websocket.Conn) error {
		require.Nil(t, conn.WriteMessage(websocket.Binary, hello))
		_, data, err := conn.ReadMessage()
		require.Nil(t, err)
		require.Equal(t, string(hello), string(data))
		return nil
	})
}

var testMessageNumber = 10000

func TestConcurrentReadWrite(t *testing.T) {
	readMu := sync.Mutex{}
	writeMu := sync.Mutex{}
	runTestWithHandles(t, func(conn websocket.Conn) error {
		for i := 0; i < 200; i++ {
			go func() {
				for {
					readMu.Lock()
					tp, buf, err := conn.ReadMessage()
					readMu.Unlock()
					if err != nil {
						if errors.Is(err, tnet.ErrConnClosed) {
							return
						}
						t.Log("conn.ReadMessage", err)
					}
					require.Nil(t, err)
					require.Equal(t, websocket.Binary, tp)
					writeMu.Lock()
					require.Nil(t, conn.WriteMessage(websocket.Binary, buf))
					writeMu.Unlock()
				}
			}()
		}
		time.Sleep(time.Second)
		return nil
	}, func(conn websocket.Conn) error {
		for i := 0; i < testMessageNumber; i++ {
			require.Nil(t, conn.WriteMessage(websocket.Binary, hello))
		}
		for i := 0; i < testMessageNumber; i++ {
			_, data, err := conn.ReadMessage()
			require.Nil(t, err)
			require.Equal(t, hello, data)
		}
		return nil
	}, nil)
}

func TestOptions(t *testing.T) {
	runTestWithHandles(t, func(conn websocket.Conn) error {
		assert.NotNil(t, conn.LocalAddr())
		assert.NotNil(t, conn.RemoteAddr())
		assert.Nil(t, conn.SetDeadline(time.Now().Add(time.Second)))
		assert.Nil(t, conn.SetReadDeadline(time.Now().Add(time.Second)))
		assert.Nil(t, conn.SetWriteDeadline(time.Now().Add(time.Second)))
		assert.Nil(t, conn.SetIdleTimeout(time.Second))
		assert.Nil(t, conn.SetReadIdleTimeout(time.Second))
		assert.Nil(t, conn.SetWriteIdleTimeout(time.Second))
		data, ok := conn.GetMetaData().([]byte)
		assert.True(t, ok)
		assert.Equal(t, string(hello), string(data))
		tp, buf, err := conn.ReadMessage()
		if err != nil {
			return err
		}
		assert.Equal(t, websocket.Binary, tp)
		assert.Nil(t, conn.WriteMessage(websocket.Binary, buf))
		return nil
	}, func(conn websocket.Conn) error {
		require.Nil(t, conn.WriteMessage(websocket.Binary, world))
		_, data, err := conn.ReadMessage()
		require.Nil(t, err)
		require.Equal(t, string(world), string(data))
		return nil
	}, nil,
		websocket.WithHookBeforeHandshake(func(ctx context.Context) (context.Context, error) {
			_, ok := websocket.RemoteAddrFromContext(ctx)
			require.True(t, ok)
			_, ok = websocket.UpgraderFromContext(ctx)
			require.True(t, ok)
			return ctx, nil
		}),
		websocket.WithHookAfterHandshake(
			func(ctx context.Context, c websocket.Conn) error {
				_, ok := websocket.LocalAddrFromContext(ctx)
				require.True(t, ok)
				c.SetMetaData(hello)
				return nil
			}),
		websocket.WithNewHandshakeContext(func() context.Context { return context.Background() }),
		websocket.WithIdleTimeout(time.Second),
		websocket.WithOnClosed(func(c websocket.Conn) error { return nil }),
		websocket.WithTCPKeepAlive(time.Second),
	)
}

func TestNextMessageReaderWriter(t *testing.T) {
	runTestWithHandles(t, func(conn websocket.Conn) error {
		tp, r, err := conn.NextMessageReader()
		if errors.Is(err, tnet.ErrConnClosed) {
			return nil
		}
		assert.Equal(t, websocket.Text, tp)
		data, err := io.ReadAll(r)
		assert.Nil(t, err)
		assert.Equal(t, string(hello), string(data))
		assert.Nil(t, conn.WriteMessage(tp, data))
		return nil
	}, func(conn websocket.Conn) error {
		w, err := conn.NextMessageWriter(websocket.Text)
		require.Nil(t, err)
		n, err := w.Write(hello[:2])
		require.Nil(t, err)
		require.Equal(t, 2, n)
		n, err = w.Write(hello[2:])
		require.Nil(t, err)
		require.Equal(t, 3, n)
		require.Nil(t, w.Close())
		tp, data, err := conn.ReadMessage()
		require.Nil(t, err)
		require.Equal(t, websocket.Text, tp)
		require.Equal(t, string(hello), string(data))
		return nil
	}, nil)
}

func TestWritev(t *testing.T) {
	runTestWithHandles(t, func(conn websocket.Conn) error {
		tp, data, err := conn.ReadMessage()
		if errors.Is(err, tnet.ErrConnClosed) {
			return nil
		}
		require.Nil(t, err)
		require.Equal(t, websocket.Binary, tp)
		require.Equal(t, append(hello, world...), data)
		require.Nil(t, conn.WritevMessage(websocket.Binary, hello, world))
		return nil
	}, func(conn websocket.Conn) error {
		require.Nil(t, conn.WritevMessage(websocket.Binary, hello, world))
		tp, data, err := conn.ReadMessage()
		require.Nil(t, err)
		require.Equal(t, websocket.Binary, tp)
		require.Equal(t, append(hello, world...), data)
		return nil
	}, nil)
}

func TestSubProtocols(t *testing.T) {
	runTestWithHandles(t, func(conn websocket.Conn) error {
		require.Equal(t, "superchat", conn.Subprotocol())
		tp, data, err := conn.ReadMessage()
		if errors.Is(err, tnet.ErrConnClosed) {
			return nil
		}
		assert.Nil(t, conn.WriteMessage(tp, data))
		return nil
	}, func(conn websocket.Conn) error {
		require.Equal(t, "superchat", conn.Subprotocol())
		require.Nil(t, conn.WriteMessage(websocket.Text, hello))
		_, data, err := conn.ReadMessage()
		require.Nil(t, err)
		require.Equal(t, hello, data)
		return nil
	}, []websocket.ClientOption{
		websocket.WithTimeout(time.Second),
		websocket.WithSubProtocols([]string{"chat", "superchat"}),
	}, websocket.WithProtocolSelect(func(b []byte) bool {
		switch s := string(b); s {
		case "chat":
			return true
		default:
			return false
		}
	}), websocket.WithProtocolCustom(func(b []byte) (string, bool) {
		if string(b) == "chat, superchat" {
			return "superchat", true
		}
		return "", true
	}))
}

func TestReadWrite(t *testing.T) {
	runTestWithHandles(t, func(conn websocket.Conn) error {
		buf := make([]byte, 6)
		n, err := conn.Read(buf)
		if err != nil {
			t.Log("server read", err)
			return err
		}
		_, err = conn.Write(buf[:n])
		if err != nil {
			t.Log("server write", err)
			return err
		}
		return nil
	}, func(conn websocket.Conn) error {
		n, err := conn.Write(hello)
		require.Nil(t, err)
		require.Equal(t, len(hello), n)
		buf := make([]byte, len(hello))
		n, err = conn.Read(buf)
		require.Nil(t, err)
		require.Equal(t, len(hello), n)
		require.Equal(t, hello, buf)

		n, err = conn.Write(hello)
		require.Nil(t, err)
		require.Equal(t, len(hello), n)
		n, err = conn.Write(world)
		require.Nil(t, err)
		require.Equal(t, len(world), n)

		buf = make([]byte, len(hello)+len(world))
		n, err = conn.Read(buf[:2])
		require.Nil(t, err)
		require.Equal(t, 2, n)
		require.Equal(t, hello[:2], buf[:2])
		n, err = conn.Read(buf[2:5])
		require.Nil(t, err)
		require.Equal(t, 3, n)
		require.Equal(t, hello[2:5], buf[2:5])
		conn.Close()
		return nil
	}, []websocket.ClientOption{
		websocket.WithClientMessageType(websocket.Binary),
	}, websocket.WithServerMessageType(websocket.Binary))
}

func TestReadWriteErrMessageType(t *testing.T) {
	runTestWithHandles(t, func(conn websocket.Conn) error {
		buf := make([]byte, 6)
		_, err := conn.Read(buf)
		require.NotNil(t, err)
		return nil
	}, func(conn websocket.Conn) error {
		require.Nil(t, conn.WriteMessage(websocket.Text, hello))
		time.Sleep(time.Millisecond)
		return nil
	}, []websocket.ClientOption{
		websocket.WithClientMessageType(websocket.Binary),
	}, websocket.WithServerMessageType(websocket.Binary))
}

func TestReadWriteWithNoMessageType(t *testing.T) {
	runTestWithHandles(t, func(conn websocket.Conn) error {
		buf := make([]byte, 6)
		n, err := conn.Read(buf)
		require.NotNil(t, err)
		_, err = conn.Write(buf[:n])
		require.NotNil(t, err)
		return nil
	}, func(conn websocket.Conn) error {
		require.Nil(t, conn.WriteMessage(websocket.Binary, hello))
		_, err := conn.Write(hello)
		require.NotNil(t, err)
		buf := make([]byte, len(hello))
		_, err = conn.Read(buf)
		require.NotNil(t, err)
		return nil
	}, nil)
}

func BenchmarkWriteMessage(b *testing.B) {
	payloads := map[string][]byte{
		"2b":    []byte("hi"),                         // 2 bytes
		"11b":   []byte("hello world"),                // 11 bytes
		"128b":  bytes.Repeat([]byte("hello"), 25),    // ~128 bytes
		"256b":  bytes.Repeat([]byte("hello"), 50),    // ~256 bytes
		"512b":  bytes.Repeat([]byte("hello"), 100),   // ~512 bytes
		"768b":  bytes.Repeat([]byte("hello"), 150),   // ~768 bytes
		"896b":  bytes.Repeat([]byte("hello"), 175),   // ~896 bytes
		"1024b": bytes.Repeat([]byte("hello"), 200),   // ~1024 bytes (1KB)
		"1152b": bytes.Repeat([]byte("hello"), 225),   // ~1152 bytes
		"1280b": bytes.Repeat([]byte("hello"), 250),   // ~1280 bytes
		"1408b": bytes.Repeat([]byte("hello"), 275),   // ~1408 bytes
		"1536b": bytes.Repeat([]byte("hello"), 300),   // ~1536 bytes
		"2048b": bytes.Repeat([]byte("hello"), 400),   // ~2048 bytes (2KB)
		"4k":    bytes.Repeat([]byte("hello"), 800),   // ~4KB
		"8k":    bytes.Repeat([]byte("hello"), 1600),  // ~8KB
		"16k":   bytes.Repeat([]byte("hello"), 3200),  // ~16KB
		"32k":   bytes.Repeat([]byte("hello"), 6400),  // ~32KB
		"64k":   bytes.Repeat([]byte("hello"), 12800), // ~64KB
		"128k":  bytes.Repeat([]byte("hello"), 25600), // ~128KB
	}

	cases := []struct {
		name          string
		combineWrites bool
	}{
		{"legacy_mode", false},
		{"combined_writes", true},
	}

	for size, payload := range payloads {
		b.Run(size, func(b *testing.B) {
			for _, tc := range cases {
				b.Run(tc.name, func(b *testing.B) {
					done := make(chan struct{})
					addr, cancel := runBenchServer(b, func(c websocket.Conn) error {
						io.Copy(io.Discard, c)
						return nil
					}, done)
					defer cancel()

					url := "ws://" + addr
					opts := []websocket.ClientOption{
						websocket.WithClientCombinedWrites(tc.combineWrites),
					}
					c, err := websocket.Dial(url, opts...)
					if err != nil {
						b.Fatal(err)
					}
					defer c.Close()

					b.ResetTimer()
					b.SetBytes(int64(len(payload)))
					for i := 0; i < b.N; i++ {
						if err := c.WriteMessage(websocket.Binary, payload); err != nil {
							b.Fatal(err)
						}
					}
					cancel()
					<-done
				})
			}
		})
	}
}

func runTestWithHandles(
	t *testing.T,
	serverHandle websocket.Handler,
	clientHandle websocket.Handler,
	dialOpts []websocket.ClientOption,
	opts ...websocket.ServerOption,
) {
	runWSTestWithHandles(t, serverHandle, clientHandle, dialOpts, opts...)
	dialOpts = append(dialOpts, websocket.WithClientTLSConfig(&stdtls.Config{InsecureSkipVerify: true}))
	opts = append(opts, websocket.WithServerTLSConfig(getTLSCfg()))
	runWSSTestWithHandles(t, serverHandle, clientHandle, dialOpts, opts...)
}

func getTLSCfg() *stdtls.Config {
	cert, err := stdtls.LoadX509KeyPair("testdata/server.crt", "testdata/server.key")
	if err != nil {
		log.Fatal(err)
	}
	return &stdtls.Config{Certificates: []stdtls.Certificate{cert}}
}

func runWSTestWithHandles(
	t *testing.T,
	serverHandle websocket.Handler,
	clientHandle websocket.Handler,
	dialOpts []websocket.ClientOption,
	opts ...websocket.ServerOption,
) {
	done := make(chan struct{})
	addr, cancel := runServer(t, serverHandle, done, opts...)
	conn, err := websocket.Dial(wsURLPrefix+addr, dialOpts...)
	require.Nil(t, err)
	require.Nil(t, clientHandle(conn))
	cancel()
	<-done
}

func runWSSTestWithHandles(
	t *testing.T,
	serverHandle websocket.Handler,
	clientHandle websocket.Handler,
	dialOpts []websocket.ClientOption,
	opts ...websocket.ServerOption,
) {
	done := make(chan struct{})
	addr, cancel := runServer(t, serverHandle, done, opts...)
	time.Sleep(10 * time.Millisecond)
	conn, err := websocket.Dial(wssURLPrefix+addr, dialOpts...)
	require.Nil(t, err)
	err = clientHandle(conn)
	require.Nil(t, err, "handler err: %+v", err)
	cancel()
	<-done
}

func runServer(
	t *testing.T,
	h websocket.Handler,
	done chan struct{},
	opts ...websocket.ServerOption,
) (string, context.CancelFunc) {
	ln, err := tnet.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Log("tnet.Listen", err)
	}
	require.Nil(t, err)
	s, err := websocket.NewService(ln, h, opts...)
	require.Nil(t, err)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		s.Serve(ctx)
		done <- struct{}{}
	}()
	return ln.Addr().String(), cancel
}

func runBenchServer(
	b *testing.B,
	h websocket.Handler,
	done chan struct{},
	opts ...websocket.ServerOption,
) (string, context.CancelFunc) {
	ln, err := tnet.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		b.Fatal(err)
	}
	s, err := websocket.NewService(ln, h, opts...)
	if err != nil {
		b.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		s.Serve(ctx)
		done <- struct{}{}
	}()
	return ln.Addr().String(), cancel
}
