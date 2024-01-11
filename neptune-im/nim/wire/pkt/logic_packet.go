package pkt

import (
	"google.golang.org/protobuf/proto"
	"io"
	"neptune-im/nim/util/endian"
	"neptune-im/nim/wire"
)

type LogicPacket struct {
	Header
	Body []byte
}

type HeaderOption func(header *Header)

func NewLogicPacket(command string, options ...HeaderOption) *LogicPacket {
	// 1. 初始化消息
	packet := &LogicPacket{}
	// 2. 设置协议内容
	packet.Command = command
	for _, option := range options {
		option(&packet.Header)
	}
	// 3. 设置序列号
	if packet.Sequence == 0 {
		packet.Sequence = wire.Seq.Next()
	}
	return packet
}

func WithChannelID(channelId string) HeaderOption {
	return func(header *Header) {
		header.ChannelId = channelId
	}
}

func WithSequence(sequence uint32) HeaderOption {
	return func(header *Header) {
		header.Sequence = sequence
	}
}


func WithStatus(status MessageStatus) HeaderOption {
	return func(header *Header) {
		header.Status = status
	}
}

func WithDestination(destination string) HeaderOption {
	return func(header *Header) {
		header.Destination = destination
	}
}

func (packet *LogicPacket) Decode(reader io.Reader) error {
	// 1. 读取消息头: 这怎么读取出来的?
	header, err := endian.ReadBytes(reader)
	if err != nil {
		return err
	}
	// 2. 解析消息头
	_ = proto.Unmarshal(header, &packet.Header)
	// 3. 读取消息内容
	packet.Body, err = endian.ReadBytes(reader)
	if err != nil {
		return err
	}
	return nil
}

func (packet *LogicPacket) Encode(writer io.Writer) error {
	// 1. 消息头序列化成字节数据
	header, err := proto.Marshal(&packet.Header)
	if err != nil {
		return err
	}
	// 2. 写入消息头
	if err = endian.WriteBytes(writer, header); err != nil {
		return err
	}
	// 3. 写入消息内容
	if err = endian.WriteBytes(writer, packet.Body); err != nil {
		return err
	}
	return nil
}

