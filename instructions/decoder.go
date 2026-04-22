package instructions

import (
	"github.com/zxh0/jvm.go/instructions/base"
)

// Decode 将字节数组解码为指令数组
func Decode(code []byte) []base.Instruction {
	// 创建 code 读取器
	reader := base.NewCodeReader(code)
	// 构造保存指令的数组
	decoded := make([]base.Instruction, len(code))

	// 持续读取
	for reader.Position() < len(code) {
		// 从 reader 中读取一个 opcode，并解析为对应指令，对指令执行初始化，然后保存到指定位置
		decoded[reader.Position()] = decodeInstruction(reader)
	}

	return decoded
}

// decodeInstruction 从 reader 中读取一个 opcode，并解析为对应指令，对指令执行初始化
func decodeInstruction(reader *base.CodeReader) base.Instruction {
	// 读取 opcode
	opcode := reader.ReadUint8()
	// 根据 opcode 创建指令
	instr := newInstruction(opcode)
	// 执行指令内部值的初始化操作
	instr.FetchOperands(reader)
	return instr
}
