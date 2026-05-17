package websocket_test

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"git.woa.com/trpc-go/tnet/extensions/websocket"
	"github.com/stretchr/testify/require"
)

type handshakeObservation struct {
	wsKey  string
	xToken string
	err    error
}

type staticHandshakeHeader string

func (h staticHandshakeHeader) WriteTo(w io.Writer) (int64, error) {
	n, err := io.WriteString(w, string(h))
	return int64(n), err
}

func TestClientHandshakeRequestHeader(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = ln.Close() })
	serverDone := make(chan struct{})
	t.Cleanup(func() { close(serverDone) })
	obsCh := make(chan handshakeObservation, 1)
	go func() {
		obs := handshakeObservation{}
		c, err := ln.Accept()
		if err != nil {
			obs.err = err
			obsCh <- obs
			return
		}
		defer c.Close()
		br := bufio.NewReader(c)
		if _, err := br.ReadString('\n'); err != nil {
			obs.err = err
			obsCh <- obs
			return
		}
		for {
			line, err := br.ReadString('\n')
			if err != nil {
				obs.err = err
				obsCh <- obs
				return
			}
			line = strings.TrimRight(line, "\r\n")
			if line == "" {
				break
			}
			key, value, ok := strings.Cut(line, ":")
			if !ok {
				continue
			}
			key = strings.TrimSpace(key)
			value = strings.TrimSpace(value)
			switch strings.ToLower(key) {
			case "sec-websocket-key":
				obs.wsKey = value
			case "x-token":
				obs.xToken = value
			}
		}
		if obs.wsKey == "" {
			obs.err = errors.New("missing Sec-WebSocket-Key header")
			obsCh <- obs
			return
		}
		accept := computeWebSocketAccept(obs.wsKey)
		resp := "HTTP/1.1 101 Switching Protocols\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Accept: " + accept + "\r\n\r\n"
		if _, err := c.Write([]byte(resp)); err != nil {
			obs.err = err
		}
		obsCh <- obs
		select {
		case <-serverDone:
		case <-time.After(time.Second):
		}
	}()
	c, err := websocket.Dial("ws://"+ln.Addr().String()+"/", websocket.WithClientHandshakeRequestHeader(staticHandshakeHeader("X-Token: abc\r\n")))
	require.NoError(t, err)
	require.NoError(t, c.Close())
	obs := <-obsCh
	require.NoError(t, obs.err)
	require.Equal(t, "abc", obs.xToken)
}

func TestClientHandshakeHeaderAndBufferedBytes(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = ln.Close() })
	serverDone := make(chan struct{})
	t.Cleanup(func() { close(serverDone) })
	obsCh := make(chan handshakeObservation, 1)
	go func() {
		obs := handshakeObservation{}
		c, err := ln.Accept()
		if err != nil {
			obs.err = err
			obsCh <- obs
			return
		}
		defer c.Close()
		br := bufio.NewReader(c)
		if _, err := br.ReadString('\n'); err != nil {
			obs.err = err
			obsCh <- obs
			return
		}
		for {
			line, err := br.ReadString('\n')
			if err != nil {
				obs.err = err
				obsCh <- obs
				return
			}
			line = strings.TrimRight(line, "\r\n")
			if line == "" {
				break
			}
			key, value, ok := strings.Cut(line, ":")
			if !ok {
				continue
			}
			key = strings.TrimSpace(key)
			value = strings.TrimSpace(value)
			switch strings.ToLower(key) {
			case "sec-websocket-key":
				obs.wsKey = value
			case "x-token":
				obs.xToken = value
			}
		}
		if obs.wsKey == "" {
			obs.err = errors.New("missing Sec-WebSocket-Key header")
			obsCh <- obs
			return
		}
		accept := computeWebSocketAccept(obs.wsKey)
		resp := "HTTP/1.1 101 Switching Protocols\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Accept: " + accept + "\r\nX-Server: ok\r\n\r\n"
		payload := bytes.Repeat([]byte("a"), 8192)
		frame := buildTextFrame(payload)
		if _, err := c.Write(append([]byte(resp), frame...)); err != nil {
			obs.err = err
		}
		obsCh <- obs
		select {
		case <-serverDone:
		case <-time.After(time.Second):
		}
	}()
	reqHeader := http.Header{}
	reqHeader.Set("X-Token", "abc")
	var gotServerHeader string
	c, err := websocket.Dial("ws://"+ln.Addr().String()+"/", websocket.WithClientHandshakeRequestHeaderHTTP(reqHeader), websocket.WithClientOnHandshakeResponseHeader(func(key, value []byte) error {
		if strings.EqualFold(string(key), "x-server") {
			gotServerHeader = string(value)
		}
		return nil
	}))
	require.NoError(t, err)
	require.Equal(t, "ok", gotServerHeader)
	require.NoError(t, c.SetReadDeadline(time.Now().Add(time.Second)))
	tp, msg, err := c.ReadMessage()
	require.NoError(t, err)
	require.Equal(t, websocket.Text, tp)
	require.Equal(t, 8192, len(msg))
	require.Equal(t, byte('a'), msg[0])
	require.Equal(t, byte('a'), msg[len(msg)-1])
	require.NoError(t, c.Close())
	obs := <-obsCh
	require.NoError(t, obs.err)
	require.Equal(t, "abc", obs.xToken)
}

