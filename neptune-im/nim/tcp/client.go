package tcp

import (
	"errors"
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
	writeWait   time.Duration = 10 * time.Second
	readWait    time.Duration = 10 * time.Second
	hearBeat    time.Duration = 10 * time.Second
	connectWait time.Duration = 10 * time.Second
)

type ClientOptions struct {
	connectWait time.Duration
	readWait    time.Duration
	writeWait   time.Duration
	heartBeat   time.Duration
}

type Client struct {
	sync.Mutex
	nim.Dialer
	clientId   string
	clientName string
	state      int32
	conn       nim.Conn
	options    ClientOptions
	once       sync.Once
}

func NewClient(clientId, clientName string, options ClientOptions) nim.Client {
	if options.connectWait == 0 {
		options.connectWait = connectWait
	}
	if options.writeWait == 0 {
		options.writeWait = writeWait
	}
	if options.readWait == 0 {
		options.readWait = readWait
	}
	if options.heartBeat == 0 {
		options.heartBeat = hearBeat
	}
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
	// 1. 解析地址是否正确
	if _, err := url.Parse(address); err != nil {

	}
	// 2. 更新客户端状态
	if !atomic.CompareAndSwapInt32(&client.state, 0, 1) {

	}
	// 3. 握手建立连接
	rawconn, err := client.DialAndHandshake(nim.DialerContext{
		ID:      client.clientId,
		Name:    client.clientName,
		Address: address,
		Timeout: client.options.connectWait,
	})
	if err != nil {

	}
	// 4. 封装连接
	conn := NewConn(rawconn)
	// 5. 心跳检测
	if client.options.heartBeat > 0 {
		go func() {
			if err := client.heartBeat(conn); err != nil {

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
		"client_id":   client.clientId,
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

func (client *Client) SendMessage(message []byte) error {
	// 1. 判断连接状态是否正常
	if atomic.LoadInt32(&client.state) == 0 {
		return errors.New("client state is closed")
	}
	// 2. 上锁
	client.Lock()
	defer client.Unlock()
	if client.conn == nil {
		return errors.New("client connect interrupted")
	}
	// 3. 设置写超时时间
	_ = client.conn.SetWriteDeadline(time.Now().Add(client.options.writeWait))
	// 4. 发送消息
	return client.conn.WriteFrame(nim.OpBinary, message)
}

func (client *Client) ReadMessage() (nim.Frame, error) {
	if client.state == 0 || client.conn == nil {
		return nil, errors.New("client connect interrupted")
	}
	if client.options.heartBeat > 0 {
		_ = client.conn.SetReadDeadline(time.Now().Add(client.options.readWait))
	}
	frame, err := client.conn.ReadFrame()
	if err != nil {
		return nil, err
	}
	if frame.GetOpCode() == nim.OpClose {
		return nil, errors.New("server close connect")
	}
	return frame, nil
}

func (client *Client) Close() {
	client.once.Do(func() {
		if client.conn == nil {
			return
		}
		_ = client.conn.WriteFrame(nim.OpClose, nil)
		_ = client.conn.Close()
		atomic.CompareAndSwapInt32(&client.state, 1, 0)
	})
}
