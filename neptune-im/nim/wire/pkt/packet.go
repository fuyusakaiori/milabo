package pkt

import "io"

type Packet interface {
	Decode(reader io.Reader)
	Encode(writer io.Writer)
}
