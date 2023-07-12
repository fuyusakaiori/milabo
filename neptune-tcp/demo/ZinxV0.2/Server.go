package main

import (
	"neptune-go/src/zinx/znet"
)

func main() {
	// 1. 创建服务器对象
	server := znet.NewServer()
	// 2. 调用服务器方法
	server.Serve()
}
