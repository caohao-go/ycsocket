package websocket

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCombinedWritesOptions(t *testing.T) {
	t.Run("server option", func(t *testing.T) {
		opts := &serverOptions{}
		WithServerCombinedWrites(true)(opts)
		assert.True(t, opts.combineWrites)
		WithServerCombinedWrites(false)(opts)
		assert.False(t, opts.combineWrites)
	})
	t.Run("client option", func(t *testing.T) {
		opts := &clientOptions{}
		WithClientCombinedWrites(true)(opts)
		assert.True(t, opts.combineWrites)
		WithClientCombinedWrites(false)(opts)
		assert.False(t, opts.combineWrites)
	})
}

func TestHandshakeHeaderWriterWrite(t *testing.T) {
	buf := new(bytes.Buffer)
	w := &handshakeHeaderWriter{w: buf}
	n, err := w.Write([]byte("abc"))
	assert.NoError(t, err)
	assert.Equal(t, 3, n)
	assert.Equal(t, int64(3), w.n)
	assert.Equal(t, "abc", buf.String())
}
