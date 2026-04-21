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
		// 
		decoded[reader.Position()] = decodeInstruction(reader)
	}

	return decoded
}

// 
func decodeInstruction(reader *base.CodeReader) base.Instruction {
	// 读取 opcode
	opcode := reader.ReadUint8()
	// 
	instr := newInstruction(opcode)
	instr.FetchOperands(reader)
	return instr
}