type statusErrorObservation struct {
	status int
	reason string
	resp   string
}

func TestClientOnHandshakeResponseStatusError(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = ln.Close() })
	statusCh := make(chan statusErrorObservation, 1)
	go func() {
		c, err := ln.Accept()
		if err != nil {
			return
		}
		defer c.Close()
		br := bufio.NewReader(c)
		for {
			line, err := br.ReadString('\n')
			if err != nil {
				return
			}
			if line == "\r\n" {
				break
			}
		}
		body := "nope"
		resp := "HTTP/1.1 403 Forbidden\r\nContent-Type: text/plain\r\nContent-Length: 4\r\nConnection: close\r\n\r\n" + body
		_, _ = c.Write([]byte(resp))
	}()
	_, err = websocket.Dial("ws://"+ln.Addr().String()+"/", websocket.WithClientOnHandshakeResponseStatusError(func(status int, reason []byte, resp io.Reader) {
		b, _ := io.ReadAll(io.LimitReader(resp, 1024))
		statusCh <- statusErrorObservation{status: status, reason: string(reason), resp: string(b)}
	}))
	require.Error(t, err)
	var obs statusErrorObservation
	require.Eventually(t, func() bool {
		select {
		case obs = <-statusCh:
			return true
		default:
			return false
		}
	}, time.Second, 10*time.Millisecond)
	require.Equal(t, 403, obs.status)
	require.Contains(t, obs.reason, "Forbidden")
	require.Contains(t, obs.resp, "HTTP/1.1 403 Forbidden")
	require.Contains(t, obs.resp, "nope")
}

func computeWebSocketAccept(key string) string {
	sum := sha1.Sum([]byte(key + "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"))
	return base64.StdEncoding.EncodeToString(sum[:])
}

func buildTextFrame(payload []byte) []byte {
	n := len(payload)
	if n <= 125 {
		frame := make([]byte, 2, 2+n)
		frame[0] = 0x81
		frame[1] = byte(n)
		return append(frame, payload...)
	}
	if n <= 65535 {
		frame := make([]byte, 4, 4+n)
		frame[0] = 0x81
		frame[1] = 126
		binary.BigEndian.PutUint16(frame[2:], uint16(n))
		return append(frame, payload...)
	}
	frame := make([]byte, 10, 10+n)
	frame[0] = 0x81
	frame[1] = 127
	binary.BigEndian.PutUint64(frame[2:], uint64(n))
	return append(frame, payload...)
}
