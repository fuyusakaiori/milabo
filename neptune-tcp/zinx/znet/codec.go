package znet

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"neptune-golang/neptune-tcp/zinx/utils"
	"neptune-golang/neptune-tcp/zinx/ziface"
)

type Codec struct {
}

func NewCodec() (codec ziface.ICodec) {
	return &Codec{}
}

func (codec *Codec) GetHeadLength() uint32 {
	// 序列号 4B + 消息长度 4B
	return 8
}

func (codec *Codec) Encode(message ziface.IMessage) (data []byte, err error) {
	// 1. 创建缓冲区
	buf := bytes.NewBuffer([]byte{})
	// 2. 写入序列号
	if err := binary.Write(buf, binary.LittleEndian, message.GetMessageID()); err != nil {
		fmt.Println("[zinx] write message id err", err)
		return nil, err
	}
	// 3. 写入消息长度
	if err := binary.Write(buf, binary.LittleEndian, message.GetMessageLength()); err != nil {
		fmt.Println("[zinx] write message length err", err)
		return nil, err
	}
	// 4. 写入消息内容
	if err := binary.Write(buf, binary.LittleEndian, message.GetMessageData()); err != nil {
		fmt.Println("[zinx] write message data err", err)
		return nil, err
	}
	return buf.Bytes(), nil
}

func (codec *Codec) Decode(data []byte) (message ziface.IMessage, err error) {
	// 1. 获取输入流
	reader := bytes.NewReader(data)
	response := &Message{}
	// 2. 读取消息序列号
	// TODO 为什么要取址
	if err := binary.Read(reader, binary.LittleEndian, &response.MessageID); err != nil {
		fmt.Println("[zinx] read message id err", err)
		return nil, err
	}
	// 3. 读取消息长度
	if err := binary.Read(reader, binary.LittleEndian, &response.MessageLength); err != nil {
		fmt.Println("[zinx] read message length err", err)
		return nil, err
	}
	// 4. 判断消息长度是否超过限制: 如果超过限制, 直接抛出异常
	if utils.Config.ZinxMaxPackage > 0 && response.GetMessageLength() > utils.Config.ZinxMaxPackage {
		return nil, errors.New("[zinx] receive package size too large")
	}
	return response, nil
}
