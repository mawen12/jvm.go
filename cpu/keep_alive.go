package cpu

import (
	"sync"
)

// TODO: use WaitGroup?
var (
	// 统计非 daemon 线程数量的计数器
	aliveCount = 0
	lock       = &sync.Mutex{}
	cond       = sync.NewCond(lock)
)

// nonDaemonThreadStart 非 daemon 线程 + 1
func nonDaemonThreadStart() {
	lock.Lock()
	defer lock.Unlock()

	aliveCount++
}

// nonDaemonThreadStop 非 daemon 线程停止，数量 - 1，当为0时，唤醒等待的 cond
func nonDaemonThreadStop() {
	lock.Lock()
	defer lock.Unlock()

	// 非 daemon 线程数 - 1
	aliveCount--
	// 如果非 daemon 线程数为0，则使用通知，会唤醒正在 Wait 的 cond
	if aliveCount == 0 {
		cond.Broadcast()
	}
}

// KeepAlive 当 非 daemon 线程数量 > 0 时，阻塞等待
func KeepAlive() {
	lock.Lock()
	defer lock.Unlock()

	// 当非 daemon 线程数量 > 0 时，进行等待状态
	if aliveCount > 0 {
		cond.Wait()
	}
}
