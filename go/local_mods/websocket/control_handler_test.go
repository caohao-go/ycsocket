package websocket_test

import (
	"strings"
	"testing"
	"time"

	"git.woa.com/trpc-go/tnet/extensions/websocket"
	"github.com/stretchr/testify/require"
)

func TestReadAnyMessage(t *testing.T) {
	runTestWithHandles(t, func(conn websocket.Conn) error {
		tp, buf, err := conn.ReadMessage()
		if err != nil {
			return err
		}
		require.Nil(t, conn.WriteMessage(tp, buf))
		return nil
	}, func(conn websocket.Conn) error {
		require.Nil(t, conn.WriteMessage(websocket.Binary, hello))
		_, _, err := conn.ReadMessage()
		require.Nil(t, err)
		return nil
	})
}

func TestControlHandlers(t *testing.T) {
	runTestWithHandles(t, func(conn websocket.Conn) error {
		tp, buf, err := conn.ReadMessage()
		if err != nil {
			return err
		}
		require.Equal(t, websocket.Binary, tp)
		require.Nil(t, conn.WriteMessage(websocket.Binary, buf))
		return nil
	}, func(conn websocket.Conn) error {
		require.Nil(t, conn.WriteMessage(websocket.Ping, world))
		_, data, err := conn.ReadMessage()
		require.Nil(t, err)
		require.Equal(t, hello, data)

		require.Nil(t, conn.WriteMessage(websocket.Pong, world))
		_, data, err = conn.ReadMessage()
		require.Nil(t, err)
		require.Equal(t, world, data)

		require.Nil(t, conn.WriteMessage(websocket.Binary, hello))
		_, data, err = conn.ReadMessage()
		require.Nil(t, err)
		require.Equal(t, string(hello), string(data))

		return nil
	}, nil, websocket.WithPingHandler(func(c websocket.Conn, b []byte) error {
		require.Equal(t, world, b)
		require.Nil(t, c.WriteMessage(websocket.Binary, hello))
		return nil
	}), websocket.WithPongHandler(func(c websocket.Conn, b []byte) error {
		require.Equal(t, world, b)
		require.Nil(t, c.WriteMessage(websocket.Binary, world))
		return nil
	}))
}

func TestDefaultControlHandlers(t *testing.T) {
	runTestWithHandles(t, func(conn websocket.Conn) error {
		conn.SetPingHandler(nil)
		conn.SetPongHandler(nil)
		tp, buf, err := conn.ReadMessage()
		if err != nil {
			return err
		}
		require.Equal(t, websocket.Binary, tp)
		require.Nil(t, conn.WriteMessage(websocket.Binary, buf))
		return nil
	}, func(conn websocket.Conn) error {
		require.Nil(t, conn.WriteMessage(websocket.Ping, world))
		require.Nil(t, conn.WriteMessage(websocket.Pong, world))
		require.Nil(t, conn.WriteMessage(websocket.Binary, hello))
		_, data, err := conn.ReadMessage()
		require.Nil(t, err)
		require.Equal(t, hello, data)
		return nil
	})
}

func TestNormalClose(t *testing.T) {
	runTestWithHandles(t, func(conn websocket.Conn) error {
		require.Nil(t, conn.Close())
		return nil
	}, func(conn websocket.Conn) error {
		if err := conn.WriteMessage(websocket.Text, hello); err != nil {
			// Ignore `conn is closed` and `broken pipes` errors because the server may have been already closed.
			if strings.Contains(err.Error(), "closed") || strings.Contains(err.Error(), "broken pipe") {
				return nil
			}
			return err
		}
		_, _, err := conn.ReadMessage()
		require.NotNil(t, err)
		time.Sleep(time.Millisecond)
		return nil
	}, nil, websocket.WithOnClosed(func(c websocket.Conn) error { return nil }))
}
