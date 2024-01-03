package pkt

import (
	"io"
)

type BasicPacket struct {

}

func (packet *BasicPacket) Decode(reader io.Reader) {
	//TODO implement me
	panic("implement me")
}

func (packet *BasicPacket) Encode(writer io.Writer) {
	//TODO implement me
	panic("implement me")
}

