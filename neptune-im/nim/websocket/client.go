package websocket

import (
	"errors"
	"fmt"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/sirupsen/logrus"
	"neptune-im/nim"
	"net"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
)

const (
	writeWaitTime   time.Duration = 10 * time.Second
	readWaitTime    time.Duration = 10 * time.Second
	connectWaitTime time.Duration = 10 * time.Second
)

type ClientOptions struct {
	heartBeat time.Duration
	readWait  time.Duration
	writeWait time.Duration
}

type Client struct {
	sync.Mutex
	// Client 无法直接使用封装的 Conn
	net.Conn
	nim.Dialer
	clientId   string
	clientName string
	state      int32
	options    ClientOptions
	once       sync.Once
}

func NewClient(clientId, clientName string, options ClientOptions) nim.Client {
	// 1. 设置配置
	if options.writeWait == 0 {
		options.writeWait = writeWaitTime
	}
	if options.readWait == 0 {
		options.readWait = readWaitTime
	}
	// 2. 初始话客户端
	return &Client{
		clientId:   clientId,
		clientName: clientName,
		options:    options,
	}
}

func (client *Client) GetClientID() string {
	return client.clientId
}

func (client *Client) GetClientName() string {
	return client.clientName
}

func (client *Client) SetDialer(dialer nim.Dialer) {
	client.Dialer = dialer
}

func (client *Client) Connect(address string) error {
	// 1. 解析地址
	if _, err := url.Parse(address); err != nil {
		return err
	}
	// 2. 原子性更新客户端连接状态 (断线重连可能造成并发?)
	if !atomic.CompareAndSwapInt32(&client.state, 0, 1) {
		return errors.New("client has connected")
	}
	// 3. 客户端建立连接
	rawconn, err := client.Dialer.DialAndHandshake(nim.DialerContext{
		ID:      client.clientId,
		Name:    client.clientName,
		Address: address,
		Timeout: connectWaitTime,
	})
	// 4. 连接建立失败, 重新更新状态
	if err != nil {
		if !atomic.CompareAndSwapInt32(&client.state, 1, 0) {
			return errors.New( fmt.Sprintf("client connected fail and update state fail, %v", err))
		}
		return errors.New(fmt.Sprintf("client connected fail, %v", err))
	}
	// 5. 设置连接
	client.Conn = rawconn
	// 6. 开启协程发送心跳
	if client.options.heartBeat > 0 {
		go func() {
			if err := client.heartBeat(rawconn); err != nil {

			}
		}()
	}
	return nil
}

func (client *Client) heartBeat(conn net.Conn) error {
	// 1. 初始化定时器
	ticker := time.NewTicker(client.options.heartBeat)
	// 2. 定义定时器的事件
	for range ticker.C {
		// 2.1 发送 ping 包
		if err := client.ping(conn); err != nil {

		}
	}
	return nil
}

// ping 为什么发送 ping 包也需要上锁
func (client *Client) ping(conn net.Conn) error {
	logger := logrus.WithFields(logrus.Fields{
		"client_id": client.clientId,
		"client_name": client.clientName,
	})
	client.Lock()
	defer client.Unlock()
	// 1. 重置写超时时间
	_ = conn.SetWriteDeadline(time.Now().Add(client.options.writeWait))
	// 2. 发送 ping 消息
	logger.Tracef("send ping message to server")

	return wsutil.WriteClientMessage(conn, ws.OpPing, nil)
}

// SendMessage 为什么发消息上锁收消息不上锁？
func (client *Client) SendMessage(message []byte) error {
	client.Lock()
	defer client.Unlock()
	// 1. 判断连接是否为空
	if client.Conn == nil {
		return errors.New("client connect interrupted")
	}
	// 2. 重置写超时时间
	_ = client.Conn.SetWriteDeadline(time.Now().Add(client.options.writeWait))
	// 3. 发送消息
	return wsutil.WriteClientMessage(client.Conn, ws.OpBinary, message)
}

// ReadMessage 方法非线程安全
func (client *Client) ReadMessage() (nim.Frame, error) {
	// 1. 判断连接是否为空
	if client.Conn == nil {
		return nil, errors.New("client connect interrupted")
	}
	// 2. 重置读超时时间 (为什么要重置读超时时间)
	if client.options.heartBeat > 0 {
		_ = client.Conn.SetReadDeadline(time.Now().Add(client.options.readWait))
	}
	// 3. 从连接中读取消息
	frame, err := ws.ReadFrame(client.Conn)
	if err != nil {
		return nil, err
	}
	// 4. 判断是否为关闭连接的消息
	if frame.Header.OpCode == ws.OpClose {
		return nil, errors.New("server close connect")
	}
	return &Frame{
		Frame: frame,
	}, nil
}

func (client *Client) Close() {
	client.once.Do(func() {
		// 1. 判断连接是否为空
		if client.Conn == nil {
			return
		}
		// 2. 发送关闭连接的消息
		_ = wsutil.WriteClientMessage(client.Conn, ws.OpClose, nil)
		// 3. 关闭连接
		_ = client.Conn.Close()
		// 4. 更新客户端状态
		if !atomic.CompareAndSwapInt32(&client.state, 1, 0) {

		}
	})
}
