package znet

import (
	"fmt"
	"io"
	"neptune-golang/neptune-tcp/zinx/utils"
	"neptune-golang/neptune-tcp/zinx/ziface"
	"net"
	"sync"
)

type Connection struct {
	// 连接 ID
	ConnID uint32
	// 连接
	Conn *net.TCPConn
	// 连接状态
	isClosed bool
	// 处理器
	Router ziface.IRouter
	// TODO 负责交换退出消息的管道 (goroutine)
	ExitChan chan bool
	// TODO 负责交换客户端消息的管道
	MessageChan chan []byte
	// 连接所属服务器
	Server ziface.IServer
	// 附加参数
	properties   map[string]interface{}
	propertyLock sync.RWMutex
}

func (conn *Connection) StartConn() {
	fmt.Println("Conn Start... ConnID", conn.ConnID)
	// 1. 执行读取函数
	go conn.ReadConn()
	// 2. 执行写入函数
	go conn.WriteConn()
	// 3. 执行回调
	conn.Server.GetOnConnStart(conn)
}

func (conn *Connection) StopConn() {
	fmt.Println("Conn Stop.. ConnID", conn.ConnID)
	// 1. 检查连接是否已经关闭
	if conn.isClosed {
		fmt.Println("connection already close, ConnID", conn.ConnID)
		return
	}
	// 2. 执行回调
	conn.Server.GetOnConnStop(conn)
	// 3. 如果没有关闭, 那么关闭连接
	conn.isClosed = true
	if err := conn.Conn.Close(); err != nil {
		fmt.Println("Conn Stop err, ConnID", err, conn.ConnID)
	}
	// 4. 关闭之前发送关闭消息
	conn.ExitChan <- true
	// TODO 4. 释放管道资源
	close(conn.ExitChan)
	close(conn.MessageChan)
	// 5. 移除连接
	conn.Server.GetConnManager().CloseConnection(conn)
}

func (conn *Connection) GetTCPConn() *net.TCPConn {
	// 注: 不要写成递归调用
	return conn.Conn
}

func (conn *Connection) GetConnID() uint32 {
	return conn.ConnID
}

func (conn *Connection) RemoteAddr() net.Addr {
	return conn.Conn.RemoteAddr()
}

func (conn *Connection) SendMessage(id uint32, data []byte) error {
	// 1. 获取编解码器
	codec := NewCodec()
	// 2. 封装消息
	message := NewMessage(id, data)
	// 3. 编码
	buf, err := codec.Encode(message)
	if err != nil {
		fmt.Println("[zinx] send encode buf err", err)
		return err
	}
	// 4. 发送数据
	conn.MessageChan <- buf
	return nil
}

func (conn *Connection) ReadConn() {
	fmt.Println("Reader Goroutine is Running... ConnID", conn.ConnID)
	// 1. 函数退出后释放资源
	defer fmt.Println("Reader Goroutine is Exit... ConnID", conn.ConnID)
	defer conn.StopConn()
	for {
		// 2. 获取定长解码器
		codec := NewCodec()
		// 3. 读取消息体的头信
		headBuf := make([]byte, codec.GetHeadLength())
		if _, err := io.ReadFull(conn.Conn, headBuf); err != nil {
			fmt.Println("[zinx] read head buf err", err)
			return
		}
		// 4. 解码器
		message, err := codec.Decode(headBuf)
		// TODO 暂时没有考虑接收到的消息序列号
		if err != nil || message.GetMessageID() < 0 {
			fmt.Println("[zinx] read decode head buf err", err)
			return
		}
		// 5. 读取消息体
		// TODO 暂时没有考虑解决半包问题
		dataBuf := make([]byte, message.GetMessageLength())
		if _, err := io.ReadFull(conn.Conn, dataBuf); err != nil {
			fmt.Println("[zinx] read decode body buf err", err)
			return
		}
		// 6. 向消息体中填充内容
		message.SetMessageData(dataBuf)
		// 7. 封装请求
		req := Request{
			Message: message,
			Conn:    conn,
		}
		// 4. 处理数据
		conn.Router.SendMessageToTaskQueue(&req)
	}
}

func (conn *Connection) WriteConn() {
	fmt.Println("Writer Goroutine is Running... ConnID", conn.ConnID)
	defer fmt.Println("Writer Goroutine is Exit... ConnID", conn.ConnID)
	// 1. 循环阻塞读取读通道交付的数据
	for {
		select {
		// 2. 如果收到通道中的消息, 那么就转发个客户端
		case data := <-conn.MessageChan:
			if _, err := conn.Conn.Write(data); err != nil {
				fmt.Println("[zinx] send buf err")
				return
			}
		// 3. 如果收到关闭消息, 那么就直接退出
		case <-conn.ExitChan:
			return
		}
	}
}

func (conn *Connection) SetConnectionProperty(key string, value interface{}) {
	conn.propertyLock.Lock()
	defer conn.propertyLock.Unlock()
	conn.properties[key] = value
}

func (conn *Connection) GetConnectionProperty(key string) (value interface{}) {
	conn.propertyLock.RLock()
	defer conn.propertyLock.RUnlock()
	result, ok := conn.properties[key]
	if !ok {
		fmt.Println("[zinx] get property doesn't exit")
		return nil
	}
	return result
}

func (conn *Connection) RemoveConnectionProperty(key string) {
	conn.propertyLock.Lock()
	defer conn.propertyLock.Unlock()
	_, ok := conn.properties[key]
	if !ok {
		fmt.Println("[zinx] remove property doesn't exit")
		return
	}
	delete(conn.properties, key)
}

func NewConn(connID uint32, conn *net.TCPConn, router ziface.IRouter, server ziface.IServer) *Connection {
	// 1. 创建连接
	connection := &Connection{
		ConnID:      connID,
		Conn:        conn,
		isClosed:    false,
		Router:      router,
		ExitChan:    make(chan bool, 1),
		MessageChan: make(chan []byte),
		Server:      server,
	}
	// 2. 添加连接
	connection.Server.GetConnManager().AddConnection(connection)
	fmt.Println("now ", connection.Server.GetConnManager().GetConnectionCount(), "limit ", utils.Config.ZinxMaxConn)
	return connection
}
