// Copyright (c) 2024 The horm-database Authors. All rights reserved.
// This file Author:  CaoHao <18500482693@163.com> .
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package codec

import (
	"context"
	"net"
	"strings"
	"sync"
	"time"

	"server_golang/common/errs"
)

// Msg is the context information for a request
type Msg struct {
	context           context.Context
	frameCodec        interface{}
	requestTimeout    time.Duration
	serializationType int
	callerServiceName string
	callerMethod      string
	calleeServiceName string
	calleeMethod      string
	callRPCName       string
	serverRespError   error
	clientRespError   error
	serverReqHead     interface{}
	serverRespHead    interface{}
	clientReqHead     interface{}
	clientRespHead    interface{}
	localAddr         net.Addr
	remoteAddr        net.Addr
	logSeq            int
	namespace         string
	env               string
	requestID         uint64
	spanID            uint64
	traceID           string
}

const MsgCtxKey = "CTX_MSG"

// NewMessage create an empty message, and put it into ctx,
func NewMessage(ctx context.Context) (context.Context, *Msg) {
	m := msgPool.Get().(*Msg)
	ctx = context.WithValue(ctx, MsgCtxKey, m)
	m.context = ctx
	return ctx, m
}

// Message returns the message of context.
func Message(ctx context.Context) *Msg {
	val := ctx.Value(MsgCtxKey)
	m, ok := val.(*Msg)
	if !ok {
		return &Msg{context: ctx}
	}
	return m
}

func NewAsyncMessage(ctx context.Context, duration time.Duration) (context.Context, context.CancelFunc, *Msg) {
	tCtx, tCancel := context.WithTimeout(context.Background(), duration)

	asyncCtx, asyncMsg := NewMessage(tCtx)
	CopyMsg(asyncMsg, Message(ctx))

	return asyncCtx, tCancel, asyncMsg
}

// CloneContext copies the context and gets a context that retains the value and does not cancel,
// which is used for asynchronous processing by handler, leaving the original timeout control
// and retains the original context information.
//
// After the rpc handler function returns, ctx will be canceled, and put the ctx's Msg back into pool,
// and the associated Metrics and log will be released.
//
// When the handler function runs asynchronously,
// this method needs to be called before starting goroutine to copy the context,
// leaving the original timeout control, and retains the information in Msg for Metrics.
//
// Retain the log context for printing the associated log,
// keep other value in context, such as tracing context, etc.
func CloneContext(ctx context.Context) context.Context {
	oldMsg := Message(ctx)
	newCtx, newMsg := NewMessage(detach(ctx))
	CopyMsg(newMsg, oldMsg)
	return newCtx
}

var msgPool = sync.Pool{
	New: func() interface{} {
		return &Msg{}
	},
}

// RecycleMessage reset message, then put it to pool.
func RecycleMessage(msg *Msg) {
	msg.reset()
	msgPool.Put(msg)
}

func (m *Msg) reset() {
	m.context = nil
	m.frameCodec = nil
	m.requestTimeout = 0
	m.serializationType = 0
	m.callerServiceName = ""
	m.callerMethod = ""
	m.calleeServiceName = ""
	m.calleeMethod = ""
	m.callRPCName = ""
	m.serverRespError = nil
	m.clientRespError = nil
	m.serverReqHead = nil
	m.serverRespHead = nil
	m.clientReqHead = nil
	m.clientRespHead = nil
	m.localAddr = nil
	m.remoteAddr = nil
	//m.logger = nil
	m.namespace = ""
	m.env = ""
	m.requestID = 0
	m.logSeq = 0
}

// Context return context in message
func (m *Msg) Context() context.Context {
	return m.context
}

// Namespace returns namespace.
func (m *Msg) Namespace() string {
	return m.namespace
}

// Env returns environment.
func (m *Msg) Env() string {
	return m.env
}

// WithNamespace sets namespace.
func (m *Msg) WithNamespace(namespace string) {
	m.namespace = namespace
}

