package router

import (
	"fmt"
	"neptune-go/src/zinx/ziface"
	"neptune-go/src/zinx/znet"
)

type HelloHandler struct {
	znet.BaseHandler
}

func (router *HelloHandler) Handle(request ziface.IRequest) {
	fmt.Println("[zinx] hello router handle... ")
	// 1. 读取消息
	id := request.GetMessage().GetMessageID()
	length := request.GetMessage().GetMessageLength()
	data := request.GetMessage().GetMessageData()
	fmt.Println("[zinx] hello handler id=", id, "\tlength=", length, "\tdata=", string(data))
	// 2. 写回消息
	if err := request.GetConn().SendMessage(id, []byte("hello handler: "+string(data))); err != nil {
		fmt.Println("[zinx] hello handler write err", err)
	}
}
