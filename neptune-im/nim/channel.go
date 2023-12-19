package nim

import (
	"time"
)

type ChannelImpl struct {
	Conn
}

func NewChannel(channelId string, conn Conn) Channel {
	return nil
}

func (channel *ChannelImpl) GetChannelID() string {
	//TODO implement me
	panic("implement me")
}

func (channel *ChannelImpl) PushMessage(message []byte) error {
	//TODO implement me
	panic("implement me")
}

func (channel *ChannelImpl) ReadLoop(listener MessageListener) error {
	//TODO implement me
	panic("implement me")
}

func (channel *ChannelImpl) SetWriteWait(timeout time.Duration) {
	//TODO implement me
	panic("implement me")
}

func (channel *ChannelImpl) SetReadWait(timeout time.Duration) {
	//TODO implement me
	panic("implement me")
}