// WithEnv sets environment.
func (m *Msg) WithEnv(env string) {
	m.env = env
}

// RemoteAddr returns remote address.
func (m *Msg) RemoteAddr() net.Addr {
	return m.remoteAddr
}

// WithRemoteAddr sets remote address.
func (m *Msg) WithRemoteAddr(addr net.Addr) {
	m.remoteAddr = addr
}

// LocalAddr returns local address.
func (m *Msg) LocalAddr() net.Addr {
	return m.localAddr
}

// WithLocalAddr set local address.
func (m *Msg) WithLocalAddr(addr net.Addr) {
	m.localAddr = addr
}

// RequestTimeout returns request timeout set by upstream business protocol.
func (m *Msg) RequestTimeout() time.Duration {
	return m.requestTimeout
}

// WithRequestTimeout sets request timeout.
func (m *Msg) WithRequestTimeout(t time.Duration) {
	m.requestTimeout = t
}

// FrameCodec returns frame codec.
func (m *Msg) FrameCodec() interface{} {
	return m.frameCodec
}

// WithFrameCodec sets frame codec.
func (m *Msg) WithFrameCodec(f interface{}) {
	m.frameCodec = f
}

// SerializationType returns the value of body serialization.
func (m *Msg) SerializationType() int {
	return m.serializationType
}

// WithSerializationType sets body serialization type of body.
func (m *Msg) WithSerializationType(t int) {
	m.serializationType = t
}

// CallerServiceName returns caller service name.
func (m *Msg) CallerServiceName() string {
	return m.callerServiceName
}

// WithCallerServiceName sets caller servie name.
func (m *Msg) WithCallerServiceName(s string) {
	m.callerServiceName = s
}

// CallerMethod returns callee method.
func (m *Msg) CallerMethod() string {
	return m.callerMethod
}

// WithCallerMethod sets callee method.
func (m *Msg) WithCallerMethod(s string) {
	m.callerMethod = s
}

// CalleeServiceName returns callee service name.
func (m *Msg) CalleeServiceName() string {
	return m.calleeServiceName
}

// WithCalleeServiceName sets callee service name.
func (m *Msg) WithCalleeServiceName(s string) {
	m.calleeServiceName = s
}

// CalleeMethod returns callee method.
func (m *Msg) CalleeMethod() string {
	return m.calleeMethod
}

// WithCalleeMethod sets callee method.
func (m *Msg) WithCalleeMethod(s string) {
	m.calleeMethod = s
}

// CallRPCName returns call rpc name.
func (m *Msg) CallRPCName() string {
	return m.callRPCName
}

// WithCallRPCName sets call rpc name.
func (m *Msg) WithCallRPCName(s string) {
	if m.callRPCName == s {
		return
	}

	m.callRPCName = s

	if strings.Count(s, "/") > 0 {
		i := strings.Index(s, "/") + 1
		m.WithCalleeMethod(s[i:])
		m.calleeServiceName = s[:i-1]
	} else {
		m.calleeServiceName = s
	}
}

// ServerRespError returns server response error.
func (m *Msg) ServerRespError() *errs.Error {
	if m.serverRespError == nil {
		return nil
	}
	e, ok := m.serverRespError.(*errs.Error)
	if !ok {
		return &errs.Error{
			Type: errs.ETypeSystem,
			Code: errs.ErrUnknown,
			Msg:  m.serverRespError.Error(),
		}
	}
	return e
}

// WithServerRespError sets server response error.
func (m *Msg) WithServerRespError(e error) {
	m.serverRespError = e
}

// ClientRespError returns client response error, which created when client call downstream.
func (m *Msg) ClientRespError() error {
	return m.clientRespError
}

// WithClientRespError sets client response err
func (m *Msg) WithClientRespError(e error) {
	m.clientRespError = e
}

// ServerReqHead returns the package head of request
func (m *Msg) ServerReqHead() interface{} {
	return m.serverReqHead
}

