package ziface

type IConnManager interface {
	AddConnection(connection IConnection)

	GetConnection(connID uint32) (connection IConnection, err error)

	CloseConnection(connection IConnection)

	GetConnectionCount() (count uint32)

	CloseConnections()
}
