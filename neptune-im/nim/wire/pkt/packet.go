package pkt

import (
	"bytes"
	"errors"
	"io"
	"neptune-im/nim/wire"
	"reflect"
)

// 协议类型


type Packet interface {
	// Decode 解码
	Decode(reader io.Reader) error
	// Encode 编码
	Encode(writer io.Writer) error
}

func Decode(reader io.Reader) (interface{}, error) {
	// 1. 为什么魔术字段不封装在协议里
	magic := wire.Magic{}
	// 2. 读取魔术字段
	if _, err := io.ReadFull(reader, magic[:]); err != nil {
		return nil, err
	}
	// 3. 判断协议类型
	switch magic {
	// 3.1 解析基础协议
	case wire.MagicBasicPacket:
		packet := new(BasicPacket)
		// 3.1.1 解析内容
		if err := packet.Decode(reader); err != nil {
			return nil, err
		}
		// 3.1.2 返回协议内容
		return packet, nil
	// 3.2 解析逻辑协议
	case wire.MagicLogicPacket:
		packet := new(LogicPacket)
		// 3.2.1 解析内容
		if err := packet.Decode(reader); err != nil {
			return nil, err
		}
		// 3.2.2 解析内容
		return packet, nil
	default:
		return nil, errors.New("not support the protocol")
	}
}

func Encode(packet Packet) []byte {
	// 1. 初始化缓冲区
	buf := new(bytes.Buffer)
	// 2. 获取协议类型
	kind := reflect.TypeOf(packet).Elem()
	// 3. 判断协议类型
	if kind.AssignableTo(reflect.TypeOf(BasicPacket{})) {
		buf.Write(wire.MagicBasicPacket[:])
	} else if kind.AssignableTo(reflect.TypeOf(LogicPacket{})) {
		buf.Write(wire.MagicLogicPacket[:])
	} else {
		return nil
	}
	// 4. 魔术字段编码到消息中
	_ = packet.Encode(buf)
	// 5. 返回魔术字段内容
	return buf.Bytes()
}
