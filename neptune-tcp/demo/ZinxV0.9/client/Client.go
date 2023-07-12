package main

import (
	"fmt"
	"io"
	"math/rand"
	"neptune-go/src/zinx/znet"
	"net"
	"time"
)

func main() {
	fmt.Println("client start...")

	time.Sleep(time.Second)
	// 1. 连接服务器
	conn, err := net.Dial("tcp", "127.0.0.1:2333")
	if err != nil {
		fmt.Println("client start err")
		return
	}
	// 2. 向服务器发送数据
	var count uint32 = 0
	for {
		count++
		// 2.1 获取编解码器
		codec := znet.NewCodec()
		// 2.2 发送数据
		message := znet.NewMessage(uint32(rand.Intn(2)+1), []byte("ZinxV0.9"))
		buf, err := codec.Encode(message)
		if err != nil {
			fmt.Println("[zinx] write encode buf err ", err)
			return
		}
		if _, err := conn.Write(buf); err != nil {
			fmt.Println("[zinx] write buf err", err)
			return
		}
		// 2.3 接收数据
		headBuf := make([]byte, codec.GetHeadLength())
		if _, err := io.ReadFull(conn, headBuf); err != nil {
			fmt.Println("[zinx] read head buf err", err)
			return
		}
		response, err := codec.Decode(headBuf)
		if err != nil {
			fmt.Println("[zinx] read decode head buf err", err)
			return
		}
		dataBuf := make([]byte, response.GetMessageLength())
		if _, err := io.ReadFull(conn, dataBuf); err != nil {
			fmt.Println("[zinx] read body buf err", err)
			return
		}
		response.SetMessageData(dataBuf)

		fmt.Printf("%d server call back %s, length=%d\n", count, buf, response.GetMessageLength())

		time.Sleep(time.Second)
	}
}
