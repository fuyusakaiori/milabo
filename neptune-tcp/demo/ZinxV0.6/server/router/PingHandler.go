package router

import (
	"fmt"
	"neptune-go/src/zinx/ziface"
	"neptune-go/src/zinx/znet"
)

type PingHandler struct {
	// TODO go 继承?
	znet.BaseHandler
}

func (router *PingHandler) Handle(request ziface.IRequest) {
	fmt.Println("[zinx] ping router handle... ")
	// 1. 读取消息
	id := request.GetMessage().GetMessageID()
	length := request.GetMessage().GetMessageLength()
	data := request.GetMessage().GetMessageData()
	fmt.Println("[zinx] ping handler id=", id, "\tlength=", length, "\tdata=", string(data))
	// 2. 写回消息
	if err := request.GetConn().SendMessage(id, []byte("ping handler: "+string(data))); err != nil {
		fmt.Println("[zinx] ping handler write err", err)
	}
}
