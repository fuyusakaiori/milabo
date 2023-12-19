package websocket

import (
	"github.com/gobwas/ws"
	"neptune-im/nim"
	"net"
)

// Frame 封装帧
type Frame struct {
	ws.Frame
}

func (frame *Frame) SetOpCode(code nim.OpCode) {
	frame.Header.OpCode = ws.OpCode(code)
}

func (frame *Frame) GetOpCode() nim.OpCode {
	return nim.OpCode(frame.Header.OpCode)
}

func (frame *Frame) SetPayLoad(payload []byte) {
	frame.Payload = payload
}

func (frame *Frame) GetPayLoad() []byte {
	// 1. 加密
	if frame.Header.Masked {
		ws.Cipher(frame.Payload, frame.Header.Mask, 0)
	}
	// 2. 重新设置是否需要加密
	frame.Header.Masked = false
	// 3. 返回加密数据
	return frame.Payload
}

// WsConn Websocket 连接: 统一处理 Websocket 协议和 Tcp 协议
type WsConn struct {
	net.Conn
}

func NewConn(conn net.Conn) *WsConn {
	return &WsConn{
		Conn: conn,
	}
}

func (conn *WsConn) ReadFrame() (nim.Frame, error) {
	// 1. websocket 读取帧
	frame, err := ws.ReadFrame(conn.Conn)
	// 2. 是否读取成功
	if err != nil {
		return nil, err
	}
	// 3. 封装并返回读取的数据
	return &Frame{Frame: frame}, nil
}

func (conn *WsConn) WriteFrame(code nim.OpCode, payload []byte) error {
	// 1. 创建帧数据
	frame := ws.NewFrame(ws.OpCode(code), true, payload)
	// 2. 写入帧数据
	return ws.WriteFrame(conn.Conn, frame)
}

func (conn *WsConn) Flush() error {
	return nil
}



