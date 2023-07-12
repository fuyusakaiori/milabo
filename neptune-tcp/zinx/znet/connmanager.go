package znet

import (
	"errors"
	"fmt"
	"neptune-golang/neptune-tcp/zinx/ziface"
	"sync"
)

type ConnManager struct {
	// 注: 不对外提供的连接集合
	connections map[uint32]ziface.IConnection
	connLock    sync.RWMutex
}

func NewConnManager() ziface.IConnManager {
	return &ConnManager{
		connections: make(map[uint32]ziface.IConnection),
	}
}

func (conn *ConnManager) AddConnection(connection ziface.IConnection) {
	// 1. 上锁
	conn.connLock.Lock()
	defer conn.connLock.Unlock()
	// 2. 添加到集合中
	if _, result := conn.connections[connection.GetConnID()]; result {
		fmt.Println("[zinx] conn already exit, can't add this conn")
		return
	}
	conn.connections[connection.GetConnID()] = connection
	fmt.Println("[zinx] conn add to connections success, count", conn.GetConnectionCount())
}

func (conn *ConnManager) GetConnection(connID uint32) (connection ziface.IConnection, err error) {
	conn.connLock.RLock()
	defer conn.connLock.RUnlock()
	result, ok := conn.connections[connID]
	if !ok {
		fmt.Println("[zinx] conn doesn't exit")
		return nil, errors.New("[zinx] conn doesn't exit")
	}
	return result, nil
}

func (conn *ConnManager) CloseConnection(connection ziface.IConnection) {
	// 1. 上锁
	conn.connLock.Lock()
	defer conn.connLock.Unlock()
	// 2. 检验是否存在
	if _, result := conn.connections[connection.GetConnID()]; !result {
		fmt.Println("[zinx] conn doesn't exit, can't close this conn")
		return
	}
	// 3. 删除
	delete(conn.connections, connection.GetConnID())
	fmt.Println("[zinx] conn close in connection success, count ", conn.GetConnectionCount())
}

func (conn *ConnManager) GetConnectionCount() (count uint32) {
	return uint32(len(conn.connections))
}

func (conn *ConnManager) CloseConnections() {
	conn.connLock.Lock()
	defer conn.connLock.Unlock()
	for connID, connection := range conn.connections {
		// 删除
		delete(conn.connections, connID)
		// 关闭
		connection.StopConn()
		fmt.Println("[zinx] conn ", connID, " close in connection success")
	}
}
