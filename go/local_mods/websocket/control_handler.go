package websocket

import (
	"fmt"
	"io"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

type controlHandler struct {
	*wsutil.ControlHandler
	c *conn
}

func newControlHandler(c *conn, r io.Reader) *controlHandler {
	return &controlHandler{
		ControlHandler: &wsutil.ControlHandler{
			DisableSrcCiphering: true,
			Src:                 r,
			Dst:                 c.raw,
			State:               c.role,
		},
		c: c,
	}
}

func (c *controlHandler) handle(h ws.Header) error {
	switch h.OpCode {
	case ws.OpPing:
		return c.handlePing(h)
	case ws.OpPong:
		return c.handlePong(h)
	case ws.OpClose:
		return c.handleClose(h)
	}
	return fmt.Errorf("not a control frame: %v", h.OpCode)
}

func (c *controlHandler) handlePing(h ws.Header) error {
	if c.c.pingHandler == nil {
		return c.HandlePing(h)
	}
	return c.handleCustomizedPing(h)
}

func (c *controlHandler) handlePong(h ws.Header) error {
	if c.c.pongHandler == nil {
		return c.HandlePong(h)
	}
	return c.handleCustomizedPong(h)
}

func (c *controlHandler) handleClose(h ws.Header) error {
	return c.HandleClose(h)
}

func (c *controlHandler) handleCustomizedPing(h ws.Header) error {
	buf := make([]byte, h.Length) // TODO: optimize
	_, err := io.ReadFull(c.Src, buf)
	if err != nil {
		return err
	}
	return c.c.pingHandler(c.c, buf)
}

func (c *controlHandler) handleCustomizedPong(h ws.Header) error {
	buf := make([]byte, h.Length) // TODO: optimize
	_, err := io.ReadFull(c.Src, buf)
	if err != nil {
		return err
	}
	return c.c.pongHandler(c.c, buf)
}
