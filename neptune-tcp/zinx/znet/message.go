package znet

type Message struct {
	MessageID     uint32
	MessageLength uint32
	MessageData   []byte
}

func NewMessage(id uint32, data []byte) *Message {
	return &Message{
		MessageID:     id,
		MessageLength: uint32(len(data)),
		MessageData:   data,
	}
}

func (message *Message) GetMessageID() uint32 {
	return message.MessageID
}

func (message *Message) GetMessageLength() uint32 {
	return message.MessageLength
}

func (message *Message) GetMessageData() []byte {
	return message.MessageData
}

func (message *Message) SetMessageID(messageID uint32) {
	message.MessageID = messageID
}

func (message *Message) SetMessageLength(messageLength uint32) {
	message.MessageLength = messageLength
}

func (message *Message) SetMessageData(messageData []byte) {
	message.MessageData = messageData
}
