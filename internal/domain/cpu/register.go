package cpu

import (
	"github.com/sunjin110/nes_emu/internal/domain/prgrom"
)

type Register struct {
	a  byte   // Accumulator 命令の計算結果の格納
	x  byte   // 特定のアドレッシングモード (後述) でインデックスとして使われます。 INX 命令と組み合わせてループのカウンタとしても使われる様子？
	y  byte   // Xと同様
	pc uint16 // Program Counter // CPUが次に実行すべき命令のアドレスを保持する
	sp uint16 // Stack Pointer // スタックの先頭のアドレスを保持します
	// P Processor Status // ステータスレジスタ。各ビットが意味を持つ、
	//  file:///Users/sunjin/Downloads/LayerWalker.pdf
	p byte
}

func NewRegister(prgROM prgrom.PRGROM) (*Register, error) {

	return &Register{
		a:  0, // TODO 初期化
		x:  0, // TODO 初期化
		y:  0, // TODO 初期化
		pc: prgROM.InitPC(),
		sp: 0xFD, // https://www.pagetable.com/?p=410
		p:  0,    // TODO 初期化
	}, nil
}

// TODO ステータスレジスタの更新と読み取り
// file:///Users/sunjin/Downloads/LayerWalker.pdf
