package main

import (
	"fmt"
	"neptune-golang/neptune-tcp/demo/ZinxV0.9/server/router"
	"neptune-golang/neptune-tcp/zinx/ziface"
	"neptune-golang/neptune-tcp/zinx/znet"
)

func NeptuneOnConnStart(connection ziface.IConnection) {
	if err := connection.SendMessage(1, []byte("[zinx] on conn start hook")); err != nil {
		fmt.Println("[zinx] on conn start hook fail")
	}
	fmt.Println("[zinx] on conn start hook success")
}

func NeptuneOnConnStop(connection ziface.IConnection) {
	if err := connection.SendMessage(2, []byte("[zinx] on conn stop hook")); err != nil {
		fmt.Println("[zinx] on conn stop hook fail")
	}
	fmt.Println("[zinx] on conn stop hook success")
}

func main() {
	// 1. 创建服务器对象
	server := znet.NewServer()
	// 2. 调用服务器方法
	server.AddRouter(1, &router.PingHandler{})
	server.AddRouter(2, &router.HelloHandler{})
	server.SetOnConnStart(NeptuneOnConnStart)
	server.SetOnConnStop(NeptuneOnConnStop)
	server.Serve()
}
