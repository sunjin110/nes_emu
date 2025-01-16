package memory

import (
	"fmt"
)

type Memory interface {
	Read(addr uint16) (byte, error)
	Write(addr uint16, value byte) error
}

func NewMemory() Memory {
	return &memory{}
}

// PPUのメモリ構成を考える
type memory struct {
	patternTable0   patternTable   // 0x0000-0x0fff
	patternTable1   patternTable   // 01000-0x1fff
	nametable0      nametable      // 02000-0x23bf
	attributeTable0 attributeTable // 0x23c0-0x23ff
	nametable1      nametable      // 0x2400-0x27bf
	attributeTable1 attributeTable // 0x27c0-0x27ff
	nametable2      nametable      // 0x2800-0x2bbf
	attributeTable2 attributeTable // 0x2bc0-0x2bff
	nametable3      nametable      // 0x2c00-0x2fbf
	attributeTable3 attributeTable // 0x2fc0-0x2fff
	// 0x3000 - 0x3eff mirror of 0x2000-0x2eff
	backgroundPallet backgroundPallet // 0x3f00-0x3f0f
	splitePallet     splitePallet     // 0x3f10-0x3f1f
	// 0x3f20-0x3fff mirror of 0x3f00-0x3f1f
}

func (m *memory) Read(addr uint16) (byte, error) {
	switch {
	case addr <= 0x0fff:
		// 0x0000-0x0fff : patternTable0
		return m.patternTable0.data[addr], nil

	case addr <= 0x1fff:
		// 0x1000-0x1fff : patternTable1
		return m.patternTable1.data[addr-0x1000], nil

	case addr <= 0x23bf:
		// 0x2000-0x23bf : nametable0
		return m.nametable0.data[addr-0x2000], nil

	case addr <= 0x23ff:
		// 0x23c0-0x23ff : attributeTable0
		return m.attributeTable0.data[addr-0x23c0], nil

	case addr <= 0x27bf:
		// 0x2400-0x27bf : nametable1
		return m.nametable1.data[addr-0x2400], nil

	case addr <= 0x27ff:
		// 0x27c0-0x27ff : attributeTable1
		return m.attributeTable1.data[addr-0x27c0], nil

	case addr <= 0x2bbf:
		// 0x2800-0x2bbf : nametable2
		return m.nametable2.data[addr-0x2800], nil

	case addr <= 0x2bff:
		// 0x2bc0-0x2bff : attributeTable2
		return m.attributeTable2.data[addr-0x2bc0], nil

	case addr <= 0x2fbf:
		// 0x2c00-0x2fbf : nametable3
		return m.nametable3.data[addr-0x2c00], nil

	case addr <= 0x2fff:
		// 0x2fc0-0x2fff : attributeTable3
		return m.attributeTable3.data[addr-0x2fc0], nil

	case addr <= 0x3eff:
		// 0x3000-0x3eff : mirror of 0x2000-0x2eff
		// 0x3000 ～ 0x3eff は 0x2000 ～ 0x2eff のミラー領域なので
		// 0x1000 引いたアドレスでもう一度 Read() を呼ぶ
		return m.Read(addr - 0x1000)

	case addr <= 0x3f0f:
		// 0x3f00-0x3f0f : backgroundPallet
		return m.backgroundPallet.data[addr-0x3f00], nil

	case addr <= 0x3f1f:
		// 0x3f10-0x3f1f : splitePallet
		return m.splitePallet.data[addr-0x3f10], nil

	case addr <= 0x3fff:
		// 0x3f20-0x3fff : mirror of 0x3f00-0x3f1f
		// 下位 5bit (0x1f) をマスクして 0x3f00 に加算すれば実アドレスが得られる
		return m.Read(0x3f00 + (addr & 0x1f))
	default:
		return 0, fmt.Errorf("invalid addr: %x", addr)
	}
}

func (m *memory) Write(addr uint16, value byte) error {
	switch {
	case addr <= 0x0fff:
		// 0x0000-0x0fff: patternTable0
		m.patternTable0.data[addr] = value
		return nil

	case addr <= 0x1fff:
		// 0x1000-0x1fff: patternTable1
		m.patternTable1.data[addr-0x1000] = value
		return nil

	case addr <= 0x23bf:
		// 0x2000-0x23bf: nametable0
		m.nametable0.data[addr-0x2000] = value
		return nil

	case addr <= 0x23ff:
		// 0x23c0-0x23ff: attributeTable0
		m.attributeTable0.data[addr-0x23c0] = value
		return nil

	case addr <= 0x27bf:
		// 0x2400-0x27bf: nametable1
		m.nametable1.data[addr-0x2400] = value
		return nil

	case addr <= 0x27ff:
		// 0x27c0-0x27ff: attributeTable1
		m.attributeTable1.data[addr-0x27c0] = value
		return nil

	case addr <= 0x2bbf:
		// 0x2800-0x2bbf: nametable2
		m.nametable2.data[addr-0x2800] = value
		return nil

	case addr <= 0x2bff:
		// 0x2bc0-0x2bff: attributeTable2
		m.attributeTable2.data[addr-0x2bc0] = value
		return nil

	case addr <= 0x2fbf:
		// 0x2c00-0x2fbf: nametable3
		m.nametable3.data[addr-0x2c00] = value
		return nil

	case addr <= 0x2fff:
		// 0x2fc0-0x2fff: attributeTable3
		m.attributeTable3.data[addr-0x2fc0] = value
		return nil

	case addr <= 0x3eff:
		// 0x3000-0x3eff: mirror of 0x2000-0x2eff
		// ミラー先に書き込む
		return m.Write(addr-0x1000, value)

	case addr <= 0x3f0f:
		// 0x3f00-0x3f0f: backgroundPallet
		m.backgroundPallet.data[addr-0x3f00] = value
		return nil

	case addr <= 0x3f1f:
		// 0x3f10-0x3f1f: splitePallet
		m.splitePallet.data[addr-0x3f10] = value
		return nil

	case addr <= 0x3fff:
		// 0x3f20-0x3fff: mirror of 0x3f00-0x3f1f
		// 下位5ビットをマスクして 0x3f00 に加算 → そこへ書き込む
		return m.Write(0x3f00+(addr&0x1f), value)

	default:
		return fmt.Errorf("invalid addr: %x", addr)
	}
}

type patternTable struct {
	data [0x0fff + 1]byte
}

type nametable struct {
	data [0x03bf + 1]byte
}

type attributeTable struct {
	data [64]byte
}

type backgroundPallet struct {
	data [0x10]byte
}

type splitePallet struct {
	data [0x10]byte
}
