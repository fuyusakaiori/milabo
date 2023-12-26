package endian

import (
	"encoding/binary"
	"io"
)

// 读取字节流中的前 8 个比特 (1个字节)
func ReadUint8(reader io.Reader) (uint8, error) {
	// 1. 初始化字节数组
	buf := make([]byte, 1)
	// 2. 读取字节流
	if _, err := io.ReadFull(reader, buf); err != nil {
		return 0, err
	}
	return uint8(buf[0]), nil
}

func ReadUint32(reader io.Reader) (uint32, error) {
	// 1. 初始化字节数组
	bytes := make([]byte, 4)
	// 2. 读取字节流
	if _, err := io.ReadFull(reader, bytes); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(bytes), nil
}

// 读取字节流中的任意长度的字节: 规定基于 tcp 协议的自定义协议格式
func ReadBytes(reader io.Reader) ([]byte, error) {
	// 1. 读取消息长度
	length, err := ReadUint32(reader)
	if err != nil {
		return nil, err
	}
	// 2. 初始化字节数组
	buf := make([]byte, length)
	// 3. 读取消息内容
	if _, err = io.ReadFull(reader, buf); err != nil {
		return nil, err
	}
	return buf, nil
}

func WriteUint8(writer io.Writer, value uint8) error {
	// 1. 初始化字节数组
	buf := []byte{byte(value)}
	// 2. 写入连接中
	if _, err := writer.Write(buf); err != nil {
		return err
	}
	return nil
}

func WriteUint32(writer io.Writer, value uint32) error {
	// 1. 初始化字节数组
	buf := make([]byte, 4)
	// 2. 数据内容填入数组
	binary.LittleEndian.PutUint32(buf, value)
	// 2. 写入连接中
	if _, err := writer.Write(buf); err != nil {
		return err
	}
	return nil
}


func WriteBytes(writer io.Writer, bytes []byte) error {
	// 1. 计算字节数组长度
	length := len(bytes)
	// 2. 写入数据长度
	if err := WriteUint32(writer, uint32(length)); err != nil {
		return err
	}
	// 3. 写入数据
	if _ , err := writer.Write(bytes); err != nil {
		return err
	}
	return nil
}
