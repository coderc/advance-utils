package qps_limiter

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

/*
$ go test -v -bench=. -run=none -benchmem .
goos: darwin
goarch: arm64
pkg: github.com/coderc/advance-utils/qps_limiter
BenchmarkQPSLimiterGroupWithBigQps
BenchmarkQPSLimiterGroupWithBigQps-8     	1000000000	         0.06043 ns/op	       0 B/op	       0 allocs/op
BenchmarkQPSLimiterGroupWithSmallQps
BenchmarkQPSLimiterGroupWithSmallQps-8   	1000000000	         0.06066 ns/op	       0 B/op	       0 allocs/op
BenchmarkQPSLimiterGroupWithZeroQps
BenchmarkQPSLimiterGroupWithZeroQps-8    	1000000000	         0.06037 ns/op	       0 B/op	       0 allocs/op
PASS
ok  	github.com/coderc/advance-utils/qps_limiter	2.123s
*/

const (
	TESTKey                      = "test qps limiter key 01"
	TESTQPSTopLimitBigQps uint32 = 10000
	TESTTopLimitQps              = 100000
)

func BenchmarkQPSLimiterGroupWithBigQps(b *testing.B) {
	qpsLimiterGroup := new(QPSLimiterGroup)
	qpsLimiterGroup.UpdatePodInfo(1)
	qpsLimiterGroup.SetTopLimit(TESTKey, TESTQPSTopLimitBigQps)

	successCnt := make(map[int64]*uint32)
	failCnt := make(map[int64]*uint32)

	b.ResetTimer()

	for i := 0; i < 30; i++ {
		timeStamp := time.Now().Add(time.Duration(i) * time.Second).Unix()
		var successCntNum uint32 = 0
		successCnt[timeStamp] = &successCntNum

		var failCntNum uint32 = 0
		failCnt[timeStamp] = &failCntNum
	}

	g := sync.WaitGroup{}
	g.Add(TESTTopLimitQps)
	for i := 0; i < TESTTopLimitQps; i++ {
		go func() {
			defer g.Done()
			timeStamp := time.Now()
			ok, err := qpsLimiterGroup.checkWithTime(TESTKey, timeStamp)
			assert.NoErrorf(b, err, "ERROR check qps limiter failed")

			if ok {
				atomic.AddUint32(successCnt[timeStamp.Unix()], 1)
			} else {
				atomic.AddUint32(failCnt[timeStamp.Unix()], 1)
			}
		}()
	}
	g.Wait()

	for stamp, cntPtr := range successCnt {
		sucCnt := atomic.LoadUint32(cntPtr)
		failCnt := atomic.LoadUint32(failCnt[stamp])
		if sucCnt == 0 && failCnt == 0 {
			continue
		}
		//fmt.Printf("----\npass qps cnt at %d: %d\nfail qps cnt at %d: %d\n----\n", stamp, sucCnt, stamp, failCnt)
		assert.True(b, TESTQPSTopLimitBigQps >= sucCnt, fmt.Sprintf("ERROR pass qps cnt(%d) is greater than top limit(%d)", sucCnt, TESTQPSTopLimitBigQps))
	}

	qpsLimiterGroup.GC()
}

const (
	TESTQPSTopLimitSmallQps uint32 = 1
)

func BenchmarkQPSLimiterGroupWithSmallQps(b *testing.B) {
	qpsLimiterGroup := new(QPSLimiterGroup)
	qpsLimiterGroup.UpdatePodInfo(1)
	qpsLimiterGroup.SetTopLimit(TESTKey, TESTQPSTopLimitSmallQps)

	successCnt := make(map[int64]*uint32)
	failCnt := make(map[int64]*uint32)

	b.ResetTimer()

	for i := 0; i < 30; i++ {
		timeStamp := time.Now().Add(time.Duration(i) * time.Second).Unix()
		var successCntNum uint32 = 0
		successCnt[timeStamp] = &successCntNum

		var failCntNum uint32 = 0
		failCnt[timeStamp] = &failCntNum
	}

	g := sync.WaitGroup{}
	g.Add(TESTTopLimitQps)
	for i := 0; i < TESTTopLimitQps; i++ {
		go func() {
			defer g.Done()
			timeStamp := time.Now()
			ok, err := qpsLimiterGroup.checkWithTime(TESTKey, timeStamp)
			assert.NoErrorf(b, err, "ERROR check qps limiter failed")

			if ok {
				atomic.AddUint32(successCnt[timeStamp.Unix()], 1)
			} else {
				atomic.AddUint32(failCnt[timeStamp.Unix()], 1)
			}
		}()
	}
	g.Wait()

	for stamp, cntPtr := range successCnt {
		sucCnt := atomic.LoadUint32(cntPtr)
		failCnt := atomic.LoadUint32(failCnt[stamp])
		if sucCnt == 0 && failCnt == 0 {
			continue
		}
		//fmt.Printf("----\npass qps cnt at %d: %d\nfail qps cnt at %d: %d\n----\n", stamp, sucCnt, stamp, failCnt)
		assert.True(b, TESTQPSTopLimitSmallQps >= sucCnt, fmt.Sprintf("ERROR pass qps cnt(%d) is greater than top limit(%d)", sucCnt, TESTQPSTopLimitSmallQps))
	}

	qpsLimiterGroup.GC()
}

const (
	TESTQPSTopLimitZeroQps uint32 = 0
)

func BenchmarkQPSLimiterGroupWithZeroQps(b *testing.B) {
	qpsLimiterGroup := new(QPSLimiterGroup)
	qpsLimiterGroup.UpdatePodInfo(1)
	qpsLimiterGroup.SetTopLimit(TESTKey, TESTQPSTopLimitZeroQps)

	successCnt := make(map[int64]*uint32)
	failCnt := make(map[int64]*uint32)

	b.ResetTimer()

	for i := 0; i < 30; i++ {
		timeStamp := time.Now().Add(time.Duration(i) * time.Second).Unix()
		var successCntNum uint32 = 0
		successCnt[timeStamp] = &successCntNum

		var failCntNum uint32 = 0
		failCnt[timeStamp] = &failCntNum
	}

	g := sync.WaitGroup{}
	g.Add(TESTTopLimitQps)
	for i := 0; i < TESTTopLimitQps; i++ {
		go func() {
			defer g.Done()
			timeStamp := time.Now()
			ok, err := qpsLimiterGroup.checkWithTime(TESTKey, timeStamp)
			assert.NoErrorf(b, err, "ERROR check qps limiter failed")

			if ok {
				atomic.AddUint32(successCnt[timeStamp.Unix()], 1)
			} else {
				atomic.AddUint32(failCnt[timeStamp.Unix()], 1)
			}
		}()
	}
	g.Wait()

	for stamp, cntPtr := range successCnt {
		sucCnt := atomic.LoadUint32(cntPtr)
		failCnt := atomic.LoadUint32(failCnt[stamp])
		if sucCnt == 0 && failCnt == 0 {
			continue
		}
		//fmt.Printf("----\npass qps cnt at %d: %d\nfail qps cnt at %d: %d\n----\n", stamp, sucCnt, stamp, failCnt)
		assert.True(b, TESTQPSTopLimitZeroQps >= sucCnt, fmt.Sprintf("ERROR pass qps cnt(%d) is greater than top limit(%d)", sucCnt, TESTQPSTopLimitZeroQps))
	}

	qpsLimiterGroup.GC()
}

/*


 */
