package pkt

import (
	"io"
)

type LogicPacket struct {

}

func (packet *LogicPacket) Decode(reader io.Reader) {
	//TODO implement me
	panic("implement me")
}

func (packet *LogicPacket) Encode(writer io.Writer) {
	//TODO implement me
	panic("implement me")
}

