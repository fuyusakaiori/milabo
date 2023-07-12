package znet

import (
	"neptune-go/src/zinx/ziface"
)

type BaseHandler struct {
}

func (router *BaseHandler) PreHandle(request ziface.IRequest) {
}

func (router *BaseHandler) Handle(request ziface.IRequest) {

}

func (router *BaseHandler) PostHandle(request ziface.IRequest) {

}
