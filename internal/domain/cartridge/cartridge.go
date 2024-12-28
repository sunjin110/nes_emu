package cartridge

import (
	"errors"
	"fmt"

	"github.com/sunjin110/nes_emu/pkg/logger"
)

const (
	iNESHeaderSize = 16
	prgBankSize    = 16 * 1024 // 16KB
	chrBankSize    = 8 * 1024  // 8KB
)

type Cartridge struct {
	PRG []byte
	CHR []byte
	// mapperの種類 https://www.nesdev.org/wiki/Mapper
	// 現在はiNES1.0のみ対応
	MapperNo     int
	PRGBankCount int
	CHRBankCount int
}

func NewCartridge(data []byte) (*Cartridge, error) {
	if string(data[0:3]) != "NES" {
		return nil, errors.New("invalid header 0~3")
	}
	if data[3] != 0x1A {
		return nil, errors.New("invalid header 3")
	}

	// PRG-ROM
	prgBankCount := int(data[4])
	prgStart := iNESHeaderSize
	prgEnd := prgStart + prgBankCount*prgBankSize

	if prgEnd > len(data) {
		return nil, errors.New("ROM size is too short to contain full RPG-ROM")
	}
	prgData := data[prgStart:prgEnd]

	// CHR-ROM
	chrBankCount := int(data[5])
	chrStart := prgEnd
	chrEnd := chrStart + chrBankCount*chrBankSize
	if chrEnd > len(data) {
		return nil, errors.New("ROM size is too chort to contain full CHR-ROM")
	}
	chrData := data[chrStart:chrEnd]

	// 例: header[6] = 0x31 (下位4ビット=1, 上位4ビット=3)
	//    header[7] = 0x80 (上位4ビット=8, 下位4ビット=0)
	//    → mapperNo = 3 | 0x80 = 0x83 = 131
	mapperNo := (data[6] >> 4) | (data[7] & 0xF0)

	return &Cartridge{
		PRG:          prgData,
		CHR:          chrData,
		MapperNo:     int(mapperNo),
		PRGBankCount: prgBankCount,
		CHRBankCount: chrBankCount,
	}, nil
}

func (cartridge *Cartridge) ReadPRG(offset, size int) ([]byte, error) {
	if offset+size <= len(cartridge.PRG) {
		return cartridge.PRG[offset:size], nil
	}

	// 実際のデータより大きい場合はミラーリングする
	data := make([]byte, size)
	for i := 0; i < size; i++ {
		idx := offset + i
		if idx >= len(cartridge.PRG) {
			idx = idx % len(cartridge.PRG)
		}
		data[i] = cartridge.PRG[idx]
	}
	return data, nil
}

func (cartridge *Cartridge) WritePRG(data []byte, offset int) error {
	if offset+len(data) > len(cartridge.PRG) {
		return fmt.Errorf("offset+size is too large to write PRG. offset: %d, size: %d, prgSize: %d", offset, len(data), len(cartridge.PRG))
	}
	copy(cartridge.PRG[offset:], data)
	return nil
}

func (cartridge *Cartridge) ReadCHR(offset, size int) ([]byte, error) {
	if offset+size <= len(cartridge.CHR) {
		return cartridge.CHR[offset:size], nil
	}

	// 実際のデータより大きい場合はミラーリングする
	data := make([]byte, size)
	for i := 0; i < size; i++ {
		idx := offset + i
		if idx >= len(cartridge.CHR) {
			logger.Logger.Warn("ReadCHR: didn't expect mirroring")
			idx = idx % len(cartridge.CHR)
		}
		data[i] = cartridge.CHR[idx]
	}
	return data, nil
}
