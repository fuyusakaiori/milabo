package ziface

// IHandler 处理器
type IHandler interface {
	// PreHandle 前置处理
	PreHandle(request IRequest)
	// Handle 处理
	Handle(request IRequest)
	// PostHandle 后置处理
	PostHandle(request IRequest)
	// TODO 指针和接口的关系?
}
