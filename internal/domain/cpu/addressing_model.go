package cpu

// AddressingMode 引数の受け取り方
// document: https://www.nesdev.org/wiki/CPU_addressing_modes
type AddressingMode int

const (
	None AddressingMode = iota
	Implied
	Accumulator
	Immediate
	Zeropage
	ZeropageX
	ZeropageY
	Relative
	Absolute
	AbsoluteX
	AbsoluteY
	Indirect
	IndirectX
	IndirectY // IndirectXとは挙動が違うため注意すること
)
