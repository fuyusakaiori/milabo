package main

import (
	"bytes"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
)

const (
	proxyAddr string = "127.0.0.1:8083"
)

// 方法作用: 拼接代理服务器接收到的请求和将要发送的请求的 URL
// 返回结果: 第一个是返回的部分字符未转义的 URL, 第二个是返回的所有字符都会被正确转义的 URL
func joinUrlPath(targetUrl, requestUrl *url.URL) (path, rawPath string) {
	// 1. URL 转义成字符串
	targetPath, requestPath := targetUrl.EscapedPath(), requestUrl.EscapedPath()
	// 2. 判断被代理的服务器 URL 是否有 "/" 后缀; 判断接收到的请求是否有 "/" 前缀
	isTargetSlash := strings.HasSuffix(targetPath, "/")
	isRequestSlash := strings.HasPrefix(requestPath, "/")
	// 3. 根据前缀的情况判断如何组合
	switch {
	case isTargetSlash && isRequestSlash:
		return targetUrl.Path + requestUrl.Path[1:], targetPath + requestPath[1:]
	case !isRequestSlash && !isTargetSlash:
		return targetUrl.Path + "/" + requestUrl.Path, targetPath + "/" + requestPath
	}
	return targetUrl.Path + requestUrl.Path, targetPath + requestPath
}

func NewSingleHostReverseProxy(target *url.URL) *httputil.ReverseProxy {

	// 1. 创建修改请求的方法
	director := func(request *http.Request) {
		// 1. 修改请求中的协议、主机、路径 -> 指向被代理的服务器
		request.URL.Scheme = target.Scheme
		request.URL.Host = target.Host
		request.URL.Path, request.URL.RawPath = joinUrlPath(target, request.URL)
		// 2. 修改请求中路径参数
		if request.URL.RawQuery == "" || target.RawQuery == "" {
			request.URL.RawQuery = target.RawQuery + request.URL.RawQuery
		} else {
			request.URL.RawQuery = target.RawQuery + "&" + request.URL.RawQuery
		}
		// 3. 设置请求中的浏览器 -> 如果没有设置浏览器, 那么就赋值为空
		if _, exists := request.Header["User-Agent"]; !exists {
			request.Header.Set("User-Agent", "")
		}
	}
	// 2. 创建修改响应体的方法
	modify := func(response *http.Response) error {
		if response.StatusCode != http.StatusOK {
			return errors.New("reverse proxy receive response status code not equal 200")
		}
		// 1. 获取响应体中的内容
		oldPayLoad, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Fatalf("reverse proxy read response err: %v", err)
		}
		// 2. 添加新的内容
		newPayLoad := []byte("neptune golang reverse proxy " + string(oldPayLoad))
		// 3. 重新更新响应体中的内容
		response.Body = ioutil.NopCloser(bytes.NewBuffer(newPayLoad))
		response.ContentLength = int64(len(newPayLoad))
		response.Header.Set("Content-Length", strconv.Itoa(len(newPayLoad)))
		return nil
	}

	return &httputil.ReverseProxy{Director: director, ModifyResponse: modify}
}

func main() {
	// 1. 被代理的下游服务器地址 -> 被代理的服务端的 URL 会拼接上请求的 URL, 默认的 URL 重写规则
	beProxyAddr := "http://127.0.0.1:8081/base"
	// 2. 字符串解析成 beProxyUrl
	beProxyUrl, err := url.Parse(beProxyAddr)
	if err != nil {
		return
	}
	// 3. 调用内部工具类创建反向代理服务器 - 相当于 handler 而不是 server
	reverseProxy := NewSingleHostReverseProxy(beProxyUrl)
	// 4. 启动反向代理服务器
	log.Println("starting reverse proxy: " + proxyAddr)
	log.Fatalln(http.ListenAndServe(proxyAddr, reverseProxy))
}
