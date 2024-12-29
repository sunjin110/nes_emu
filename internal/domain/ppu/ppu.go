package ppu

type PPU struct {
	registers [8]byte // PPUレジスタ(0x2000〜0x2007)　0x2008-0x3fffはミラー
}

func NewPPU() *PPU {
	return &PPU{
		registers: [8]byte{},
	}
}

const (
	addrPPUStart       = 0x2000
	addrPPUMirrorStart = 0x2008
	addrPPUEnd         = 0x3FFF
)

func (p *PPU) Read(addr uint16) byte {
	return p.registers[(addr-addrPPUStart)%(addrPPUMirrorStart-addrPPUStart)]
}

func (p *PPU) Write(addr uint16, value byte) {
	p.registers[(addr-addrPPUStart)%(addrPPUMirrorStart-addrPPUStart)] = value
	// TODO 副作用の処理をかく
}

func IsPPUAddrRange(addr uint16) bool {
	return addr >= addrPPUStart && addr <= addrPPUEnd
}
