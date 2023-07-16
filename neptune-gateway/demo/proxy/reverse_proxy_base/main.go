package main

import (
	"bufio"
	"log"
	"net/http"
	"net/url"
)

const (
	proxyServerAddr   string = "127.0.0.1:8080"
	beProxyServerAddr        = "http://127.0.0.1:8081"
)

type ProxyServer struct {
	RemoteAddr string
}

func (server *ProxyServer) Run() {
	log.Printf("starting proxy server :%s", server.RemoteAddr)
	router := http.NewServeMux()
	router.HandleFunc("/", server.handler)
	realServer := &http.Server{
		Addr:    server.RemoteAddr,
		Handler: router,
	}
	log.Fatalln(realServer.ListenAndServe())
}

func (server *ProxyServer) handler(responseWriter http.ResponseWriter, request *http.Request) {
	// 1. 解析被代理的服务器的地址
	beProxy, err := url.Parse(beProxyServerAddr)
	if err != nil {
		return
	}
	// 2. 修改请求体内容 - 之前客户端请求的是代理服务器而不是真实服务器
	request.URL.Scheme = beProxy.Scheme
	request.URL.Host = beProxy.Host
	// 3. 向被代理的服务器发送请求
	tcp := http.DefaultTransport
	response, err := tcp.RoundTrip(request)
	defer response.Body.Close()
	if err != nil {
		return
	}
	// 4. 向上游返回数据
	if _, err := bufio.NewReader(response.Body).WriteTo(responseWriter); err != nil {
		responseWriter.WriteHeader(http.StatusBadGateway)
		return
	}

}

func main() {
	// URL 重写规则: 被代理服务器 URL 是 "/", 请求中的 URL 就需要写全 "/base/timeout"
	server := &ProxyServer{proxyServerAddr}
	server.Run()

}
