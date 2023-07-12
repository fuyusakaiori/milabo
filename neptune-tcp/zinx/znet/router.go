package znet

import (
	"fmt"
	"math/rand"
	"neptune-golang/neptune-tcp/zinx/utils"
	"neptune-golang/neptune-tcp/zinx/ziface"
)

type Router struct {
	// 处理器集合
	Apis map[uint32]ziface.IHandler
	// 消息队列集合: 可以只使用一个管道作为消息队列
	TaskQueues []chan ziface.IRequest
	// 最大协程数量
	MaxWorkerPoolSize uint32
}

func NewRouter() ziface.IRouter {
	return &Router{
		Apis:              make(map[uint32]ziface.IHandler),
		TaskQueues:        make([]chan ziface.IRequest, utils.Config.ZinxWorkerPoolSize),
		MaxWorkerPoolSize: utils.Config.ZinxWorkerPoolSize,
	}
}

func (router *Router) RouterHandler(request ziface.IRequest) {
	// 1. 获取处理器: 如果没有找到, 那么返回类型对应的零值; 如果存在, 那么就返回对应值
	handler, result := router.Apis[request.GetMessage().GetMessageID()]
	// 2. 检查是否存在
	if !result {
		fmt.Println("[zinx] router not found handler to handle message")
		return
	}
	// 3. 处理消息
	handler.PreHandle(request)
	handler.Handle(request)
	handler.PostHandle(request)
}

func (router *Router) AddHandler(id uint32, handler ziface.IHandler) {
	// 1. 检查是否存在
	if _, result := router.Apis[id]; result {
		fmt.Println("[zinx] already exit same message id handler in router ")
		return
	}
	// 2. 添加处理器
	router.Apis[id] = handler
}

func (router *Router) StartWorkerPool() {
	fmt.Println("[zinx] starting worker pool: worker size ", router.MaxWorkerPoolSize, " queue size", utils.Config.ZinxTaskQueueSize)
	// 直接启动
	for index := 0; index < int(router.MaxWorkerPoolSize); index++ {
		fmt.Println("[zinx] worker pool start ", index, " goroutine ")
		// 开启协程
		router.TaskQueues[index] = make(chan ziface.IRequest, utils.Config.ZinxTaskQueueSize)
		go router.StartWorker(router.TaskQueues[index])
	}
}

func (router *Router) StartWorker(taskQueue chan ziface.IRequest) {
	for {
		select {
		case request := <-taskQueue:
			router.RouterHandler(request)
		}
	}
}

func (router *Router) SendMessageToTaskQueue(request ziface.IRequest) {
	// 1. 随机负载均衡
	id := rand.Int31n(int32(router.MaxWorkerPoolSize))
	// 2. 选择消息队列, 发送消息
	fmt.Println("[zinx] read goroutine send message to no.", id, " task queue")
	router.TaskQueues[id] <- request
}
