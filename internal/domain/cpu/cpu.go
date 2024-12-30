package cpu

import (
	"fmt"

	"github.com/sunjin110/nes_emu/internal/domain/memory"
	"github.com/sunjin110/nes_emu/internal/domain/prgrom"
)

const memorySize = 16 * 1024 // 16KB

// CPU document: https://www.nesdev.org/wiki/CPU
type CPU struct {
	memory   memory.Memory
	register Register
}

func NewCPU(memory memory.Memory, prgROM prgrom.PRGROM) (*CPU, error) {
	register, err := NewRegister(prgROM)
	if err != nil {
		return nil, fmt.Errorf("failed new register. err: %w", err)
	}

	return &CPU{
		memory:   memory,
		register: *register,
	}, nil
}

// Run CPUの1サイクルの実行
// clockCount: PPUやAPUとの同期のため、実行時間にかかった実行クロック数を返す
func (cpu *CPU) Run() (clockCount uint8, err error) {
	opcode, err := cpu.fetchOpcode()
	if err != nil {
		return 0, fmt.Errorf("failed fetchOpcode. err: %w", err)
	}
	// TODO fetchOperand

	// TODO: opcodeのmnemonicの挙動を実装する
	switch opcode.Mnemonic {
	case ADC:
		cpu.adc(0) // TODO
	}

	return
}

func (cpu *CPU) Interrupt() error {
	// 割り込み処理
	return nil
}

func (cpu *CPU) Reset() error {
	r, err := NewRegister(cpu.memory.GetPRGROM())
	if err != nil {
		return fmt.Errorf("CPU: failed reset. err: %w", err)
	}
	cpu.register = *r
	return nil
}

// fetchOpcode PCから実行コードを取得する
func (cpu *CPU) fetchOpcode() (Opcode, error) {
	pc := cpu.register.pc
	opcodeByte, err := cpu.memory.Read(pc)
	if err != nil {
		return Opcode{}, fmt.Errorf("CPU: fialed read memory. addr: %x, err: %w", pc, err)
	}
	opcode, ok := Opcodes[opcodeByte]
	if !ok {
		return Opcode{}, fmt.Errorf("CPU: undefined opcode. opcodeByte: %x", opcodeByte)
	}
	return opcode, nil
}

func (cpu *CPU) fetchOperand(addressingMode AddressingMode) byte {
	// fetchOperandの実装イメージ
	// switch addressingMode {
	// case "Immediate":
	// 	// 即値: PCが指すアドレスの次のバイト
	// 	operand := cpu.Memory.Read(cpu.PC)
	// 	cpu.PC++ // PCを進める
	// 	return operand

	// case "Absolute":
	// 	// 絶対アドレッシング: 次の2バイトがアドレスを示す
	// 	low := cpu.Memory.Read(cpu.PC)
	// 	high := cpu.Memory.Read(cpu.PC + 1)
	// 	address := uint16(high)<<8 | uint16(low)
	// 	cpu.PC += 2 // PCを2バイト進める
	// 	return cpu.Memory.Read(address)

	// // 他のアドレッシングモードも実装
	// default:
	// 	panic("Unknown addressing mode")
	// }
	panic("todo")
}

// 加算処理
func (cpu *CPU) adc(operand byte) {
	// 8bit以上の計算ができるように
	var carry byte
	if cpu.getFlag(carryFlag) {
		carry = 1
	}

	// 16ビットで計算してキャリーを考慮
	result := uint16(cpu.register.a) + uint16(operand) + uint16(carry)

	// キャリーフラグの更新
	cpu.setFlag(carryFlag, result > 0xFF)

	// 結果を8ビットに収める
	cpu.register.a = byte(result & 0xFF)

	// ゼロフラグ
	cpu.setFlag(zeroFlag, cpu.register.a == 0)

	// ネガティブフラグの更新
	cpu.setFlag(negativeFlag, cpu.register.a&0x80 != 0)
}

func (cpu *CPU) setFlag(flag statusFlag, value bool) {
	if value {
		cpu.register.p |= flag.toByte() // OR
	} else {
		cpu.register.p &^= flag.toByte() // AND NOT
	}
}

func (cpu *CPU) getFlag(flag statusFlag) bool {
	return cpu.register.p&flag.toByte() != 0
}

func (cpu *CPU) incrementPC(count uint16) {
	cpu.register.pc += count
}

/**
# cpu実行順序
## fetch
- プログラムカウンタ(PC)が指している場所のROMから命令を読み込む。
命令によっては引数(オペランド)があることもあり、その場合はオペランドも読み込む

- この時、次のfetchのために次の命令を指すようにプログラムカウンタ(PCの値を更新する)

## decode
- ROMから読み込んだ命令の内容を解釈する

## execute
- 命令ごとに決められた演算を行う。これにより、レジスタの値やRAMに保存されている値が更新される。
*/

/**
CPU命令割り込み
RESET: 起動時とリセットボタンが押された時
NMI: ハードウェア割り込み。PPUが描画完了したことをCPUに知らせる時に使用
IRQ: APUのフレームシーケンサが発生させ利割り込み
BRK: BRK命令を実行した時に発生するもの
*/

/**
アドレッシングモード
命令を実行する時に引数を指定する方法が13種類ある、それをアドレッシングモードと呼ぶ
document: https://www.nesdev.org/wiki/CPU_addressing_modes

- Implied: 引数なし
- Accumulator: Aレジスタを利用する
- Immediate: 定数を指定する(8bitまで)
- Zeropage: アドレスを指定
- Zeropage, X: 配列などを渡せる
- Zeropage, Y: Zeropage, Xと同じ、レジスタの利用する箇所が違う
- Relative: 分岐命令で利用されるやつ
- Absolute: 変数に入っているものを指定するやつ
- Absolute, X: Zeropage, Xと同じだが変数を指定
- Absolute, Y: Absolute, Xと同じ、レジスタの利用する箇所が違う
- (Indirect): 引数として16bitの値を利用する、括弧は参照はずしなのでIM16をアドレスとしてみた時のIM16番地にあるアタういを表す
- (Indirect, X): 8bitのアドレス(IM8 + X)に格納されているアドレスを操作対象とする -> 配列のN番目を取得する的な
- (Indirect), Y: 上記とは違う
*/

/**
(Indirect, X)と(Indirect), Yの違い

特徴	(Indirect, X)	(Indirect), Y
参照の段階数	1段階（ゼロページから直接間接アドレスを取得）	2段階（ゼロページ参照 + オフセット加算）
操作順序	即値 + X → ゼロページ参照	即値 → ゼロページ参照 → + Y
主な用途	ポインタテーブル（複数ポインタの中の1つを選ぶ）	配列やデータブロック（ベースアドレスにオフセットを加える）
例えるなら	「テーブルから直接1つの値を取る」	「ベースアドレスにオフセットを足して次の値を取る」

# (Indirect), Yが必要な理由
base addressを相対的に決めてデータアクセスができるため便利、例えばスプライトデータが$4000からだとして、そのデータを取得する場合
IndirectXだけの場合は常に$4000を意識しながらデータをアクセスする必要がある
*/
