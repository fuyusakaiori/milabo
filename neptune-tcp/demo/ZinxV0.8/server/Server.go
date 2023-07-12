package main

import (
	"neptune-go/src/nep/ZinxV0.8/server/router"
	"neptune-go/src/zinx/znet"
)

func main() {
	// 1. 创建服务器对象
	server := znet.NewServer()
	// 2. 调用服务器方法
	server.AddRouter(1, &router.PingHandler{})
	server.AddRouter(2, &router.HelloHandler{})
	server.Serve()
}
