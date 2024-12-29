package memory

import (
	"fmt"

	"github.com/sunjin110/nes_emu/internal/domain/ram"
	"github.com/sunjin110/nes_emu/pkg/logger"
)

type Memory struct {
	ram    ram.RAM          // RAM:ワーキングメモリ(0x0000-0x07ff) 0x0800-0x1fffはミラー
	ppu    [8]byte          // PPUレジスタ(0x2000〜0x2007)　0x2008-0x3fffはミラー
	io     [0x20]byte       // I/Oレジスタ(0x4000〜0x401F)
	prgRom [PRGROMSize]byte // PRG-ROM(0x8000〜0xFFFF)
}

func NewMemory(prgRom [PRGROMSize]byte) *Memory {
	return &Memory{
		ram:    *ram.NewRAM(),
		ppu:    [8]byte{},
		io:     [0x20]byte{},
		prgRom: prgRom,
	}
}

const (
	PRGROMSize = 0x8000
)

const (
	addrRAMStart          = 0x0000
	addrRAMMirrorStart    = 0x0800
	addrRAMEnd            = 0x1FFF
	addrPPUStart          = 0x2000
	addrPPUMirrorStart    = 0x2008
	addrPPUEnd            = 0x3FFF
	addrAPUIOStart        = 0x4000
	addrAPUIOEnd          = 0x4015
	addrControllerIOStart = 0x4016
	addrControllerIOEnd   = 0x4017
	addrPRGROMStart       = 0x8000
	addrPRGROMEnd         = 0xFFFF
)

func (memory *Memory) Read(addr uint16) (byte, error) {
	switch {
	case addr >= addrRAMStart && addr <= addrRAMEnd: // RAM
		return memory.ram.Read(addr), nil
	case addr >= addrPPUStart && addr <= addrPPUEnd: // PPU
		return memory.ppu[(addr-addrPPUStart)%(addrPPUMirrorStart-addrPPUStart)], nil
	case addr >= addrAPUIOStart && addr <= addrAPUIOEnd: // APU I/O
		return memory.io[(addr - addrAPUIOStart)], nil
	case addr >= addrControllerIOStart && addr <= addrControllerIOEnd:
		return memory.io[(addr - addrAPUIOStart)], nil
	case addr >= addrPRGROMStart && addr <= addrPRGROMEnd:
		return memory.prgRom[(addr - addrPRGROMStart)], nil
	default:
		logger.Logger.Error("invalid addr is specified", "addr", addr)
		return 0, fmt.Errorf("Memory: invalid addr is specified. addr: %b", addr)
	}
}

func (memory *Memory) Write(addr uint16, value byte) error {
	switch {
	case addr >= addrRAMStart && addr <= addrRAMEnd: // RAM
		memory.ram.Write(addr, value)
	case addr >= addrPPUStart && addr <= addrPPUEnd: // PPU
		memory.ppu[(addr-addrPPUStart)%(addrPPUMirrorStart-addrPPUStart)] = value
	case addr >= addrAPUIOStart && addr <= addrAPUIOEnd: // APU I/O
		memory.io[(addr - addrAPUIOStart)] = value
	case addr >= addrControllerIOStart && addr <= addrControllerIOEnd:
		memory.io[(addr - addrAPUIOStart)] = value
	case addr >= addrPRGROMStart && addr <= addrPRGROMEnd:
		return fmt.Errorf("Memory: PRGROM is not allowed write. addr: %b", addr)
	default:
		return fmt.Errorf("Memory: invalid addr is specified. addr: %b", addr)
	}
	return nil
}
