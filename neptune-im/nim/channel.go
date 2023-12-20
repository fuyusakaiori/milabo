package nim

import (
	"errors"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

const (
	writeChanBuf  int           = 5
	writeWaitTime time.Duration = 10 * time.Second
	readWaitTime  time.Duration = 10 * time.Second
)

type ChannelImpl struct {
	Conn
	sync.Mutex
	channelId string
	writeChan chan []byte
	writeWait time.Duration
	readWait  time.Duration
	once      sync.Once
}

func NewChannel(channelId string, conn Conn) Channel {
	logger := logrus.WithFields(logrus.Fields{
		"module":    "tcp_channel",
		"channelId": channelId,
	})
	// 1. 实例化 channel
	channel := &ChannelImpl{
		Conn:      conn,
		channelId: channelId,
		writeChan: make(chan []byte, writeChanBuf),
		writeWait: writeWaitTime,
		readWait:  readWaitTime,
	}
	// 2. 启动协程写入消息
	go func() {
		if err := channel.sendLoop(); err != nil {
			logger.Errorf("channel start goroutine write message fail, err - %v", err)
		}
	}()
	return channel
}

func (channel *ChannelImpl) GetChannelID() string {
	//TODO implement me
	panic("implement me")
}

// PushMessage 发送消息
func (channel *ChannelImpl) SendMessage(message []byte) error {

	// 2. 异步写入管道 (为什么可以确保线程安全?)
	channel.writeChan <- message
	return nil
}

// sendLoop 接收上层传递的消息并发送
func (channel *ChannelImpl) sendLoop() error {
	for {
		select {
		case message := <-channel.writeChan:
			if err := channel.WriteFrame(OpBinary, message); err != nil {
				// TODO 如果写入出现错误, 那么协程就会终止, 是否合理
				return err
			}
			// 消费完缓冲区中的所有的元素
			for index := 0; index < len(channel.writeChan); index++ {
				message = <-channel.writeChan
				if err := channel.WriteFrame(OpBinary, message); err != nil {
					return err
				}
			}
		}
	}
}

// ReadLoop 读取消息
func (channel *ChannelImpl) ReceiveMessage(listener MessageListener) error {
	// 0. 上锁: 防止重复调用
	channel.Lock()
	defer channel.Unlock()
	logger := logrus.WithFields(logrus.Fields{
		"struct":    "ChannelImpl",
		"func":      "ReceiveMessage",
		"channelId": channel.channelId,
	})
	for {
		// 1. 设置读超时时间
		_ = channel.SetReadDeadline(time.Now().Add(channel.readWait))
		// 2. 读取帧数据
		frame, err := channel.ReadFrame()
		if err != nil {
			// TODO 如果读取消息出现错误, 就直接终止协程, 是否合理
			return err
		}
		if frame.GetOpCode() == OpClose {
			return errors.New("remote side closed channel")
		}
		if frame.GetOpCode() == OpPing {
			_ = channel.WriteFrame(OpPing, nil)
			logger.Infof("receive ping and send pong")
			continue
		}
		if len(frame.GetPayLoad()) == 0 {
			continue
		}
		// 3. 实际处理收到的消息: 不希望调用方直接使用 channel 所以定义的参数是 agent
		go listener.Receive(channel, frame.GetPayLoad())
	}
}

// WriteFrame 重写 WsConn 的方法
func (channel *ChannelImpl) WriteFrame(code OpCode, message []byte) error {
	// 1. 调用 net.Conn 的设置写入超时的方法
	_ = channel.Conn.SetWriteDeadline(time.Now().Add(channel.writeWait))
	// 2. 调用 WsConn 的写入帧的方法
	return channel.Conn.WriteFrame(code, message)
}

func (channel *ChannelImpl) SetWriteWait(timeout time.Duration) {
	//TODO implement me
	panic("implement me")
}

func (channel *ChannelImpl) SetReadWait(timeout time.Duration) {
	//TODO implement me
	panic("implement me")
}
