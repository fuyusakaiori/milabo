package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"neptune-go/src/zinx/ziface"
)

type Configuration struct {
	// Server
	Server    ziface.IServer
	IP        string
	IPVersion string
	Port      uint32
	Name      string

	// Zinx
	ZinxVersion        string
	ZinxMaxConn        uint32
	ZinxMaxPackage     uint32
	ZinxWorkerPoolSize uint32
	ZinxTaskQueueSize  uint32
}

func (config *Configuration) Reload() {
	// 1. 读取配置文件
	data, err := ioutil.ReadFile("config/zinx.json")
	if err != nil {
		fmt.Println("[zinx] reload zinx config err", err)
		return
	}
	// 2. JSON -> Config
	if err := json.Unmarshal(data, &Config); err != nil {
		fmt.Println("[zinx] convert zinx config err", err)
		return
	}
}

var Config *Configuration

// 默认在导包的时候执行, 仅执行一次
func init() {
	// 1. 执行默认配置
	Config = &Configuration{
		Name:               "ZinxServer",
		IP:                 "0.0.0.0",
		IPVersion:          "tcp4",
		Port:               8999,
		ZinxVersion:        "V0.4",
		ZinxMaxConn:        1000,
		ZinxMaxPackage:     4096,
		ZinxWorkerPoolSize: 10,
		ZinxTaskQueueSize:  100,
	}
	// 2. 执行性自定义配置
	Config.Reload()
}
