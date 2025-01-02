package bit_helper

func BytesToUint16(lower byte, upper byte) uint16 {
	return uint16(lower) | (uint16(upper) << 8)
}

func Uint16ToBytes(value uint16) (lower byte, upper byte) {
	return byte(value & 0xFF), byte(value >> 8)
}
