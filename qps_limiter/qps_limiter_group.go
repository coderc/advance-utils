package qps_limiter

import (
	"fmt"
	"sync/atomic"
	"time"
)

type (
	// QPSLimiterGroup qps限制器组(支持以字符串为key的多个限制器)
	QPSLimiterGroup struct {
		podCnt        uint32
		qpsLimiterMap map[string]*qpsLimiter
	}
)

// UpdatePodInfo 更新pod信息    ****必须设置 !!没有默认值****
func (q *QPSLimiterGroup) UpdatePodInfo(podCnt int) {
	if q == nil || podCnt <= 0 {
		return
	}

	atomic.StoreUint32(&q.podCnt, uint32(podCnt))
}

// SetTopLimit 设置qps限制
func (q *QPSLimiterGroup) SetTopLimit(key string, topLimit uint32) {
	if q == nil {
		return
	}
	if q.qpsLimiterMap == nil {
		q.qpsLimiterMap = make(map[string]*qpsLimiter)
	}
	if _, ok := q.qpsLimiterMap[key]; !ok {
		q.qpsLimiterMap[key] = &qpsLimiter{}
	}
	q.qpsLimiterMap[key].setTopLimit(topLimit)
}

// Check 检查当前时刻是否超过限制
func (q *QPSLimiterGroup) Check(key string) (bool, error) {
	return q.checkWithTime(key, time.Now())
}

func (q *QPSLimiterGroup) checkWithTime(key string, timeStamp time.Time) (bool, error) {
	if q == nil || q.podCnt <= 0 {
		return false, fmt.Errorf("pod info is not complete")
	}
	if _, ok := q.qpsLimiterMap[key]; !ok {
		return false, fmt.Errorf("key %s is not exist", key)
	}
	return q.qpsLimiterMap[key].consumeWithCheck(q.podCnt, timeStamp), nil
}

// GC 清理过期的qps消耗记录
func (q *QPSLimiterGroup) GC() {
	timeStamp := time.Now().Add(-time.Second * 10).Unix()
	for _, qpsLimiter := range q.qpsLimiterMap {
		qpsLimiter.qpsConsumed.Range(func(key, value any) bool {
			if key.(int64) < timeStamp {
				qpsLimiter.qpsConsumed.Delete(key)
			}
			return true
		})
	}
}
