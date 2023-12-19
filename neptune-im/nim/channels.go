package nim

// ChannelMap 连接管理器
type ChannelMap interface {
	Get(channelId string) (Channel, bool)
	Put(channel Channel)
	Remove(channelId string)
	List() []Channel
}