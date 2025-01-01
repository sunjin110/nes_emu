package cpu

type InterruptType int

const (
	InterruptTypeBRK InterruptType = iota
	InterruptTypeNMI
	InterruptTypeReset
	InterruptTypeIRQ
)
