package main

import (
	"neptune-golang/neptune-tcp/demo/ZinxV0.3/server/router"
	"neptune-golang/neptune-tcp/zinx/znet"
)

func main() {
	// 1. 创建服务器对象
	server := znet.NewServer()
	// 2. 调用服务器方法
	server.AddRouter(1, &router.PingRouter{})
	server.Serve()
}
