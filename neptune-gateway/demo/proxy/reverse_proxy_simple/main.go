package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

const (
	proxyAddr string = "127.0.0.1:8083"
)

func main() {
	// 1. 被代理的下游服务器地址 -> 被代理的服务端的 URL 会拼接上请求的 URL, 默认的 URL 重写规则
	beProxyAddr := "http://127.0.0.1:8081/base"
	// 2. 字符串解析成 beProxyUrl
	beProxyUrl, err := url.Parse(beProxyAddr)
	if err != nil {
		return
	}
	// 3. 调用内部工具类创建反向代理服务器 - 相当于 handler 而不是 server
	reverseProxy := httputil.NewSingleHostReverseProxy(beProxyUrl)
	// 4. 启动反向代理服务器
	log.Println("starting reverse proxy: " + proxyAddr)
	log.Fatalln(http.ListenAndServe(proxyAddr, reverseProxy))
}
