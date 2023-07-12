package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
)

// Proxy 代理服务器
type Proxy struct {
}

func (proxy *Proxy) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	// 1. 获取上游请求
	method, host, remoteAddr := request.Method, request.Host, request.RemoteAddr
	fmt.Printf("received request - method = %v, host = %v, remoteAddr = %v", method, host, remoteAddr)
	// 2. 复制上游请求并增加新的信息
	proxyRequest := new(http.Request)
	// 2.1 浅拷贝请求对象 - 避免后续修改内容影响到原来请求 - 使用值复制只会将指针指向同一块地址
	*proxyRequest = *request
	// 2.2 在请求中追加请求发送方的 IP 地址 - 最终接收方就可以知道请求经过了多少层代理
	if clientIP, _, err := net.SplitHostPort(remoteAddr); err == nil {
		// 2.2.1 判断请求头中 X-Forwarded-For 字段是否已经有 IP 信息 -> 如果有的话, 就需要分割 slice 然后将发送方的 IP 信息追加
		if prior, exists := proxyRequest.Header["X-Forwarded-For"]; exists {
			clientIP = strings.Join(prior, ",") + ", " + clientIP
		}
		// 2.2.2 设置 X-Forwarded-For 字段
		proxyRequest.Header.Set("X-Forwarded-For", clientIP)
	}
	// 3. 向下游发送请求
	// 3.1 获取传输层
	tcp := http.DefaultTransport
	// 3.2 通过 tcp 发送 http 请求
	proxyResponse, err := tcp.RoundTrip(proxyRequest)
	defer proxyResponse.Body.Close()
	if err != nil {
		responseWriter.WriteHeader(http.StatusBadGateway)
		return
	}
	// 4. 向上游返回响应 - 将代理获取的响应拷贝到将要返回的结果中
	for key, value := range proxyResponse.Header {
		for _, field := range value {
			responseWriter.Header().Add(key, field)
		}
	}
	responseWriter.WriteHeader(proxyResponse.StatusCode)
	if _, err := io.Copy(responseWriter, proxyResponse.Body); err != nil {
		responseWriter.WriteHeader(http.StatusBadGateway)
		return
	}

}

// TODO 为什么启动服务器后访问网站失败, 甚至都没有请求进来
func main() {
	fmt.Println("start...")
	http.Handle("/", &Proxy{})
	_ = http.ListenAndServe("127.0.0.1:8080", nil)
}
