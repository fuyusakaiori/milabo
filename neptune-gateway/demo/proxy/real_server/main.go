package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	StatusServerError int = 500
)

type RealServer struct {
	// 服务端地址
	RemoteAddr string
}

func (server *RealServer) Run() {
	// 1. 输出启动日志
	log.Printf("starting http server :%s", server.RemoteAddr)
	// 2. 创建路由
	router := http.NewServeMux()
	// 3. 绑定函数
	router.HandleFunc("/", server.HelloHandler)
	router.HandleFunc("/base/error", server.ErrorHandler)
	router.HandleFunc("/base/timeout", server.TimeoutHandler)
	// 4. 配置服务器
	realServer := &http.Server{
		Addr: server.RemoteAddr,
		// 写入超时: 从接收 HTTP Request Header结束到完成 HTTP Response 写入结束为止所花的时间
		WriteTimeout: time.Second * 10,
		Handler:      router,
	}
	// 5. 协程启动
	go func() {
		log.Fatalln(realServer.ListenAndServe())
	}()
}

func (server *RealServer) HelloHandler(responseWriter http.ResponseWriter, request *http.Request) {
	urlAndPath := fmt.Sprintf("https://%s%s\n", request.RemoteAddr, request.URL.Path)
	if _, err := io.WriteString(responseWriter, urlAndPath); err != nil {
		log.Fatalf("url and path write response writer err: %v", err)
	}

}

func (server *RealServer) ErrorHandler(responseWriter http.ResponseWriter, request *http.Request) {
	responseWriter.WriteHeader(StatusServerError)
	if _, err := io.WriteString(responseWriter, "error handler"); err != nil {
		log.Fatalf("url and path write response writer err: %v", err)
	}
}

func (server *RealServer) TimeoutHandler(responseWriter http.ResponseWriter, request *http.Request) {
	log.Println("time out handler")
	time.Sleep(time.Second * 6)
	responseWriter.WriteHeader(http.StatusOK)
	if _, err := io.WriteString(responseWriter, "timeout handler"); err != nil {
		log.Fatalf("url and path write response writer err: %v", err)
	}
}

func main() {
	// 1. 创建服务器
	realServerA := &RealServer{RemoteAddr: "127.0.0.1:8081"}
	realServerA.Run()
	realServerB := &RealServer{RemoteAddr: "127.0.0.1:8082"}
	realServerB.Run()
	// 2. 创建信号管道
	quit := make(chan os.Signal)
	// 3. 监听信号
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	// 4. 获取监听到的信号
	<-quit
}
