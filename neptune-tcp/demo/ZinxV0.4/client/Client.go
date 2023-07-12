package main

import (
	"fmt"
	"net"
	"time"
)

func main() {
	fmt.Println("client start...")

	time.Sleep(time.Second)
	// 1. 连接服务器
	connection, err := net.Dial("tcp", "127.0.0.1:4396")
	if err != nil {
		fmt.Println("client start err")
		return
	}
	// 2. 向服务器发送数据
	var count uint32 = 0
	for {
		count++
		if _, err := connection.Write([]byte("Hello ZinxV0.3")); err != nil {
			fmt.Println("write connection err", err)
			return
		}
		buf := make([]byte, 2048)
		length, err := connection.Read(buf)
		if err != nil {
			fmt.Println("read connection err", err)
			return
		}
		fmt.Printf("%d server call back %s, length=%d\n", count, buf, length)

		time.Sleep(time.Second)
	}
}
