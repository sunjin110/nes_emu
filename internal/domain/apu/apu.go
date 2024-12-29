package apu

// Audio Processing Unit

type APU struct {
	registers [0x16]byte
}

const (
	addrAPUIOStart = 0x4000
	addrAPUIOEnd   = 0x4015
)

func NewAPU() *APU {
	return &APU{
		registers: [0x16]byte{},
	}
}

func (a *APU) Read(addr uint16) byte {
	return a.registers[addr-addrAPUIOStart]
}

func (a *APU) Write(addr uint16, value byte) {
	a.registers[addr-addrAPUIOStart] = value
	// TODO 必要に応じて音声処理の副作用
}

func IsAPUAddrRange(addr uint16) bool {
	return addr >= addrAPUIOStart && addr <= addrAPUIOEnd
}
