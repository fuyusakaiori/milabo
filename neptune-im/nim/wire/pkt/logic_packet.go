package pkt

import (
	"io"
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
	//TODO implement me
	panic("implement me")
}

func (packet *LogicPacket) Encode(writer io.Writer) error {
	//TODO implement me
	panic("implement me")
}

