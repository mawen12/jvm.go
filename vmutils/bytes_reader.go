package vmutils

import (
	"encoding/binary"
)

// 基础的字节阅读器，用于将读取的字节转换为 16-, 32-, or 64-bit 无符号整数.
type BytesReader struct {
	// 实现将字节读取为 16-, 32-, or 64-bit 无符号整数.
	byteOrder binary.ByteOrder
	// 待读取的字节
	data []byte
	// 偏移量，每次读取类型时的起始位置，读取后将偏移量增加对应数量
	// uint8 -> 8-bit -> +1
	// uint16 -> 16-bit -> +2
	// unt32 -> 32-bit -> +4
	// uint64 -> 64-bit -> +8
	position int
}

func NewBytesReader(data []byte, byteOrder binary.ByteOrder) BytesReader {
	return BytesReader{
		byteOrder: byteOrder,
		data:      data,
		position:  0,
	}
}

func (reader *BytesReader) Position() int {
	return reader.position
}

// ReadUint8 读取 1 byte 的内容
func (reader *BytesReader) ReadUint8() uint8 {
	i := reader.data[reader.position]
	reader.position++
	return i
}

// ReadUint8 读取 2 bytes 的内容
func (reader *BytesReader) ReadUint16() uint16 {
	// 将 bytes 转换为 uint16
	i := reader.byteOrder.Uint16(reader.data[reader.position:])
	reader.position += 2
	return i
}

// ReadUint8 读取 4 bytes 的内容
func (reader *BytesReader) ReadUint32() uint32 {
	// 将 bytes 转换为 uint32
	i := reader.byteOrder.Uint32(reader.data[reader.position:])
	reader.position += 4
	return i
}

// ReadUint8 读取 8 bytes 的内容
func (reader *BytesReader) ReadUint64() uint64 {
	// 将 bytes 转换为 uint64
	i := reader.byteOrder.Uint64(reader.data[reader.position:])
	reader.position += 8
	return i
}

// // ReadUint8 读取 n bytes 的内容
func (reader *BytesReader) ReadBytes(n int) []byte {
	bytes := reader.data[reader.position : reader.position+n]
	reader.position += n
	return bytes
}
