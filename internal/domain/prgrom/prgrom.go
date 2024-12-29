package prgrom

const (
	PRGROMSize = 32 * 1024 // 32KB

	addrPRGROMStart = 0x8000
	addrPRGROMEnd   = 0xFFFF
)

// TODO bank switchをする場合はここの実装を変更する必要がある
// 現在は32KBまでのPRGROMデータのみ対応する
type PRGROM struct {
	data [PRGROMSize]byte // 32KB
}

func NewPRGROM(data [PRGROMSize]byte) *PRGROM {
	return &PRGROM{
		data: data,
	}
}

func (rom *PRGROM) Read(addr uint16) byte {
	return rom.data[rom.relativeAddr(addr)]
}

// InitPC program counterの初期値を取得する
func (rom *PRGROM) InitPC() uint16 {
	// https://www.pagetable.com/?p=410
	pcLower := rom.data[rom.relativeAddr(0xFFFC)]
	pcUpper := rom.data[rom.relativeAddr(0xFFFD)]
	pc := uint16(pcLower) | (uint16(pcUpper) << 8)
	return pc
}

func (*PRGROM) relativeAddr(absAddr uint16) uint16 {
	return absAddr - addrPRGROMStart
}

func IsPRGRomRange(addr uint16) bool {
	return addr >= addrPRGROMStart && addr <= addrPRGROMEnd
}
