package prgrom

import "github.com/sunjin110/nes_emu/pkg/bit_helper"

const (
	PRGROMSize = 32 * 1024 // 32KB

	addrPRGROMStart = 0x8000
	addrPRGROMEnd   = 0xFFFF
)

// TODO bank switchをする場合はここの実装を変更する必要がある
// 現在は32KBまでのPRGROMデータのみ対応する
type FixedPRGROM struct {
	data [PRGROMSize]byte // 32KB
}

func NewFixedPRGROM(data [PRGROMSize]byte) PRGROM {
	return &FixedPRGROM{
		data: data,
	}
}

func (rom *FixedPRGROM) Read(addr uint16) byte {
	return rom.data[rom.relativeAddr(addr)]
}

// InitPC program counterの初期値を取得する
func (rom *FixedPRGROM) InitPC() uint16 {
	// https://www.pagetable.com/?p=410
	pcLower := rom.data[rom.relativeAddr(0xFFFC)]
	pcUpper := rom.data[rom.relativeAddr(0xFFFD)]
	pc := bit_helper.BytesToUint16(pcLower, pcUpper)
	return pc
}

func (*FixedPRGROM) relativeAddr(absAddr uint16) uint16 {
	return absAddr - addrPRGROMStart
}

func IsPRGRomRange(addr uint16) bool {
	return addr >= addrPRGROMStart && addr <= addrPRGROMEnd
}
