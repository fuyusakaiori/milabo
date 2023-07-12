package main

import (
	"neptune-golang/neptune-tcp/demo/ZinxV0.7/server/router"
	"neptune-golang/neptune-tcp/zinx/znet"
)

func main() {
	// 1. 创建服务器对象
	server := znet.NewServer()
	// 2. 调用服务器方法
	server.AddRouter(1, &router.PingHandler{})
	server.AddRouter(2, &router.HelloHandler{})
	server.Serve()
}
