package pkt

import (
	"io"
	"neptune-im/nim/util/endian"
)

const (
	CodePing = uint16(1)
	CodePong = uint16(2)
)

// BasicPacket 基础协议: 处理轻量消息
// 消息大小: 魔术字段(4B) + 消息类型(2B) + 消息长度(2B)
// 设计原因: web 端不开放 websocket 心跳协议, 只能在业务层支持心跳协议?
type BasicPacket struct {
	// Code 消息类型
	Code uint16
	// 消息长度
	Length uint16
	// 消息内容
	Body []byte
}

func (packet *BasicPacket) Decode(reader io.Reader) error {
	var err error
	// 1. 读取消息类型
	if packet.Code, err = endian.ReadUint16(reader); err != nil {
		return err
	}
	// 2. 读取消息长度
	if packet.Length, err = endian.ReadUint16(reader); err != nil {
		return err
	}
	// 3. 读取消息内容
	if packet.Body, err = endian.ReadFixedBytes(int(packet.Length), reader); err != nil {
		return err
	}
	return nil
}

func (packet *BasicPacket) Encode(writer io.Writer) error {
	// 1. 写入消息类型
	if err := endian.WriteUint16(writer, packet.Code); err != nil {
		return err
	}
	// 2. 写入消息长度
	if err := endian.WriteUint16(writer, packet.Length); err != nil {
		return err
	}
	// 3. 写入消息内容
	if err := endian.WriteBytes(writer, packet.Body); err != nil {
		return err
	}
	return nil
}

