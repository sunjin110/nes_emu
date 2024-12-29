package memory

import (
	"fmt"

	"github.com/sunjin110/nes_emu/internal/domain/apu"
	"github.com/sunjin110/nes_emu/internal/domain/controller"
	"github.com/sunjin110/nes_emu/internal/domain/ppu"
	"github.com/sunjin110/nes_emu/internal/domain/prgrom"
	"github.com/sunjin110/nes_emu/internal/domain/ram"
	"github.com/sunjin110/nes_emu/pkg/logger"
)

type Memory struct {
	ram        ram.RAM               // RAM:ワーキングメモリ(0x0000-0x07ff) 0x0800-0x1fffはミラー
	ppu        ppu.PPU               // PPUレジスタ(0x2000〜0x2007)　0x2008-0x3fffはミラー
	apu        apu.APU               // APU(0x4000-0x4015)
	controller controller.Controller // Controller(0x4016-4017)
	prgROM     prgrom.PRGROM         // PRG-ROM(0x8000〜0xFFFF)
}

func NewMemory(prgROM prgrom.PRGROM) *Memory {
	return &Memory{
		ram:        *ram.NewRAM(),
		ppu:        *ppu.NewPPU(),
		apu:        *apu.NewAPU(),
		controller: *controller.NewController(),
		prgROM:     prgROM,
	}
}

func (memory *Memory) Read(addr uint16) (byte, error) {
	switch {
	case ram.IsRAMRange(addr): // RAM
		return memory.ram.Read(addr), nil
	case ppu.IsPPUAddrRange(addr): // PPU
		return memory.ppu.Read(addr), nil
	case apu.IsAPUAddrRange(addr): // APU
		return memory.apu.Read(addr), nil
	case controller.IsControllerAddr(addr):
		return memory.controller.Read(addr), nil
	case prgrom.IsPRGRomRange(addr):
		return memory.prgROM.Read(addr), nil
	default:
		logger.Logger.Error("invalid addr is specified", "addr", addr)
		return 0, fmt.Errorf("Memory: invalid addr is specified. addr: %b", addr)
	}
}

func (memory *Memory) Write(addr uint16, value byte) error {
	switch {
	case ram.IsRAMRange(addr): // RAM
		memory.ram.Write(addr, value)
	case ppu.IsPPUAddrRange(addr): // PPU
		memory.ppu.Write(addr, value)
	case apu.IsAPUAddrRange(addr): // APU
		memory.apu.Write(addr, value)
	case controller.IsControllerAddr(addr):
		memory.controller.Write(addr, value)
	case prgrom.IsPRGRomRange(addr):
		return fmt.Errorf("Memory: PRGROM is not allowed write. addr: %b", addr)
	default:
		return fmt.Errorf("Memory: invalid addr is specified. addr: %b", addr)
	}
	return nil
}
