package znet

import "neptune-golang/neptune-tcp/zinx/ziface"

// Request 请求
type Request struct {
	Message ziface.IMessage
	Conn    ziface.IConnection
}

func (request *Request) GetMessage() ziface.IMessage {
	return request.Message
}

func (request *Request) GetConn() ziface.IConnection {
	return request.Conn
}
