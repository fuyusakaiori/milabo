package router

import (
	"fmt"
	"neptune-go/src/zinx/ziface"
	"neptune-go/src/zinx/znet"
)

type PingRouter struct {
	// TODO go 继承?
	znet.BaseHandler
}

func (router *PingRouter) PreHandle(request ziface.IRequest) {
	fmt.Println("ping router before handle... ")
	if _, err := request.GetConn().GetTCPConn().Write([]byte("before ping\t")); err != nil {
		fmt.Println("ping router before handle err", err)
	}
}

func (router *PingRouter) Handle(request ziface.IRequest) {
	fmt.Println("ping router handle... ")
	if _, err := request.GetConn().GetTCPConn().Write([]byte("ping ping ping\t")); err != nil {
		fmt.Println("ping router handle err", err)
	}
}

func (router *PingRouter) PostHandle(request ziface.IRequest) {
	fmt.Println("ping router after handle... ")
	if _, err := request.GetConn().GetTCPConn().Write([]byte("after ping\t")); err != nil {
		fmt.Println("ping router after handle err", err)
	}
}
