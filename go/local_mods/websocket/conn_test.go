package websocket

import (
	"bytes"
	"io"
	"net"
	"testing"
	"time"

	"github.com/gobwas/ws"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestConnSetErr(t *testing.T) {
	var rc rawConnection
	c := &conn{
		raw: rc,
	}
	require.NotNil(t, c.SetOnRequest(nil))
	require.NotNil(t, c.SetOnClosed(func(c Conn) error { return nil }))
}

func TestWriteMessage(t *testing.T) {
	tests := []struct {
		name          string
		role          ws.State
		combineWrites bool
		payload       []byte
		writeErr      error
		wantWrites    int
		wantErr       bool
	}{
		{
			name:          "server success",
			role:          ws.StateServerSide,
			combineWrites: true,
			payload:       []byte("hello"),
			wantWrites:    1,
		},
		{
			name:          "client success",
			role:          ws.StateClientSide,
			combineWrites: true,
			payload:       []byte("hello"),
			wantWrites:    1,
		},
		{
			name:          "write error",
			role:          ws.StateServerSide,
			combineWrites: true,
			payload:       []byte("hello"),
			writeErr:      errors.New("write failed"),
			wantWrites:    1,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockConn{err: tt.writeErr}
			c := &conn{
				raw:           mock,
				role:          tt.role,
				combineWrites: tt.combineWrites,
			}

			err := c.WriteMessage(Binary, tt.payload)
			if (err != nil) != tt.wantErr {
				t.Errorf("WriteMessage() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.writeErr != nil && err.Error() != tt.writeErr.Error() {
				t.Errorf("WriteMessage() error = %v, want %v", err, tt.writeErr)
			}

			if mock.writeCount != tt.wantWrites {
				t.Errorf("WriteMessage() writes = %v, want %v", mock.writeCount, tt.wantWrites)
			}
		})
	}
}

func TestBufWriter(t *testing.T) {
	tests := []struct {
		name      string
		writes    [][]byte
		writer    io.Writer
		wantErr   error
		wantBytes []byte
	}{
		{
			name: "single write success",
			writes: [][]byte{
				[]byte("hello"),
			},
			writer:    newMockWriter(nil),
			wantBytes: []byte("hello"),
		},
		{
			name: "nil writer",
			writes: [][]byte{
				[]byte("hello"),
			},
			writer:  nil,
			wantErr: errors.New("writer is nil"),
		},
		{
			name: "write error",
			writes: [][]byte{
				[]byte("hello"),
			},
			writer:  newMockWriter(errors.New("write failed")),
			wantErr: errors.New("write failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotBytes []byte
			bw := &bufWriter{w: tt.writer}
			if tt.writer != nil {
				if mw, ok := tt.writer.(*mockWriter); ok {
					mw.onWrite = func(p []byte) (int, error) {
						if mw.err != nil {
							return 0, mw.err
						}
						gotBytes = append(gotBytes, p...)
						return len(p), nil
					}
				}
			}

			// Write all data
			for _, data := range tt.writes {
				n, err := bw.Write(data)
				if err != nil {
					t.Fatalf("Write() error = %v", err)
				}
				if n != len(data) {
					t.Errorf("Write() n = %v, want %v", n, len(data))
				}
			}

			// Flush and check error
			err := bw.Flush()
			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("Flush() error = nil, wantErr %v", tt.wantErr)
				} else if err.Error() != tt.wantErr.Error() {
					t.Errorf("Flush() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("Flush() unexpected error = %v", err)
			}

			if !bytes.Equal(gotBytes, tt.wantBytes) {
				t.Errorf("got bytes = %v, want %v", gotBytes, tt.wantBytes)
			}
		})
	}
}

// mockWriter 增加错误字段
type mockWriter struct {
	onWrite func(p []byte) (int, error)
	err     error
}

func newMockWriter(err error) *mockWriter {
	return &mockWriter{err: err}
}

func (m *mockWriter) Write(p []byte) (n int, err error) {
	if m.onWrite != nil {
		return m.onWrite(p)
	}
	if m.err != nil {
		return 0, m.err
	}
	return len(p), nil
}

// mockConn 增加错误字段
type mockConn struct {
	writeCount int
	err        error
}

func (m *mockConn) Write(p []byte) (n int, err error) {
	m.writeCount++
	if m.err != nil {
		return 0, m.err
	}
	return len(p), nil
}

// Implement other required methods of rawConnection interface...
func (m *mockConn) Read(b []byte) (n int, err error)          { return 0, nil }
func (m *mockConn) Close() error                              { return nil }
func (m *mockConn) LocalAddr() net.Addr                       { return nil }
func (m *mockConn) RemoteAddr() net.Addr                      { return nil }
func (m *mockConn) SetDeadline(t time.Time) error             { return nil }
func (m *mockConn) SetReadDeadline(t time.Time) error         { return nil }
func (m *mockConn) SetWriteDeadline(t time.Time) error        { return nil }
func (m *mockConn) Writev(p ...[]byte) (int, error)           { return 0, nil }
func (m *mockConn) SetIdleTimeout(d time.Duration) error      { return nil }
func (m *mockConn) SetWriteIdleTimeout(d time.Duration) error { return nil }
func (m *mockConn) SetReadIdleTimeout(d time.Duration) error  { return nil }
func (m *mockConn) SetMetaData(interface{})                   {}
func (m *mockConn) GetMetaData() interface{}                  { return nil }
