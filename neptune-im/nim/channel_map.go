package nim

import (
	"github.com/sirupsen/logrus"
	"sync"
)

// ChannelMap 连接管理器
type ChannelMap interface {
	Get(channelId string) (Channel, bool)
	Put(channel Channel)
	Remove(channelId string)
	List() []Channel
}

type ChannelMapImpl struct {
	channels *sync.Map
}

func NewChannelMap() ChannelMap {
	return &ChannelMapImpl{
		channels: new(sync.Map),
	}
}

func (channelMap *ChannelMapImpl) Get(channelId string) (Channel, bool) {
	// 1. 判断 channel id 是否为空
	if channelId == "" {
		logrus.WithFields(logrus.Fields{
			"struct": "ChannelMapImpl",
			"func":   "Get",
		}).Errorf("channel id is empty")
		return nil, false
	}
	// 2. 获取 channel
	if channel, ok := channelMap.channels.Load(channelId); ok {
		return channel.(Channel), true
	}
	return nil, false
}

func (channelMap *ChannelMapImpl) Put(channel Channel) {
	// 1. 判断 channel 是否为空
	if channel == nil || channel.GetChannelID() == "" {
		logrus.WithFields(logrus.Fields{
			"struct": "ChannelMapImpl",
			"func":   "Put",
		}).Errorf("channel is nil or channel id is empty")
		return
	}
	// 2. 放入 channel
	channelMap.channels.Store(channel.GetChannelID(), channel)
}

func (channelMap *ChannelMapImpl) Remove(channelId string) {
	// 1. 判断 channel id 是否为空
	if channelId == "" {
		logrus.WithFields(logrus.Fields{
			"struct": "ChannelMapImpl",
			"func":   "Remove",
		}).Errorf("channel id is empty")
		return
	}
	// 2. 移除 channel
	channelMap.channels.Delete(channelId)
}

func (channelMap *ChannelMapImpl) List() []Channel {
	// 1. 初始化切片
	channels := make([]Channel, 0)
	// 2. 遍历哈希表
	channelMap.channels.Range(func(channelId, channel any) bool {
		channels = append(channels, channel.(Channel))
		return true
	})
	return channels
}
