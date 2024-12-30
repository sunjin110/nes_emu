package prgrom

import "fmt"

const (
	SwitchablePRGROMBankSize = 16 * 1024 // 16KB
)

type SwitchablePRGROM struct {
	banks          [][SwitchablePRGROMBankSize]byte // 16KBごとのバンクを配列として管理
	fixedBank      int                              // 固定バンク（0xC000〜0xFFFFに常にマッピングされる）
	switchableBank int                              // 切り替え可能バンク(0x8000〜0xBFFFにマッピングされる)
}

func NewSwitchablePRGROM(banks [][SwitchablePRGROMBankSize]byte, fixedBank int, switchableBank int) PRGROM {
	return &SwitchablePRGROM{
		banks:          banks,
		fixedBank:      fixedBank,
		switchableBank: switchableBank,
	}
}

func (rom *SwitchablePRGROM) Read(addr uint16) byte {
	if addr >= 0x8000 && addr < 0xC000 {
		// 切り替え可能バンク（0x8000〜0xBFFF）
		return rom.banks[rom.switchableBank][addr-0x8000]
	} else if addr >= 0xC000 && addr <= 0xFFFF {
		// 固定バンク（0xC000〜0xFFFF）
		return rom.banks[rom.fixedBank][addr-0xC000]
	}
	return 0xFF // 無効アドレスの場合
}

func (rom *SwitchablePRGROM) InitPC() uint16 {
	// https://www.pagetable.com/?p=410
	pcLower := rom.banks[rom.fixedBank][0xFFFC-0xC000]
	pcUpper := rom.banks[rom.fixedBank][0xFFFD-0xC000]
	pc := uint16(pcLower) | (uint16(pcUpper) << 8)
	return pc
}

func (rom *SwitchablePRGROM) SwitchBank(bank int) error {
	if bank >= len(rom.banks) {
		return fmt.Errorf("SwitchablePRGROM: invalid bank no. bank: %d, bank.len: %d", bank, len(rom.banks))
	}
	rom.switchableBank = bank
	return nil
}
