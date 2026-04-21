package rtda

import (
	"github.com/zxh0/jvm.go/rtda/heap"
)

/*
	thread
		所属的线程

	cachedFrames
		缓存的帧

	frameCount
		帧总数

	maxFrame
		最大允许缓存的帧数量
*/
type FrameCache struct {
	thread       *Thread
	cachedFrames []*Frame
	frameCount   uint
	maxFrame     uint
}

// newFrameCache 创建线程专属的帧缓存
func newFrameCache(thread *Thread, maxFrame uint) *FrameCache {
	return &FrameCache{
		thread:       thread,
		maxFrame:     maxFrame,
		cachedFrames: make([]*Frame, maxFrame),
	}
}

func (cache *FrameCache) borrowFrame(method *heap.Method) *Frame {
	if cache.frameCount > 0 { // 尝试从缓存中读取
		for i, frame := range cache.cachedFrames {
			if frame != nil &&
				frame.maxLocals >= method.MaxLocals &&
				frame.maxStack >= method.MaxStack {

				cache.frameCount--
				cache.cachedFrames[i] = nil
				frame.reset(method)
				return frame
			}
		}
	}
	// 创建帧
	return newFrame(cache.thread, method)
}

// returnFrame 尝试将帧加入到缓存中
func (cache *FrameCache) returnFrame(frame *Frame) {
	// 检查缓存是否达到上限
	if cache.frameCount < cache.maxFrame { // 未达到上限
		// 遍历帧缓存
		for i, cachedFrame := range cache.cachedFrames {
			// 找到空位
			if cachedFrame == nil {
				// 将frame加入到缓存中
				cache.cachedFrames[i] = frame
				// 缓存总数+1
				cache.frameCount++
				return
			}
		}
	} else { // 已达到上限
		// 遍历帧缓存
		for _, cachedFrame := range cache.cachedFrames {
			// 比较参数帧与缓存帧的最大本地变量数
			if frame.maxLocals > cachedFrame.maxLocals { // 比缓存帧的最大本地变量数要大
				// 使用参数帧覆盖缓存帧的最大本地变量数
				cachedFrame.maxLocals = frame.maxLocals
				// 使用参数帧覆盖缓存帧的本地变量表
				cachedFrame.LocalVars = frame.LocalVars
				// 参数帧的最大本地变量数重置为0,这样下一次的比较中就不会再次进入了
				frame.maxLocals = 0
			}
			// 比较参数帧与缓存帧的最大栈深度
			if frame.maxStack > cachedFrame.maxStack {
				// 使用参数帧覆盖缓存帧的最大栈深度
				cachedFrame.maxStack = frame.maxStack
				// 使用参数帧覆盖缓存帧的操作栈
				cachedFrame.OperandStack = frame.OperandStack
				// 参数帧的最大栈深度重置为0,这样下一次的比较中就不会再次进入了
				frame.maxStack = 0
			}
		}
	}
}
