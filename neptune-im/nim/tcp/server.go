package tcp

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"neptune-im/nim"
	"neptune-im/nim/websocket"
	"net"
	"sync"
	"time"
)


type ServerOptions struct {
	connectWait time.Duration
	readWait    time.Duration
	writeWait   time.Duration
}

type Server struct {
	nim.ChannelMap
	nim.Acceptor
	nim.MessageListener
	nim.StateListener

	address string
	once    sync.Once
	options ServerOptions
}

func NewServer() nim.Server {

	return &Server{

	}
}

func (server *Server) Start() error {
	logrus.WithFields(logrus.Fields{
		"module": "tcp_server",
		"address": server.address,
		"server_id": 0,
	})
	// 1. 设置状态监听器
	if server.StateListener == nil {
		return errors.New("state listener is nil")
	}
	// 2. 设置连接器
	if server.Acceptor == nil {
		server.Acceptor = newAcceptor()
	}
	// 3. 设置连接管理器
	if server.ChannelMap == nil {
		server.ChannelMap = nim.NewChannelMap()
	}
	// 4. 开始监听
	listener, err := net.Listen("tcp", server.address)
	if err != nil {
		return errors.New(fmt.Sprintf("tcp listener start failed, %v", err))
	}
	// 5. 处理连接
	for  {
		// 5.1 建立连接
		rawconn, err := listener.Accept()
		if err != nil {

		}
		// 5.2 处理连接
		go func(rawconn net.Conn) {
			// 5.2.1 封装连接
			conn := websocket.NewConn(rawconn)
			// 5.2.2 建立连接
			channelId, err := server.Accept(conn, server.options.connectWait)
			if err != nil {

			}
			// 5.2.3 封装管道
			channel := nim.NewChannel(channelId, conn)
			channel.SetWriteWait(server.options.writeWait)
			channel.SetReadWait(server.options.readWait)
			// 5.2.4 添加管道
			server.Put(channel)
			// 5.2.5 处理消息
			if err := channel.ReceiveMessage(server.MessageListener); err != nil {

			}
			// 5.2.6 如果出现异常退出就移除连接
			server.Remove(channelId)
			// 5.2.7 关闭连接
			_ = server.Disconnect(channelId)
			_ = channel.Close()
		}(rawconn)
	}
}

func (server *Server) SendMessage(channelId string, message []byte) error {
	// 1. 获取连接
	if channel, ok := server.Get(channelId); ok {
		// 2. 通过连接发送消息
		return channel.SendMessage(message)
	}
	return errors.New(fmt.Sprintf("channel is not exist, channel id = %v", channelId))
}

func (server *Server) Shutdown() {
	logrus.WithFields(logrus.Fields{
		"module": "tcp_server",
		"address": server.address,
		"server_id": 0,
	})
	server.once.Do(func() {
		// 1. 获取所有连接
		channels := server.List()
		// 2. 遍历所有连接
		for _, channel := range channels {
			// 3. 移除连接
			server.Remove(channel.GetChannelID())
			// 4. 关闭连接
			_ = channel.Close()
		}
	})
}

func (server *Server) SetChannelMap(channelMap nim.ChannelMap) {
	server.ChannelMap = channelMap
}

func (server *Server) SetReadWait(timeout time.Duration) {
	server.options.readWait = timeout
}

func (server *Server) SetAcceptor(acceptor nim.Acceptor) {
	server.Acceptor = acceptor
}

func (server *Server) SetStateListener(listener nim.StateListener) {
	server.StateListener = listener
}

type defaultAcceptor struct {

}

func newAcceptor() nim.Acceptor {
	return &defaultAcceptor{}
}

func (acceptor *defaultAcceptor) Accept(conn nim.Conn, timeout time.Duration) (string, error) {
	//TODO implement me
	panic("implement me")
}


