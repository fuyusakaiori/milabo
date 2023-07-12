package ziface

type ICodec interface {
	GetHeadLength() uint32

	Encode(message IMessage) (data []byte, err error)

	Decode(data []byte) (message IMessage, err error)
}
