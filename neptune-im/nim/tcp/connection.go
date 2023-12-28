package tcp

import (
	"neptune-im/nim"
	"neptune-im/nim/util/endian"
	"net"
)

type Frame struct {
	OpCode  nim.OpCode
	PayLoad []byte
}

func (frame *Frame) SetOpCode(code nim.OpCode) {
	frame.OpCode = code
}

func (frame *Frame) GetOpCode() nim.OpCode {
	return frame.OpCode
}

func (frame *Frame) SetPayLoad(payload []byte) {
	frame.PayLoad = payload
}

func (frame *Frame) GetPayLoad() []byte {
	return frame.PayLoad
}

type TcpConn struct {
	net.Conn
}

func NewConn(conn net.Conn) nim.Conn {
	return &TcpConn{
		Conn: conn,
	}
}

func (conn *TcpConn) ReadFrame() (nim.Frame, error) {
	// 1. 读取事件类型
	opcode, err := endian.ReadUint8(conn)
	if err != nil {
		return nil, err
	}
	// 2. 读取消息内容
	message, err := endian.ReadBytes(conn)
	if err != nil {
		return nil, err
	}
	// 3. 封装消息体
	return &Frame{
		OpCode:  nim.OpCode(opcode),
		PayLoad: message,
	}, nil
}

func (conn *TcpConn) WriteFrame(code nim.OpCode, bytes []byte) error {
	// 1. 写入事件类型
	if err := endian.WriteUint8(conn, uint8(code)); err != nil {
		return err
	}
	// 2. 写入数据内容
	if err := endian.WriteBytes(conn, bytes); err != nil {
		return err
	}
	return nil
}

func (conn *TcpConn) Flush() error {
	return nil
}
