package cpu

type InterruptType int

const (
	InterruptTypeBRK InterruptType = iota
	InterruptTypeNMI               // Non Maskable Interrupt
	InterruptTypeReset
	InterruptTypeIRQ
)
