package ziface

type IMessage interface {
	GetMessageID() uint32
	GetMessageLength() uint32
	GetMessageData() []byte

	SetMessageID(messageID uint32)
	SetMessageLength(messageLength uint32)
	SetMessageData(messageData []byte)
}
