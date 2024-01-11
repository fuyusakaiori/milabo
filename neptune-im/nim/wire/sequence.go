package wire

import (
	"math"
	"sync/atomic"
)

var Seq = sequence{number: 1}

// sequence 序列号
type sequence struct {
	number uint32
}

// Next 获取下一个序列号
func (seq *sequence) Next() uint32 {
	// 1. 原子性递增
	next := atomic.AddUint32(&seq.number, 1)
	// 2. 判断序列号是否已经用尽
	if next == math.MaxUint32 {
		// 2.1 如果已经用尽, 那么重新置为初始值
		if atomic.CompareAndSwapUint32(&seq.number, next, 1) {
			return 1
		}
		// 2.2 如果 cas 失败, 那么递归调用
		return seq.Next()
	}
	return next
}