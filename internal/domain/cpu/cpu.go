package cpu

import (
	"errors"
	"fmt"

	"github.com/sunjin110/nes_emu/internal/domain/memory"
	"github.com/sunjin110/nes_emu/internal/domain/prgrom"
	"github.com/sunjin110/nes_emu/pkg/bit_helper"
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
func (cpu *CPU) Run() (cycles uint8, err error) {
	opcode, err := cpu.fetchOpcode()
	if err != nil {
		return 0, fmt.Errorf("failed fetchOpcode. err: %w", err)
	}
	switch opcode.Mnemonic {
	case ADC:
		cycles, err := cpu.adc(opcode)
		if err != nil {
			return 0, fmt.Errorf("failed cpu.adc. opcode: %+v, err: %w", opcode, err)
		}
		return cycles, nil
	case LDA:
		// todo

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

// fetchArg 引数を取得する
// additionalCycle: IndirectY, Relative, AbsoluteX, AbsoluteYで追加cycleが発生する可能性があるため
func (cpu *CPU) fetchArg(mode AddressingMode) (value byte, additionalCycle uint8, err error) {
	switch mode {
	case Implied:
		return 0, 0, errors.New("featchArg: Implied dose not have an arg")
	case Accumulator:
		return cpu.register.a, 0, nil
	case Immediate:
		value, err = cpu.memory.Read(cpu.register.pc + 1)
		if err != nil {
			return 0, 0, fmt.Errorf("fetchArg: Immediate failed get memory value. addr: %x, err: %w", cpu.register.pc+1, err)
		}
		return value, 0, nil
	default:
		addr, additionalCycle, err := cpu.fetchAddr(mode)
		if err != nil {
			return 0, 0, fmt.Errorf("fetchArg: failed fetchAddr. mode: %d, err: %w", mode, err)
		}
		value, err = cpu.memory.Read(addr)
		if err != nil {
			return 0, 0, fmt.Errorf("fetchArg: failed get memory value. mode: %d, addr: %x, err: %w", mode, addr, err)
		}
		return value, additionalCycle, nil
	}
}

// fetchArg アドレスを取得する
// additionalCycle: IndirectY, Relative, AbsoluteX, AbsoluteYで追加cycleが発生する可能性があるため
func (cpu *CPU) fetchAddr(mode AddressingMode) (addr uint16, additionalCycle uint8, err error) {

	switch mode {
	case Absolute:
		lower, err := cpu.memory.Read(cpu.register.pc + 1)
		if err != nil {
			return 0, 0, fmt.Errorf("fetchAddr: Absolute: failed read lower part. addr: %x, err: %w", cpu.register.pc+1, err)
		}
		upper, err := cpu.memory.Read(cpu.register.pc + 2)
		if err != nil {
			return 0, 0, fmt.Errorf("fetchAddr: Absolute: failed read upper part. addr: %x, err: %w", cpu.register.pc+2, err)
		}
		addr = bit_helper.BytesToUint16(lower, upper)
		return addr, 0, nil
	case Zeropage:
		addrUint8, err := cpu.memory.Read(cpu.register.pc + 1)
		if err != nil {
			return 0, 0, fmt.Errorf("fetchAddr: Zeropage: failed read addr: %x, err: %w", cpu.register.pc+1, err)
		}
		return uint16(addrUint8), 0, nil
	case ZeropageX:
		addrUint8, err := cpu.memory.Read(cpu.register.pc + 1)
		if err != nil {
			return 0, 0, fmt.Errorf("fetchAddr: ZeropageX: failed read addr: %x, err: %w", cpu.register.pc+1, err)
		}
		addrUint8 += cpu.register.x
		return uint16(addrUint8), 0, nil
	case ZeropageY:
		addrUint8, err := cpu.memory.Read(cpu.register.pc + 1)
		if err != nil {
			return 0, 0, fmt.Errorf("fetchAddr: ZeropageY: failed read addr: %x, err: %w", cpu.register.pc+1, err)
		}
		addrUint8 += cpu.register.y
		return uint16(addrUint8), 0, nil
	case AbsoluteX:
		lower, err := cpu.memory.Read(cpu.register.pc + 1)
		if err != nil {
			return 0, 0, fmt.Errorf("fetchAddr: AbsoluteX: failed read lower. addr: %x, err: %w", cpu.register.pc+1, err)
		}

		upper, err := cpu.memory.Read(cpu.register.pc + 2)
		if err != nil {
			return 0, 0, fmt.Errorf("fetchAddr: AbsoluteX: failed read upper. addr: %x, err: %w", cpu.register.pc+2, err)
		}

		addr := bit_helper.BytesToUint16(lower, upper)
		beforeAddr := addr
		addr += uint16(cpu.register.x)

		// ページ境界クロスチェック
		// ページ境界を跨いだ場合は、cycle数を+1する
		if (beforeAddr & 0xFF00) != (addr & 0xFF00) {
			additionalCycle = 1
		}
		return addr, additionalCycle, nil
	case AbsoluteY:
		lower, err := cpu.memory.Read(cpu.register.pc + 1)
		if err != nil {
			return 0, 0, fmt.Errorf("fetchAddr: AbsoluteY: failed read lower. addr: %x, err: %w", cpu.register.pc+1, err)
		}

		upper, err := cpu.memory.Read(cpu.register.pc + 2)
		if err != nil {
			return 0, 0, fmt.Errorf("fetchAddr: AbsoluteY: failed read upper. addr: %x, err: %w", cpu.register.pc+2, err)
		}

		addr := bit_helper.BytesToUint16(lower, upper)
		beforeAddr := addr
		addr += uint16(cpu.register.y)

		// ページ境界クロスチェック
		// ページ境界を跨いだ場合は、cycle数を+1する
		if (beforeAddr & 0xFF00) != (addr & 0xFF00) {
			additionalCycle = 1
		}
		return addr, additionalCycle, nil
	case Relative:
		// if文後のジャンプ先のPCを計算するときに利用する
		offset, err := cpu.memory.Read(cpu.register.pc + 1)
		if err != nil {
			return 0, 0, fmt.Errorf("featchAddr: Relative: failed read offset. addr: %x, err: %w", cpu.register.pc+1, err)
		}

		// 符号付き解釈になるように先にint8にする
		signedOffset := int8(offset)

		// 次に実行される予定のPCを符号付きで取得する
		signedPC := int32(cpu.register.pc) + 2

		// ジャンプ先のPCを計算する
		signedAddr := signedPC + int32(signedOffset)
		if signedAddr < 0 || signedAddr > 0xFFFF {
			return 0, 0, fmt.Errorf("fetchAddr: Relative: Invalid addr. sighendAddr: %x, signedPC: %x, signedOffset: %x", signedAddr, signedPC, signedOffset)
		}

		addr := uint16(signedAddr)

		// ページ境界クロスチェック
		// ページ境界を跨いだ場合は、cycle数を+1する
		// ページクロスで +1 クロック、Relative はブランチ命令で使われるが、ブランチ成立時にはさらに +1 されることに注意する
		// ここでページ境界のクロスをチェックするのに signedPC を使用している理由は、相対アドレッシングにおけるページ境界の比較基準が 次に実行される予定のPC（PC + 2） だからです。
		if (uint16(signedPC) & 0xFF00) != (addr & 0xFF00) {
			additionalCycle = 1
		}
		return addr, additionalCycle, nil
	case IndirectX:
		// *(lower + X)
		indirectLower, err := cpu.memory.Read(cpu.register.pc + 1)
		if err != nil {
			return 0, 0, fmt.Errorf("featchAddr: IndirectX: failed read offset. addr: %x, err: %w", cpu.register.pc+1, err)
		}

		lowerAddr := uint16(indirectLower) + uint16(cpu.register.x)
		upperAddr := uint16(lowerAddr + 1)

		lower, err := cpu.memory.Read(lowerAddr)
		if err != nil {
			return 0, 0, fmt.Errorf("fetchAddr: IndirectX: failed read lower. addr: %x, err: %w", lowerAddr, err)
		}
		upper, err := cpu.memory.Read(upperAddr)
		if err != nil {
			return 0, 0, fmt.Errorf("fetchAddr: IndirectX: failed read upper. addr: %x, err: %w", upperAddr, err)
		}

		addr := bit_helper.BytesToUint16(lower, upper)
		return addr, 0, nil
	case IndirectY:
		// *(lower) + Y
		lowerAddr, err := cpu.memory.Read(cpu.register.pc + 1)
		if err != nil {
			return 0, 0, fmt.Errorf("fetchAddr: IndirectY: failed read lowerAddr. addr: %x, err: %w", cpu.register.pc+1, err)
		}
		upperAddr := lowerAddr + 1

		lower, err := cpu.memory.Read(uint16(lowerAddr))
		if err != nil {
			return 0, 0, fmt.Errorf("fetchAddr: IndirectY: failed read lower. addr: %x, err: %w", lowerAddr, err)
		}
		upper, err := cpu.memory.Read(uint16(upperAddr))
		if err != nil {
			return 0, 0, fmt.Errorf("fetchAddr: IndirectY: failed read upper. addr: %x, err: %w", upperAddr, err)
		}

		addr := bit_helper.BytesToUint16(lower, upper)
		beforeAddr := addr

		addr += uint16(cpu.register.y)
		if (beforeAddr & 0xFF00) != (addr & 0xFF00) {
			additionalCycle = 1
		}
		return addr, additionalCycle, nil
	case Indirect:
		// **(addr)

		indirectLower, err := cpu.memory.Read(cpu.register.pc + 1)
		if err != nil {
			return 0, 0, fmt.Errorf("fetchAddr: Indirect: failed read indirectLower. addr: %x, err: %w", cpu.register.pc+1, err)
		}

		indirectUpper, err := cpu.memory.Read(cpu.register.pc + 2)
		if err != nil {
			return 0, 0, fmt.Errorf("fetchAddr: Indirect: failed read indirectUpper. addr: %x, err: %w", cpu.register.pc+2, err)
		}

		// インクリメントにおいて下位バイトからのキャリーを無視するために、下位バイトに加算してからキャストする
		// 符号なし整数の加算のオーバーフロー時の挙動を期待しているので、未定義かも

		lowerAddr := bit_helper.BytesToUint16(indirectLower, indirectUpper)

		// 6502 CPU のバグ：下位バイトが 0xFF の場合、次のアドレスはページ境界をまたがず、下位バイトのみラップアラウンドする。
		upperAddr := bit_helper.BytesToUint16(indirectLower+1, indirectUpper)

		lower, err := cpu.memory.Read(lowerAddr)
		if err != nil {
			return 0, 0, fmt.Errorf("fetchAddr: Indirect: failed read lower. addr: %x, err: %w", lowerAddr, err)
		}

		upper, err := cpu.memory.Read(upperAddr)
		if err != nil {
			return 0, 0, fmt.Errorf("fetchAddr: Indirect: failed read upper. addr: %x, err: %w", upperAddr, err)
		}

		addr := bit_helper.BytesToUint16(lower, upper)
		return addr, 0, nil

	default:
		return 0, 0, fmt.Errorf("fetchAddr: invalid addressing mode was specified. mode: %d", mode)
	}
}

// 加算処理
func (cpu *CPU) adc(opcode Opcode) (cycles uint8, err error) {
	if opcode.Mnemonic != ADC {
		return 0, fmt.Errorf("invalid mnemonic was specified. mnemonic: %v", opcode.Mnemonic)
	}

	operand, additionalCycle, err := cpu.fetchArg(opcode.AddressingMode)
	if err != nil {
		return 0, fmt.Errorf("CPU: adc: failed fetchArg. err: %w", err)
	}

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
	cpu.setA(byte(result & 0xFF))

	// ゼロフラグ
	cpu.setFlag(zeroFlag, cpu.register.a == 0)

	// ネガティブフラグの更新
	cpu.setFlag(negativeFlag, cpu.register.a&0x80 != 0)

	// overflow
	// http://forums.nesdev.com/viewtopic.php?t=6331
	overflow := ((cpu.register.a^operand)&0x80 == 0) && ((cpu.register.a^byte(result))&0x80 != 0)
	cpu.setFlag(overflowFlag, overflow)

	// PC
	cpu.incrementPC(uint16(opcode.Length))
	return opcode.Cycles + additionalCycle, nil
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

func (cpu *CPU) setPC(count uint16) {
	cpu.register.pc = count
}

func (cpu *CPU) setA(a byte) {
	cpu.register.a = a
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