// WithServerReqHead sets the package head of request
func (m *Msg) WithServerReqHead(h interface{}) {
	m.serverReqHead = h
}

// ServerRespHead returns the package head of response
func (m *Msg) ServerRespHead() interface{} {
	return m.serverRespHead
}

// WithServerRespHead sets the package head returns to upstream.
func (m *Msg) WithServerRespHead(h interface{}) {
	m.serverRespHead = h
}

// ClientReqHead returns the request package head of client,
func (m *Msg) ClientReqHead() interface{} {
	return m.clientReqHead
}

// WithClientReqHead sets the request package head of client.
func (m *Msg) WithClientReqHead(h interface{}) {
	m.clientReqHead = h
}

// ClientRespHead returns the response package head of client.
func (m *Msg) ClientRespHead() interface{} {
	return m.clientRespHead
}

// WithClientRespHead sets the response package head of client.
func (m *Msg) WithClientRespHead(h interface{}) {
	m.clientRespHead = h
}

// WithRequestID sets request id.
func (m *Msg) WithRequestID(id uint64) {
	m.requestID = id
}

// RequestID returns request id.
func (m *Msg) RequestID() uint64 {
	return m.requestID
}

// WithSpanID sets span id.
func (m *Msg) WithSpanID(id uint64) {
	m.spanID = id
}

// SpanID returns span id.
func (m *Msg) SpanID() uint64 {
	return m.spanID
}

// WithTraceID sets trace id.
func (m *Msg) WithTraceID(id string) {
	m.traceID = id
}

// TraceID returns trace id.
func (m *Msg) TraceID() string {
	return m.traceID
}

// WithLogger sets log into context message.
//func (m *Msg) WithLogger(l logger.Logger) {
//	m.logger = l
//}

// Logger returns log from context message.
//func (m *Msg) Logger() logger.Logger {
//	return m.logger
//}

// LogSeq returns logger sequence
func (m *Msg) LogSeq() int {
	m.logSeq++

	if m.logSeq > 999999999 {
		m.logSeq = 1
	}

	return m.logSeq
}

// CopyMsg copy src Msg to dst.
func CopyMsg(dst, src *Msg) {
	if dst == nil || src == nil {
		return
	}
	dst.WithFrameCodec(src.FrameCodec())
	dst.WithRequestTimeout(src.RequestTimeout())
	dst.WithSerializationType(src.SerializationType())
	dst.WithCallerServiceName(src.CallerServiceName())
	dst.WithCallerMethod(src.CallerMethod())
	dst.WithCalleeServiceName(src.CalleeServiceName())
	dst.WithCalleeMethod(src.CalleeMethod())
	dst.WithCallRPCName(src.CallRPCName())
	dst.WithServerRespError(src.ServerRespError())
	dst.WithClientRespError(src.ClientRespError())
	dst.WithServerReqHead(src.ServerReqHead())
	dst.WithServerRespHead(src.ServerRespHead())
	dst.WithClientReqHead(src.ClientReqHead())
	dst.WithClientRespHead(src.ClientRespHead())
	dst.WithLocalAddr(src.LocalAddr())
	dst.WithRemoteAddr(src.RemoteAddr())
	//dst.WithLogger(src.Logger())
	dst.WithNamespace(src.Namespace())
	dst.WithEnv(src.Env())
	dst.WithRequestID(src.RequestID())
	dst.WithSpanID(src.SpanID())
	dst.WithTraceID(src.TraceID())
	dst.logSeq = src.logSeq
}

type detachedContext struct{ parent context.Context }

func detach(ctx context.Context) context.Context { return detachedContext{ctx} }

// Deadline implements context.Deadline
func (v detachedContext) Deadline() (time.Time, bool) { return time.Time{}, false }

// Done implements context.Done
func (v detachedContext) Done() <-chan struct{} { return nil }

// Err implements context.Err
func (v detachedContext) Err() error { return nil }

// Value implements context.Value
func (v detachedContext) Value(key interface{}) interface{} { return v.parent.Value(key) }
