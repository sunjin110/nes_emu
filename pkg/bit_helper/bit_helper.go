package bit_helper

func BytesToUint16(lower byte, upper byte) uint16 {
	return uint16(lower) | (uint16(upper) << 8)
}
