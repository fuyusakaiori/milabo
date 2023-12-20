package nim

import (
	"net"
	"time"
)

// 协议服务端
type Server interface {
	// Start: 启动服务端
	Start() error
	// Shutdown: 关闭服务端
	Shutdown()
	// SetChannelMap: 设置连接管理器
	SetChannelMap(channelMap ChannelMap)
	// SetReadWait 设置连接读超时
	SetReadWait(timeout time.Duration)
	// SetAcceptor: 设置 Acceptor
	SetAcceptor(acceptor Acceptor)
	// SetStateListener: 设置连接状态监听器
	SetStateListener(listener StateListener)
}

// Acceptor 连接接收器
type Acceptor interface {
	// Accept: 建立连接返回 ChannelID
	Accept(conn Conn, timeout time.Duration) (string, error)
}

// StateListener 连接状态监听器
type StateListener interface {
	// Disconnect 断开连接
	Disconnect(channelId string) error
}

// MessageListener 消息监听器
type MessageListener interface {
	// Receive: 接收消息
	Receive(agent Agent, message []byte)
}

type Channel interface {
	Conn
	Agent
	ReceiveMessage(listener MessageListener) error
	SetWriteWait(timeout time.Duration)
	SetReadWait(timeout time.Duration)
	Close() error
}

// Agent 消息发送方
type Agent interface {
	// GetChannelID 获取 ChannelID
	GetChannelID() string
	// PushMessage 发送消息
	SendMessage(message []byte) error
}

// Client 客户端
type Client interface {
	// GetClientID 获取客户端 ID
	GetClientID() string
	// GetClientName 获取客户端名称
	GetClientName() string
	// SetDialer: 设置建立连接相关信息
	SetDialer(dialer Dialer)
	// Connect 建立连接 ip:port
	Connect(address string)
	// SendMessage: 发送消息
	SendMessage(message []byte)
	// ReadMessage: 读取消息
	ReadMessage() (Frame, error)
	// Close: 关闭客户端
	Close()
}

// Dialer 连接器
type Dialer interface {
	// DialAndHandshake: 握手建立连接
	DialAndHandshake(ctx DialerContext) (net.Conn, error)
}

// DialerContext 连接器上下文
type DialerContext struct {
	ID      string
	Name    string
	Address string
	Timeout time.Duration
}


// Conn 封装后的网络连接
type Conn interface {
	net.Conn
	ReadFrame() (Frame, error)
	WriteFrame(code OpCode, bytes []byte) error
	Flush() error
}

// OpCode 二进制表示操作类型
type OpCode byte

const (
	OpContinuation OpCode = 0x0
	OpText         OpCode = 0x1
	OpBinary       OpCode = 0x2
	OpClose        OpCode = 0x8
	OpPing         OpCode = 0x9
	OpPong         OpCode = 0xa
)

type Frame interface {
	SetOpCode(code OpCode)
	GetOpCode() OpCode
	SetPayLoad(bytes []byte)
	GetPayLoad() []byte
}
