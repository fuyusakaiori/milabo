package nim

import (
	"net"
	"time"
)

// 协议服务端
type Server interface {
	// 设置 Acceptor: 客户端回调处理握手逻辑
	SetAcceptor(acceptor Acceptor)

}

// Acceptor 连接接收器
type Acceptor interface {
	// Accept: 建立连接返回 Channel
	Accept(conn Conn, timeout time.Duration)
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
	OpText OpCode = 0x1
	OpBinary OpCode = 0x2
	OpClose OpCode = 0x8
	OpPing OpCode = 0x9
	OpPong OpCode = 0xa
)

type Frame interface {
	SetOpCode(code OpCode)
	GetOpCode() OpCode
	SetPayLoad(bytes []byte)
	GetPayLoad() []byte
}
