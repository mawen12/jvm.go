package rtda

import (
	"fmt"
)

// jvm stack
type Stack struct {
	// 栈的最大深度，被
	maxSize uint
	size    uint
	_top    *Frame // stack is implemented as linked list
}

func newStack(maxSize uint) *Stack {
	return &Stack{maxSize, 0, nil}
}

func (stack *Stack) isEmpty() bool {
	return stack._top == nil
}

// push 将 frame 压入 stack 中
func (stack *Stack) push(frame *Frame) {
	// 检查是否超过了 stack 的最大深度，超过了则抛出 StackOverflowError 错误
	if stack.size >= stack.maxSize {
		// todo
		panic("StackOverflowError")
	}

	// 如果栈顶已有其他 frame，则将 frame 的下一个指向原先的栈顶
	if stack._top != nil {
		frame.lower = stack._top
	}

	// 更新栈顶为刚插入的 frame
	stack._top = frame
	// 栈大小+1
	stack.size++
}

func (stack *Stack) pop() *Frame {
	if stack._top == nil {
		panic("jvm stack is empty!")
	}

	top := stack._top
	stack._top = top.lower
	top.lower = nil
	stack.size--

	return top
}

func (stack *Stack) clear() {
	for !stack.isEmpty() {
		stack.pop()
	}
}

func (stack *Stack) top() *Frame {
	if stack._top == nil {
		panic("jvm stack is empty!")
	}

	return stack._top
}

func (stack *Stack) topN(n uint) *Frame {
	if stack.size < n {
		panic(fmt.Sprintf("jvm stack size:%v n:%v", stack.size, n))
	}

	frame := stack._top
	for n > 0 {
		frame = frame.lower
		n--
	}

	return frame
}
