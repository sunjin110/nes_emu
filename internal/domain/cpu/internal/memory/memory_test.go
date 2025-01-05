package memory_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/sunjin110/nes_emu/internal/domain/cpu/internal/memory"
	"github.com/sunjin110/nes_emu/internal/domain/prgrom"
)

func TestMemory_ReadWrite(t *testing.T) {
	// メモリの初期化
	prgRom := [prgrom.PRGROMSize]byte{}
	prgRom[0] = 0x99
	mem := memory.NewMemory(prgrom.NewFixedPRGROM(prgRom))

	// RAMの書き込みと読み込み
	addrRAM := uint16(0x0000)
	mem.Write(addrRAM, 0x42)
	value, err := mem.Read(addrRAM)
	assert.NoError(t, err)
	assert.Equal(t, byte(0x42), value, "RAM: 値が正しく読み込めません")

	// RAMのミラーリング
	addrRAMMirror := uint16(0x0800)
	value, err = mem.Read(addrRAMMirror)
	assert.NoError(t, err)
	assert.Equal(t, byte(0x42), value, "RAMミラーリング: 値が正しく反映されていません")

	// PPUレジスタの書き込みと読み込み
	addrPPU := uint16(0x2000)
	mem.Write(addrPPU, 0x84)
	value, err = mem.Read(addrPPU)
	assert.NoError(t, err)
	assert.Equal(t, byte(0x84), value, "PPUレジスタ: 値が正しく読み込めません")

	// PPUレジスタのミラーリング
	addrPPUMirror := uint16(0x2008)
	value, err = mem.Read(addrPPUMirror)
	assert.NoError(t, err)
	assert.Equal(t, byte(0x84), value, "PPUミラーリング: 値が正しく反映されていません")

	// IOレジスタの書き込みと読み込み
	addrIO := uint16(0x4000)
	mem.Write(addrIO, 0xAA)
	value, err = mem.Read(addrIO)
	assert.NoError(t, err)
	assert.Equal(t, byte(0xAA), value, "IOレジスタ: 値が正しく読み込めません")

	// PRG-ROMの読み込み確認
	addrPRGROM := uint16(0x8000)
	value, err = mem.Read(addrPRGROM)
	assert.NoError(t, err)
	assert.Equal(t, byte(0x99), value, "PRG-ROM: 値が正しく読み込めません")

	// PRG-ROMの書き込み禁止確認
	err = mem.Write(addrPRGROM, 0x55)
	assert.Error(t, err, "PRG-ROM: 書き込み禁止が正しく動作していません")
}

func TestMemory_InvalidAddress(t *testing.T) {
	// メモリの初期化
	mem := memory.NewMemory(prgrom.NewFixedPRGROM([32 * 1024]byte{}))

	// 無効なアドレスの読み込み
	invalidAddr := uint16(0x8000 - 1)
	_, err := mem.Read(invalidAddr)
	assert.Error(t, err, "無効なアドレス: エラーが発生しませんでした")

	// 無効なアドレスの書き込み
	err = mem.Write(invalidAddr, 0x55)
	assert.Error(t, err, "無効なアドレス: エラーが発生しませんでした")
}
