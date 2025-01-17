package ppu

import (
	"fmt"

	"github.com/sunjin110/nes_emu/internal/domain/ppu/internal/memory"
	"github.com/sunjin110/nes_emu/internal/domain/ppu/internal/register"
)

type PPU struct {
	internalRegister register.Register
	memory           memory.Memory
}

func NewPPU() *PPU {
	return &PPU{}
}

const (
	addrPPUStart       = 0x2000
	addrPPUMirrorStart = 0x2008
	addrPPUEnd         = 0x3FFF

	// PPUCTRL これを書き込むことでPPUのさまざまな挙動を設定できるレジスタ
	ppuCTRL = 0x2000

	// PPUMASK スプライトやBGの行が、色効果の設定を制御するレジスタ
	ppuMask = 0x2001

	// PPUSTATUS は各ビットがPPUの状態を表すステータスレジスタになっている
	ppuStatus = 0x2002

	// OAMADDR Object Attribute Memoryに書き込む時のアドレスを指定するためのレジスタ、8bitの値を書き込んで、読み書きしたいOAMのアドレスを指定する
	oamAddr = 0x2003

	// OAMDATA を読み書きすることで、OAMADDRで指定したアドレスのOAMの値を読み書きする、OAMDATAに書き込むたびにOAMADDRの値がインクリメントされる
	oamData = 0x2004

	// PPUSCROLL BGスクロールの値を設定するために使用される、X方向スクロール -> Y方向スクロールの順に書き込みを行い、PPUCTRLで指定されたNametableのうちどのピクセルを左上(0x 0)に描画するかを指定する
	ppuScroll = 0x2005

	// PPUADDR CPUメモリ空間とPPUメモリ空間は直接読み書きができないので、このPPUADDRレジスタと次のPPUDATAレジスタを用意てCPUからPPUメモリ空間を読み書きする
	ppuAddr = 0x2006

	// PPUDATA PPUADDRで指定したメモリ空間のアドレスを読み書きする
	ppuData = 0x2007

	// OAMDMA レジスタに8bitの値を書き込むことでCPUメモリ空間のページ(256bite)を全てOAMに転送できる(Direct Memory Access)
	// OAMDMAレジスタに$XXを書き込んだら CPUメモリ空間の$XX00-$XXFFの1ページ分がOAMにDMA転送される
	oamDMA = 0x4014
)

// Read CPUがreadする
func (p *PPU) Read(addr uint16) (byte, error) {
	switch addr {
	case ppuStatus:
		value, err := p.readPPUStatus()
		if err != nil {
			return 0, fmt.Errorf("failed readPPUStatus. err: %w", err)
		}
		return value, nil
	case ppuData:
		value, err := p.readPPUData()
		if err != nil {
			return 0, fmt.Errorf("failed readPPUData. err: %w", err)
		}
		return value, nil
	}

	return 0, fmt.Errorf("undefined ppu register read. addr: %x", addr)
}

// Write CPUがwriteする
func (p *PPU) Write(addr uint16, value byte) error {
	switch addr {
	case ppuCTRL:
		if err := p.writePPUCtrl(value); err != nil {
			return fmt.Errorf("failed writePPUCtrl. err: %w", err)
		}
	case ppuMask:
		if err := p.writePPUMask(value); err != nil {
			return fmt.Errorf("failed writePPUMask. err: %w", err)
		}
	case ppuStatus:
		// 読み込み専用
		return fmt.Errorf("PPU: Write: 0x2002は読み込み専用. addr: %x, value: %x", addr, value)
	case oamAddr:
		if err := p.writeOAMADDR(value); err != nil {
			return fmt.Errorf("failed writeOAMADDR. err: %w", err)
		}
	case oamData:
		if err := p.writeOAMData(value); err != nil {
			return fmt.Errorf("failed writeOAMData. err: %w", err)
		}
	case ppuScroll:
		if err := p.writePPUScroll(value); err != nil {
			return fmt.Errorf("failed writePPUScroll. err: %w", err)
		}
	case ppuAddr:
		if err := p.writePPUADDR(value); err != nil {
			return fmt.Errorf("failed writePPUADDR. err: %w", err)
		}
	case ppuData:
		if err := p.writePPUData(value); err != nil {
			return fmt.Errorf("failed writePPUData. err: %w", err)
		}
	}
	return fmt.Errorf("undefined ppu register addr: %x. value: %x", addr, value)
}

func (p *PPU) writePPUCtrl(value byte) error {
	panic("todo")
}

func (p *PPU) writePPUMask(value byte) error {
	panic("todo")
}

func (p *PPU) writeOAMADDR(value byte) error {
	panic("todo")
}

func (p *PPU) writeOAMData(value byte) error {
	panic("todo")
}

func (p *PPU) writePPUScroll(value byte) error {
	panic("todo")
}

func (p *PPU) writePPUADDR(value byte) error {
	panic("todo")
}

func (p *PPU) writePPUData(value byte) error {
	panic("todo")
}

func (p *PPU) readPPUStatus() (byte, error) {
	panic("todo")
}

func (p *PPU) readPPUData() (byte, error) {
	panic("todo")
}

func IsPPUAddrRange(addr uint16) bool {
	// TODO OAM_DMAの場合どうなるかを確認すること
	return addr >= addrPPUStart && addr <= addrPPUEnd
}
