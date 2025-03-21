package dao

import (
	"project-user/internal/database"
	"project-user/internal/database/gorms"
)

// TransactionImpl 结构体实现了 Transaction 接口，用于管理数据库事务。
type TransactionImpl struct {
	conn database.DbConn // conn 存储了数据库连接对象，用于执行数据库操作。
}

// Action 方法用于执行事务内的数据库操作。
// 参数 f 是一个接受数据库连接对象并返回错误的函数，用于在事务内执行具体操作。
// 返回值是错误，如果事务操作失败，则返回相应的错误。
func (t *TransactionImpl) Action(f func(conn database.DbConn) error) error {
	t.conn.Begin()   // 开始一个新的事务。
	err := f(t.conn) // 调用传入的函数 f 执行数据库操作。
	if err != nil {
		t.conn.Rollback() // 如果操作失败，回滚事务。
		return err        // 返回错误。
	}
	t.conn.Commit() // 如果操作成功，提交事务。
	return nil      // 成功执行事务，无错误返回。
}

// NewTransaction 函数用于创建一个新的 TransactionImpl 实例。
// 返回值是 *TransactionImpl，一个指向新创建的 TransactionImpl 实例的指针。
func NewTransaction() *TransactionImpl {
	return &TransactionImpl{
		conn: gorms.NewTran(), // 使用 gorms 包的 NewTran 函数初始化 conn 字段。
	}
}
