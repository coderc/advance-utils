package qps_limiter

import (
	"sync"
	"sync/atomic"
	"time"
)

// qps 限制器(单个limiter)
type (
	qpsLimiter struct {
		qpsTopLimit uint32
		qpsConsumed sync.Map
	}
)

// setTopLimit 设置qps限制
func (q *qpsLimiter) setTopLimit(qpsTopLimit uint32) {
	if q == nil {
		return
	}

	atomic.StoreUint32(&q.qpsTopLimit, qpsTopLimit)
}

// consumeWithCheck 消耗当前时刻的qps并检查是否超过限制
func (q *qpsLimiter) consumeWithCheck(podCnt uint32, timeStamp time.Time) bool {
	if q == nil {
		return false
	}

	timeStampSecond := timeStamp.Unix()

	// 如果没有消耗过，则初始化为0
	var v uint32 = 0
	ptrAny, _ := q.qpsConsumed.LoadOrStore(timeStampSecond, &v)

	ptr := ptrAny.(*uint32)
	return atomic.AddUint32(ptr, 1) <= atomic.LoadUint32(&q.qpsTopLimit)/podCnt
}
