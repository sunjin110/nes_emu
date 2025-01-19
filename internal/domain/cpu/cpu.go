package cpu

import (
	"errors"
	"fmt"

	"github.com/sunjin110/nes_emu/internal/domain/cpu/internal/memory"
	"github.com/sunjin110/nes_emu/internal/domain/ppu"
	"github.com/sunjin110/nes_emu/internal/domain/prgrom"
	"github.com/sunjin110/nes_emu/pkg/bit_helper"
)

// CPU document: https://www.nesdev.org/wiki/CPU
type CPU struct {
	memory   memory.Memory
	register Register
}

func NewCPU(prgROM prgrom.PRGROM, ppu ppu.PPU) (*CPU, error) {
	memory := memory.NewMemory(prgROM, ppu)

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
		cycles, err = cpu.adc(opcode)
	case AND:
		cycles, err = cpu.and(opcode)
	case ASL:
		cycles, err = cpu.asl(opcode)
	case BCC:
		cycles, err = cpu.bcc(opcode)
	case BCS:
		cycles, err = cpu.bcs(opcode)
	case BEQ:
		cycles, err = cpu.beq(opcode)
	case BIT:
		cycles, err = cpu.bit(opcode)
	case BMI:
		cycles, err = cpu.bmi(opcode)
	case BNE:
		cycles, err = cpu.bne(opcode)
	case BPL:
		cycles, err = cpu.bpl(opcode)
	case BRK:
		cycles, err = cpu.brk(opcode)
	case BVC:
		cycles, err = cpu.bvc(opcode)
	case BVS:
		cycles, err = cpu.bvs(opcode)
	case CLC:
		cycles, err = cpu.clc(opcode)
	case CLD:
		cycles, err = cpu.cld(opcode)
	case CLI:
		cycles, err = cpu.cli(opcode)
	case CLV:
		cycles, err = cpu.clv(opcode)
	case CMP:
		cycles, err = cpu.cmp(opcode)
	case CPX:
		cycles, err = cpu.cpx(opcode)
	case CPY:
		cycles, err = cpu.cpy(opcode)
	case DEC:
		cycles, err = cpu.dec(opcode)
	case DEX:
		cycles, err = cpu.dex(opcode)
	case DEY:
		cycles, err = cpu.dey(opcode)
	case EOR:
		cycles, err = cpu.eor(opcode)
	case INC:
		cycles, err = cpu.inc(opcode)
	case INX:
		cycles, err = cpu.inx(opcode)
	case INY:
		cycles, err = cpu.iny(opcode)
	case JMP:
		cycles, err = cpu.jmp(opcode)
	case JSR:
		cycles, err = cpu.jsr(opcode)
	case LDA:
		cycles, err = cpu.lda(opcode)
	case LDX:
		cycles, err = cpu.ldx(opcode)
	case LDY:
		cycles, err = cpu.ldy(opcode)
	case LSR:
		cycles, err = cpu.lsr(opcode)
	case NOP:
		cycles, err = cpu.nop(opcode)
	case ORA:
		cycles, err = cpu.ora(opcode)
	case PHA:
		cycles, err = cpu.pha(opcode)
	case PHP:
		cycles, err = cpu.php(opcode)
	case PLA:
		cycles, err = cpu.pla(opcode)
	case PLP:
		cycles, err = cpu.plp(opcode)
	case ROL:
		cycles, err = cpu.rol(opcode)
	case ROR:
		cycles, err = cpu.ror(opcode)
	case RTI:
		cycles, err = cpu.rti(opcode)
	case RTS:
		cycles, err = cpu.rts(opcode)
	case SBC:
		cycles, err = cpu.sbc(opcode)
	case SEC:
		cycles, err = cpu.sec(opcode)
	case SED:
		cycles, err = cpu.sed(opcode)
	case SEI:
		cycles, err = cpu.sei(opcode)
	case STA:
		cycles, err = cpu.sta(opcode)
	case STX:
		cycles, err = cpu.stx(opcode)
	case STY:
		cycles, err = cpu.sty(opcode)
	case TAX:
		cycles, err = cpu.tax(opcode)
	case TAY:
		cycles, err = cpu.tay(opcode)
	case TSX:
		cycles, err = cpu.tsx(opcode)
	case TXA:
		cycles, err = cpu.txa(opcode)
	case TXS:
		cycles, err = cpu.txs(opcode)
	case TYA:
		cycles, err = cpu.tya(opcode)
	}
	if err != nil {
		return 0, fmt.Errorf("CPU: failed run. opcode: %+v, err: %w", opcode, err)
	}
	return cycles, nil
}

// doc: https://www.nesdev.org/wiki/CPU_interrupts
func (cpu *CPU) Interrupt(t InterruptType) error {

	nested := cpu.getFlag(interruptFlag)
	if nested && (t == InterruptTypeBRK || t == InterruptTypeIRQ) {
		// nested interrupt が許されるのは RESET と NMI のみ
		// エラーにすべき?
		return nil
	}

	// 割り込むフラグを追加する
	cpu.setFlag(interruptFlag, true)

	switch t {
	case InterruptTypeNMI:

		lowerPC, upperPC := bit_helper.Uint16ToBytes(cpu.register.pc)
		if err := cpu.pushStack(lowerPC); err != nil {
			return fmt.Errorf("CPU: Interrupt: NMI: failed push stack. err: %w", err)
		}
		if err := cpu.pushStack(upperPC); err != nil {
			return fmt.Errorf("CPU: Interrupt: NMI: failed push stack. err: %w", err)
		}

		// NMI, IRQ のときは 5, 4 bit 目を0にする
		cpu.setFlag(breakFlag, false)
		pushData := cpu.register.p | (1 << 5) // 5bit目(未使用)を必ず1にする
		if err := cpu.pushStack(pushData); err != nil {
			return fmt.Errorf("CPU: Interrupt: NMI: failed push stack. err: %w", err)
		}

		interruptLowerPC, err := cpu.memory.Read(memory.NMIInterruptLowerPCAddr)
		if err != nil {
			return fmt.Errorf("CPU: Interrupt: NMI: failed read memory. err: %w", err)
		}

		interruptUpperPC, err := cpu.memory.Read(memory.NMIInterruptUpperPCAddr)
		if err != nil {
			return fmt.Errorf("CPU: Interrupt: NMI: failed read memory. err: %w", err)
		}

		pc := bit_helper.BytesToUint16(interruptLowerPC, interruptUpperPC)
		cpu.setPC(pc)
		return nil
	case InterruptTypeReset:
		interruptLowerPC, err := cpu.memory.Read(memory.ResetInterruptLowerPCAddr)
		if err != nil {
			return fmt.Errorf("CPU: Interrupt: Reset: failed read memory. err: %w", err)
		}

		interruptUpperPC, err := cpu.memory.Read(memory.ResetInterruptUpperPCAddr)
		if err != nil {
			return fmt.Errorf("CPU: Interrupt: Reset: failed read memory. err: %w", err)
		}

		// https://www.pagetable.com/?p=410
		cpu.setSP(initSPAddr)

		pc := bit_helper.BytesToUint16(interruptLowerPC, interruptUpperPC)
		cpu.setPC(pc)
		return nil
	case InterruptTypeIRQ:
		lowerPC, upperPC := bit_helper.Uint16ToBytes(cpu.register.pc)
		if err := cpu.pushStack(lowerPC); err != nil {
			return fmt.Errorf("CPU: Interrupt: IRQ: failed push stack. err: %w", err)
		}
		if err := cpu.pushStack(upperPC); err != nil {
			return fmt.Errorf("CPU: Interrupt: IRQ: failed push stack. err: %w", err)
		}

		// NMI, IRQ のときは 5, 4 bit 目を0にする
		cpu.setFlag(breakFlag, false)
		pushData := cpu.register.p | (1 << 5) // 5bit目(未使用)を必ず1にする
		if err := cpu.pushStack(pushData); err != nil {
			return fmt.Errorf("CPU: Interrupt: IRQ: failed push stack. err: %w", err)
		}

		interruptLowerPC, err := cpu.memory.Read(memory.IRQInterruptLowerPCAddr)
		if err != nil {
			return fmt.Errorf("CPU: Interrupt: IRQ: failed read memory. err: %w", err)
		}

		interruptUpperPC, err := cpu.memory.Read(memory.IRQInterruptUpperPCAddr)
		if err != nil {
			return fmt.Errorf("CPU: Interrupt: IRQ: failed read memory. err: %w", err)
		}

		pc := bit_helper.BytesToUint16(interruptLowerPC, interruptUpperPC)
		cpu.setPC(pc)
		return nil
	case InterruptTypeBRK:
		cpu.incrementPC(1)

		lowerPC, upperPC := bit_helper.Uint16ToBytes(cpu.register.pc)
		if err := cpu.pushStack(lowerPC); err != nil {
			return fmt.Errorf("CPU: Interrupt: BRK: failed push stack. err: %w", err)
		}
		if err := cpu.pushStack(upperPC); err != nil {
			return fmt.Errorf("CPU: Interrupt: BRK: failed push stack. err: %w", err)
		}

		cpu.setFlag(breakFlag, true)
		pushData := cpu.register.p | (1 << 5) // 5bit目(未使用)を必ず1にする
		if err := cpu.pushStack(pushData); err != nil {
			return fmt.Errorf("CPU: Interrupt: BRK: failed push stack. err: %w", err)
		}

		interruptLowerPC, err := cpu.memory.Read(memory.BreakInterruptLowerPCAddr)
		if err != nil {
			return fmt.Errorf("CPU: Interrupt: BRK: failed read memory. err: %w", err)
		}

		interruptUpperPC, err := cpu.memory.Read(memory.BreakInterruptUpperPCAddr)
		if err != nil {
			return fmt.Errorf("CPU: Interrupt: BRK: failed read memory. err: %w", err)
		}

		pc := bit_helper.BytesToUint16(interruptLowerPC, interruptUpperPC)
		cpu.setPC(pc)
		return nil
	default:
		return fmt.Errorf("CPU: Interrupt: undefined interrupt type was specified. type: %d", t)
	}
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
// https://www.nesdev.org/wiki/Instruction_reference#ADC
// A = A + memory + C
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
	cpu.setFlag(negativeFlag, cpu.isNegative(cpu.register.a))

	// overflow
	// http://forums.nesdev.com/viewtopic.php?t=6331
	overflow := ((cpu.register.a^operand)&0x80 == 0) && ((cpu.register.a^byte(result))&0x80 != 0)
	cpu.setFlag(overflowFlag, overflow)

	// PC
	cpu.incrementPC(uint16(opcode.Length))
	return opcode.Cycles + additionalCycle, nil
}

// and
// document: https://www.nesdev.org/wiki/Instruction_reference#AND
// A = A & memory
func (cpu *CPU) and(opcode Opcode) (cycles uint8, err error) {
	if opcode.Mnemonic != AND {
		return 0, fmt.Errorf("invalid mnemonic was specified. mnemonic: %v", opcode.Mnemonic)
	}

	arg, additionalCycle, err := cpu.fetchArg(opcode.AddressingMode)
	if err != nil {
		return 0, fmt.Errorf("CPU: and: failed fetchArg. err: %w", err)
	}

	result := cpu.register.a & arg

	cpu.setFlag(zeroFlag, result == 0)
	cpu.setFlag(negativeFlag, cpu.isNegative(result))
	cpu.setA(result)

	cpu.incrementPC(uint16(opcode.Length))
	return opcode.Cycles + additionalCycle, nil
}

// asl
// doc: https://www.nesdev.org/wiki/Instruction_reference#ASL
// value = value << 1, or visually: C <- [76543210] <- 0
func (cpu *CPU) asl(opcode Opcode) (cycles uint8, err error) {
	if opcode.Mnemonic != ASL {
		return 0, fmt.Errorf("invalid mnemonic was specified. mnemonic: %v", opcode.Mnemonic)
	}

	arg, additionalCycles, err := cpu.fetchArg(opcode.AddressingMode)
	if err != nil {
		return 0, fmt.Errorf("CPU: asl: failed fetchArg. err: %w", err)
	}

	result := arg << 1

	// 最上位ビット(MSB)が立っているときにシフトしたらcarryする
	cpu.setFlag(carryFlag, arg&0x80 != 0)
	cpu.setFlag(zeroFlag, result == 0)
	cpu.setFlag(negativeFlag, cpu.isNegative(result))

	if opcode.AddressingMode == Accumulator {
		cpu.setA(result)
	} else {
		// ASLはメモリ書き込み時に追加サイクルを必要としないため
		addr, _, err := cpu.fetchAddr(opcode.AddressingMode)
		if err != nil {
			return 0, fmt.Errorf("CPU: asl: failed fetch addr for writing result. err: %w", err)
		}
		if err := cpu.memory.Write(addr, result); err != nil {
			return 0, fmt.Errorf("CPU: asl: failed write memory. addr: %x, err: %w", addr, err)
		}
	}

	cpu.incrementPC(uint16(opcode.Length))
	return opcode.Cycles + additionalCycles, nil
}

// bcc
// doc: https://www.nesdev.org/wiki/Instruction_reference#BCC
// PC = PC + 2 + memory (signed)
func (cpu *CPU) bcc(opcode Opcode) (cycles uint8, err error) {
	if opcode.Mnemonic != BCC {
		return 0, fmt.Errorf("invalid mnemonic was specified. mnemonic: %v", opcode.Mnemonic)
	}

	if cpu.getFlag(carryFlag) {
		cpu.incrementPC(uint16(opcode.Length))
		return opcode.Cycles, nil
	}

	addr, additionalCycles, err := cpu.fetchAddr(opcode.AddressingMode)
	if err != nil {
		return 0, fmt.Errorf("CPU: bcc: failed fetchAddr. err: %w", err)
	}

	// PCを更新
	cpu.setPC(addr)

	// 分岐成立時に+1する
	// 分岐が成立すると、分岐先のアドレスを再計算して新しい命令をフェッチする必要があり、パイプラインが「破棄」されます。
	// このパイプライン破棄により、分岐成立時には追加の1サイクルが必要になります。
	return opcode.Cycles + additionalCycles + 1, nil
}

// bcs
// doc: https://www.nesdev.org/wiki/Instruction_reference#BCS
// PC = PC + 2 + memory (signed)
func (cpu *CPU) bcs(opcode Opcode) (cycles uint8, err error) {
	if opcode.Mnemonic != BCS {
		return 0, fmt.Errorf("invalid mnemonic was specified. mnemonic: %v", opcode.Mnemonic)
	}

	if !cpu.getFlag(carryFlag) {
		cpu.incrementPC(uint16(opcode.Length))
		return opcode.Cycles, nil
	}
	addr, additionalCycles, err := cpu.fetchAddr(opcode.AddressingMode)
	if err != nil {
		return 0, fmt.Errorf("CPU: bcs: failed fetchAddr. err: %w", err)
	}

	// PCを更新
	cpu.setPC(addr)

	// 分岐成立時に+1する
	// 分岐が成立すると、分岐先のアドレスを再計算して新しい命令をフェッチする必要があり、パイプラインが「破棄」されます。
	// このパイプライン破棄により、分岐成立時には追加の1サイクルが必要になります。
	return opcode.Cycles + additionalCycles + 1, nil
}

// beq
// doc: https://www.nesdev.org/wiki/Instruction_reference#BEQ
// PC = PC + 2 + memory (signed)
func (cpu *CPU) beq(opcode Opcode) (cycles uint8, err error) {
	if opcode.Mnemonic != BEQ {
		return 0, fmt.Errorf("invalid mnemonic was specified. mnemonic: %v", opcode.Mnemonic)
	}

	if !cpu.getFlag(zeroFlag) {
		cpu.incrementPC(uint16(opcode.Length))
		return opcode.Cycles, nil
	}

	addr, additionalCycles, err := cpu.fetchAddr(opcode.AddressingMode)
	if err != nil {
		return 0, fmt.Errorf("CPU: beq: failed fetchAddr. err: %w", err)
	}

	cpu.setPC(addr)
	return opcode.Cycles + additionalCycles + 1, nil
}

// bit
// doc: https://www.nesdev.org/wiki/Instruction_reference#BIT
// A & memory
func (cpu *CPU) bit(opcode Opcode) (cycles uint8, err error) {
	if opcode.Mnemonic != BIT {
		return 0, fmt.Errorf("invalid mnemonic was specified. mnemonic: %v", opcode.Mnemonic)
	}

	arg, additionalCycles, err := cpu.fetchArg(opcode.AddressingMode)
	if err != nil {
		return 0, fmt.Errorf("CPU: bit: failed fetchArg. err: %w", err)
	}

	result := cpu.register.a & arg

	cpu.setFlag(zeroFlag, result == 0)
	cpu.setFlag(overflowFlag, (arg&0x40) == 0x40) // 6bit目が1ならoverflowをtrue
	cpu.setFlag(negativeFlag, cpu.isNegative(result))

	cpu.incrementPC(uint16(opcode.Length))
	return opcode.Cycles + additionalCycles, nil
}

// bmi Branch if Minus
// doc: https://www.nesdev.org/wiki/Instruction_reference#BMI
// PC = PC + 2 + memory (signed)
func (cpu *CPU) bmi(opcode Opcode) (cycles uint8, err error) {
	if opcode.Mnemonic != BMI {
		return 0, fmt.Errorf("invalid mnemonic was specified. mnemonic: %v", opcode.Mnemonic)
	}

	if !cpu.getFlag(negativeFlag) {
		cpu.incrementPC(uint16(opcode.Length))
		return opcode.Cycles, nil
	}

	addr, additionalCycles, err := cpu.fetchAddr(opcode.AddressingMode)
	if err != nil {
		return 0, fmt.Errorf("CPU: bmi: failed fetchAddr. err: %w", err)
	}

	cpu.setPC(addr)
	// 分岐成立時に+1する
	// 分岐が成立すると、分岐先のアドレスを再計算して新しい命令をフェッチする必要があり、パイプラインが「破棄」されます。
	// このパイプライン破棄により、分岐成立時には追加の1サイクルが必要になります。
	return opcode.Cycles + additionalCycles + 1, nil
}

// bne Branch if Not Equal
// doc: https://www.nesdev.org/wiki/Instruction_reference#BNE
// PC = PC + 2 + memory (signed)
func (cpu *CPU) bne(opcode Opcode) (cycles uint8, err error) {
	if opcode.Mnemonic != BNE {
		return 0, fmt.Errorf("invalid mnemonic was specified. mnemonic: %v", opcode.Mnemonic)
	}

	if cpu.getFlag(zeroFlag) {
		cpu.incrementPC(uint16(opcode.Length))
		return opcode.Cycles, nil
	}

	addr, additionalCycles, err := cpu.fetchAddr(opcode.AddressingMode)
	if err != nil {
		return 0, fmt.Errorf("CPU: bne: failed fetchAddr. err: %w", err)
	}

	cpu.setPC(addr)

	// 分岐成立時に+1する
	// 分岐が成立すると、分岐先のアドレスを再計算して新しい命令をフェッチする必要があり、パイプラインが「破棄」されます。
	// このパイプライン破棄により、分岐成立時には追加の1サイクルが必要になります。
	return opcode.Cycles + additionalCycles + 1, nil
}

// bpl Branch if Plus
// doc: https://www.nesdev.org/wiki/Instruction_reference#BPL
// PC = PC + 2 + memory (signed
func (cpu *CPU) bpl(opcode Opcode) (cycles uint8, err error) {
	if opcode.Mnemonic != BPL {
		return 0, fmt.Errorf("invalid mnemonic was specified. mnemonic: %v", opcode.Mnemonic)
	}

	if cpu.getFlag(negativeFlag) {
		cpu.incrementPC(uint16(opcode.Length))
		return opcode.Cycles, nil
	}

	addr, addtionalCycles, err := cpu.fetchAddr(opcode.AddressingMode)
	if err != nil {
		return 0, fmt.Errorf("CPU: bpl: failed fetchAddr. err: %w", err)
	}

	cpu.setPC(addr)

	// 分岐成立時に+1する
	// 分岐が成立すると、分岐先のアドレスを再計算して新しい命令をフェッチする必要があり、パイプラインが「破棄」されます。
	// このパイプライン破棄により、分岐成立時には追加の1サイクルが必要になります。
	return opcode.Cycles + addtionalCycles + 1, nil
}

// brk
// doc: https://www.nesdev.org/wiki/Instruction_reference#BRK
// push PC + 2 to stack
// push NV11DIZC flags to stack
// PC = ($FFFE)
func (cpu *CPU) brk(opcode Opcode) (cycles uint8, err error) {
	if opcode.Mnemonic != BRK {
		return 0, fmt.Errorf("invalid mnemonic was specified. mnemonic: %v", opcode.Mnemonic)
	}
	cpu.setFlag(breakFlag, true)
	if err := cpu.Interrupt(InterruptTypeBRK); err != nil {
		return 0, fmt.Errorf("CPU: brk: faied interrupt. err: %w", err)
	}
	return opcode.Cycles, nil
}

// bvc
// doc: https://www.nesdev.org/wiki/Instruction_reference#BVC
// PC = PC + 2 + memory (signed)
func (cpu *CPU) bvc(opcode Opcode) (cycles uint8, err error) {
	if opcode.Mnemonic != BVC {
		return 0, fmt.Errorf("invalid mnemonic was specified. mnemonic: %v", opcode.Mnemonic)
	}

	if cpu.getFlag(overflowFlag) {
		cpu.incrementPC(uint16(opcode.Length))
		return opcode.Cycles, nil
	}

	addr, additionalCycles, err := cpu.fetchAddr(opcode.AddressingMode)
	if err != nil {
		return 0, fmt.Errorf("CPU: bvc: faied interrupt. err: %w", err)
	}
	cpu.setPC(addr)

	// 分岐成立時に+1する
	// 分岐が成立すると、分岐先のアドレスを再計算して新しい命令をフェッチする必要があり、パイプラインが「破棄」されます。
	// このパイプライン破棄により、分岐成立時には追加の1サイクルが必要になります。
	return opcode.Cycles + additionalCycles + 1, nil
}

// bvc Branch if Overflow Set
// doc: https://www.nesdev.org/wiki/Instruction_reference#BVS
// PC = PC + 2 + memory (signed)
func (cpu *CPU) bvs(opcode Opcode) (cycles uint8, err error) {
	if opcode.Mnemonic != BVS {
		return 0, fmt.Errorf("invalid mnemonic was specified. mnemonic: %v", opcode.Mnemonic)
	}

	if !cpu.getFlag(overflowFlag) {
		cpu.incrementPC(uint16(opcode.Length))
		return opcode.Cycles, nil
	}

	addr, additionalCycles, err := cpu.fetchAddr(opcode.AddressingMode)
	if err != nil {
		return 0, fmt.Errorf("CPU: bvs: faied interrupt. err: %w", err)
	}
	cpu.setPC(addr)

	// 分岐成立時に+1する
	// 分岐が成立すると、分岐先のアドレスを再計算して新しい命令をフェッチする必要があり、パイプラインが「破棄」されます。
	// このパイプライン破棄により、分岐成立時には追加の1サイクルが必要になります。
	return opcode.Cycles + additionalCycles + 1, nil
}

// clc: Clear Carry
// doc: https://www.nesdev.org/wiki/Instruction_reference#CLC
// C = 0
func (cpu *CPU) clc(opcode Opcode) (cycles uint8, err error) {
	if opcode.Mnemonic != CLC {
		return 0, fmt.Errorf("invalid mnemonic was specified. mnemonic: %v", opcode.Mnemonic)
	}
	cpu.setFlag(carryFlag, false)
	cpu.incrementPC(uint16(opcode.Length))
	return opcode.Cycles, nil
}

// cld: Clear Decimal
// doc: https://www.nesdev.org/wiki/Instruction_reference#CLD
// D = 0
func (cpu *CPU) cld(opcode Opcode) (cycles uint8, err error) {
	if opcode.Mnemonic != CLD {
		return 0, fmt.Errorf("invalid mnemonic was specified. mnemonic: %v", opcode.Mnemonic)
	}
	cpu.setFlag(decimalFlag, false)
	cpu.incrementPC(uint16(opcode.Length))
	return opcode.Cycles, nil
}

// cli Clear Interrupt Disable
// doc: https://www.nesdev.org/wiki/Instruction_reference#CLI
func (cpu *CPU) cli(opcode Opcode) (cycles uint8, err error) {
	if opcode.Mnemonic != CLI {
		return 0, fmt.Errorf("invalid mnemonic was specified. mnemonic: %v", opcode.Mnemonic)
	}
	cpu.setFlag(interruptFlag, false)
	cpu.incrementPC(uint16(opcode.Length))
	return opcode.Cycles, nil
}

// clv: Clear Overflow
// doc: https://www.nesdev.org/wiki/Instruction_reference#CLV
func (cpu *CPU) clv(opcode Opcode) (cycles uint8, err error) {
	if opcode.Mnemonic != CLV {
		return 0, fmt.Errorf("invalid mnemonic was specified. mnemonic: %v", opcode.Mnemonic)
	}
	cpu.setFlag(overflowFlag, false)
	cpu.incrementPC(uint16(opcode.Length))
	return opcode.Cycles, nil
}

// cmp: Compare A
// doc: https://www.nesdev.org/wiki/Instruction_reference#CLV
// A - memory
func (cpu *CPU) cmp(opcode Opcode) (cycles uint8, err error) {
	if opcode.Mnemonic != CMP {
		return 0, fmt.Errorf("invalid mnemonic was specified. mnemonic: %v", opcode.Mnemonic)
	}

	arg, additionalCycles, err := cpu.fetchArg(opcode.AddressingMode)
	if err != nil {
		return 0, fmt.Errorf("CPU: cmp: failed fetchArg. err: %w", err)
	}

	result := cpu.register.a - arg

	cpu.setFlag(carryFlag, cpu.register.a >= arg)
	cpu.setFlag(zeroFlag, result == 0)
	cpu.setFlag(negativeFlag, cpu.isNegative(result))

	cpu.incrementPC(uint16(opcode.Length))
	return opcode.Cycles + additionalCycles, nil
}

// cpx: Compare X
func (cpu *CPU) cpx(opcode Opcode) (cycles uint8, err error) {
	if opcode.Mnemonic != CPX {
		return 0, fmt.Errorf("invalid mnemonic was specified. mnemonic: %v", opcode.Mnemonic)
	}

	arg, additionalCycles, err := cpu.fetchArg(opcode.AddressingMode)
	if err != nil {
		return 0, fmt.Errorf("CPU: cpx: failed fetchArg. err: %w", err)
	}

	result := cpu.register.x - arg

	cpu.setFlag(carryFlag, cpu.register.x >= arg)
	cpu.setFlag(zeroFlag, result == 0)
	cpu.setFlag(negativeFlag, cpu.isNegative(result))

	cpu.incrementPC(uint16(opcode.Length))
	return opcode.Cycles + additionalCycles, nil
}

// cpy: Compare Y
func (cpu *CPU) cpy(opcode Opcode) (cycles uint8, err error) {
	if opcode.Mnemonic != CPY {
		return 0, fmt.Errorf("invalid mnemonic was specified. mnemonic: %v", opcode.Mnemonic)
	}

	arg, additionalCycles, err := cpu.fetchArg(opcode.AddressingMode)
	if err != nil {
		return 0, fmt.Errorf("CPU: cpy: failed fetchArg. err: %w", err)
	}

	result := cpu.register.y - arg

	cpu.setFlag(carryFlag, cpu.register.y >= arg)
	cpu.setFlag(zeroFlag, result == 0)
	cpu.setFlag(negativeFlag, cpu.isNegative(result))

	cpu.incrementPC(uint16(opcode.Length))
	return opcode.Cycles + additionalCycles, nil
}

// dec: Decrement Memory
func (cpu *CPU) dec(opcode Opcode) (cycles uint8, err error) {
	if opcode.Mnemonic != DEC {
		return 0, fmt.Errorf("invalid mnemonic was specified. mnemonic: %v", opcode.Mnemonic)
	}

	// ImmidiateやAccumulatorは必ず渡されない

	addr, additionalCycles, err := cpu.fetchAddr(opcode.AddressingMode)
	if err != nil {
		return 0, fmt.Errorf("CPU: dec: failed fetchAddr. err: %w", err)
	}

	arg, err := cpu.memory.Read(addr)
	if err != nil {
		return 0, fmt.Errorf("CPU: dec: failed get memory value. err: %w", err)
	}

	result := arg - 1

	cpu.setFlag(zeroFlag, result == 0)
	cpu.setFlag(negativeFlag, cpu.isNegative(result))

	// write
	if err := cpu.memory.Write(addr, result); err != nil {
		return 0, fmt.Errorf("CPU: dec: failed write memory. addr: %x, value: %x, err: %w", addr, result, err)
	}

	cpu.incrementPC(uint16(opcode.Length))

	return opcode.Cycles + additionalCycles, nil
}

func (cpu *CPU) dex(opcode Opcode) (cycles uint8, err error) {
	if opcode.Mnemonic != DEX {
		return 0, fmt.Errorf("invalid mnemonic was specified. mnemonic: %v", opcode.Mnemonic)
	}

	result := cpu.register.x - 1
	cpu.setFlag(zeroFlag, result == 0)
	cpu.setFlag(negativeFlag, cpu.isNegative(result))
	cpu.setX(result)
	cpu.incrementPC(uint16(opcode.Length))
	return opcode.Cycles, nil
}

func (cpu *CPU) dey(opcode Opcode) (cycles uint8, err error) {
	if opcode.Mnemonic != DEY {
		return 0, fmt.Errorf("invalid mnemonic was specified. mnemonic: %v", opcode.Mnemonic)
	}

	result := cpu.register.y - 1
	cpu.setFlag(zeroFlag, result == 0)
	cpu.setFlag(negativeFlag, cpu.isNegative(result))
	cpu.setY(result)
	cpu.incrementPC(uint16(opcode.Length))
	return opcode.Cycles, nil
}

func (cpu *CPU) eor(opcode Opcode) (cycles uint8, err error) {
	if opcode.Mnemonic != EOR {
		return 0, fmt.Errorf("invalid mnemonic was specified. mnemonic: %v", opcode.Mnemonic)
	}

	arg, additionalCycles, err := cpu.fetchArg(opcode.AddressingMode)
	if err != nil {
		return 0, fmt.Errorf("CPU: eor: failed fetchArg. err: %w", err)
	}

	result := cpu.register.a ^ arg

	cpu.setFlag(zeroFlag, result == 0)
	cpu.setFlag(negativeFlag, cpu.isNegative(result))

	cpu.setA(result)

	cpu.incrementPC(uint16(opcode.Length))
	return opcode.Cycles + additionalCycles, nil
}

func (cpu *CPU) inc(opcode Opcode) (cycles uint8, err error) {
	if opcode.Mnemonic != INC {
		return 0, fmt.Errorf("invalid mnemonic was specified. mnemonic: %v", opcode.Mnemonic)
	}

	addr, additionalCycles, err := cpu.fetchAddr(opcode.AddressingMode)
	if err != nil {
		return 0, fmt.Errorf("CPU: inc: failed fetchAddr. err: %w", err)
	}

	arg, err := cpu.memory.Read(addr)
	if err != nil {
		return 0, fmt.Errorf("CPU: inc: failed get memory value. err: %w", err)
	}

	result := arg + 1

	cpu.setFlag(zeroFlag, result == 0)
	cpu.setFlag(negativeFlag, cpu.isNegative(result))

	if err := cpu.memory.Write(addr, result); err != nil {
		return 0, fmt.Errorf("CPU: inc: failed write memory. addr: %x, value: %x, err: %w", addr, result, err)
	}

	cpu.incrementPC(uint16(opcode.Length))
	return opcode.Cycles + additionalCycles, nil
}

func (cpu *CPU) inx(opcode Opcode) (cycles uint8, err error) {
	if opcode.Mnemonic != INX {
		return 0, fmt.Errorf("invalid mnemonic was specified. mnemonic: %v", opcode.Mnemonic)
	}

	result := cpu.register.x + 1
	cpu.setFlag(zeroFlag, result == 0)
	cpu.setFlag(negativeFlag, cpu.isNegative(result))

	cpu.setX(result)

	cpu.incrementPC(uint16(opcode.Length))
	return opcode.Cycles, nil
}

func (cpu *CPU) iny(opcode Opcode) (cycles uint8, err error) {
	if opcode.Mnemonic != INY {
		return 0, fmt.Errorf("invalid mnemonic was specified. mnemonic: %v", opcode.Mnemonic)
	}

	result := cpu.register.y + 1
	cpu.setFlag(zeroFlag, result == 0)
	cpu.setFlag(negativeFlag, cpu.isNegative(result))

	cpu.setY(result)

	cpu.incrementPC(uint16(opcode.Length))
	return opcode.Cycles, nil
}

// jmp: Jmp to Address
// doc: https://www.nesdev.org/wiki/Instruction_reference#JMP
func (cpu *CPU) jmp(opcode Opcode) (cycles uint8, err error) {
	if opcode.Mnemonic != JMP {
		return 0, fmt.Errorf("invalid mnemonic was specified. mnemonic: %v", opcode.Mnemonic)
	}

	// JMP命令は追加サイクルは発生しない
	addr, _, err := cpu.fetchAddr(opcode.AddressingMode)
	if err != nil {
		return 0, fmt.Errorf("CPU: jmp: failed fetchAddr. err: %w", err)
	}

	cpu.setPC(addr)
	return opcode.Cycles, nil
}

// jsr: Jump to Subroutine
// doc: https://www.nesdev.org/wiki/Instruction_reference#JSR
func (cpu *CPU) jsr(opcode Opcode) (cycles uint8, err error) {
	if opcode.Mnemonic != JSR {
		return 0, fmt.Errorf("invalid mnemonic was specified. mnemonic: %v", opcode.Mnemonic)
	}

	jumpAddr, _, err := cpu.fetchAddr(opcode.AddressingMode)
	if err != nil {
		return 0, fmt.Errorf("CPU: jsr: failed fetchAddr. err: %w", err)
	}

	// push
	// リターンアドレスは PC + 3 だが、それから 1 を引いたものを stack にプッシュする
	returnAddr := cpu.register.pc + 2

	// lower -> upperの順にpush
	// ここは参考と違ったので注意
	lower, upper := bit_helper.Uint16ToBytes(returnAddr)
	if err := cpu.pushStack(lower); err != nil {
		return 0, fmt.Errorf("CPU: jsr: failed push lower returnAddr to stack. err: %w", err)
	}
	if err := cpu.pushStack(upper); err != nil {
		return 0, fmt.Errorf("CPU: jsr: failed push upper returnAddr to stack. err: %w", err)
	}

	cpu.setPC(jumpAddr)
	return opcode.Cycles, nil
}

// lda Load A
func (cpu *CPU) lda(opcode Opcode) (cycles uint8, err error) {
	if opcode.Mnemonic != LDA {
		return 0, fmt.Errorf("invalid mnemonic was specified. mnemonic: %v", opcode.Mnemonic)
	}

	arg, additionalCycles, err := cpu.fetchArg(opcode.AddressingMode)
	if err != nil {
		return 0, fmt.Errorf("CPU: lda: failed fetchArg. err: %w", err)
	}

	cpu.setFlag(zeroFlag, arg == 0)
	cpu.setFlag(negativeFlag, cpu.isNegative(arg))

	cpu.setA(arg)

	cpu.incrementPC(uint16(opcode.Length))
	return opcode.Cycles + additionalCycles, nil
}

// ldx: Load X
func (cpu *CPU) ldx(opcode Opcode) (cycles uint8, err error) {
	if opcode.Mnemonic != LDA {
		return 0, fmt.Errorf("invalid mnemonic was specified. mnemonic: %v", opcode.Mnemonic)
	}

	arg, additionalCycles, err := cpu.fetchArg(opcode.AddressingMode)
	if err != nil {
		return 0, fmt.Errorf("CPU: ldx: failed fetchArg. err: %w", err)
	}

	cpu.setFlag(zeroFlag, arg == 0)
	cpu.setFlag(negativeFlag, cpu.isNegative(arg))

	cpu.setX(arg)

	cpu.incrementPC(uint16(opcode.Length))
	return opcode.Cycles + additionalCycles, nil
}

// ldy: Load Y
func (cpu *CPU) ldy(opcode Opcode) (cycles uint8, err error) {
	if opcode.Mnemonic != LDA {
		return 0, fmt.Errorf("invalid mnemonic was specified. mnemonic: %v", opcode.Mnemonic)
	}

	arg, additionalCycles, err := cpu.fetchArg(opcode.AddressingMode)
	if err != nil {
		return 0, fmt.Errorf("CPU: ldy: failed fetchArg. err: %w", err)
	}

	cpu.setFlag(zeroFlag, arg == 0)
	cpu.setFlag(negativeFlag, cpu.isNegative(arg))

	cpu.setY(arg)

	cpu.incrementPC(uint16(opcode.Length))
	return opcode.Cycles + additionalCycles, nil
}

// lsr Logical Shift Right
// value = value >> 1, or visually: 0 -> [76543210] -> C
func (cpu *CPU) lsr(opcode Opcode) (cycles uint8, err error) {
	if opcode.Mnemonic != LSR {
		return 0, fmt.Errorf("invalid mnemonic was specified. mnemonic: %v", opcode.Mnemonic)
	}

	arg, additionalCycles, err := cpu.fetchArg(opcode.AddressingMode)
	if err != nil {
		return 0, fmt.Errorf("CPU: lsr: failed fetchArg. err: %w", err)
	}

	result := arg >> 1

	// 右にシフトして、最後のビットがなくなってしまう場合にcarryフラグが立つ
	cpu.setFlag(carryFlag, (arg&1) == 1)
	cpu.setFlag(zeroFlag, result == 0)
	// LSR命令では、結果の最上位ビットが常に 0 になるため、ネガティブフラグは常にクリアされる
	cpu.setFlag(negativeFlag, false)

	if opcode.AddressingMode == Accumulator {
		cpu.setA(result)
	} else {
		// arg取得時にすでにサイクルコストを支払っているんため
		addr, _, err := cpu.fetchAddr(opcode.AddressingMode)
		if err != nil {
			return 0, fmt.Errorf("CPU: lsr: failed fetch addr for writing result. err: %w", err)
		}
		if err := cpu.memory.Write(addr, result); err != nil {
			return 0, fmt.Errorf("CPU: lsr: failed write memory. addr: %x, err: %w", addr, err)
		}
	}
	cpu.incrementPC(uint16(opcode.Length))
	return opcode.Cycles + additionalCycles, nil
}

func (cpu *CPU) nop(opcode Opcode) (cycles uint8, err error) {
	if opcode.Mnemonic != NOP {
		return 0, fmt.Errorf("invalid mnemonic was specified. mnemonic: %v", opcode.Mnemonic)
	}
	// DO noting
	cpu.incrementPC(uint16(opcode.Length))
	return opcode.Cycles, nil
}

// ora Bitwise OR
// doc: https://www.nesdev.org/wiki/Instruction_reference#ORA
func (cpu *CPU) ora(opcode Opcode) (cycles uint8, err error) {
	if opcode.Mnemonic != ORA {
		return 0, fmt.Errorf("invalid mnemonic was specified. mnemonic: %v", opcode.Mnemonic)
	}

	arg, additionalCycles, err := cpu.fetchArg(opcode.AddressingMode)
	if err != nil {
		return 0, fmt.Errorf("CPU: ora: failed fetchArg. err: %w", err)
	}

	result := cpu.register.a | arg
	cpu.setFlag(zeroFlag, result == 0)
	cpu.setFlag(negativeFlag, cpu.isNegative(result))

	cpu.setA(result)

	cpu.incrementPC(uint16(opcode.Length))
	return opcode.Cycles + additionalCycles, nil
}

// pha Push A
func (cpu *CPU) pha(opcode Opcode) (cycles uint8, err error) {
	if opcode.Mnemonic != PHA {
		return 0, fmt.Errorf("invalid mnemonic was specified. mnemonic: %v", opcode.Mnemonic)
	}

	if err := cpu.pushStack(cpu.register.a); err != nil {
		return 0, fmt.Errorf("CPU: pha: failed push a to stack. err: %w", err)
	}
	cpu.incrementPC(uint16(opcode.Length))
	return opcode.Cycles, nil
}

func (cpu *CPU) php(opcode Opcode) (cycles uint8, err error) {
	if opcode.Mnemonic != PHP {
		return 0, fmt.Errorf("invalid mnemonic was specified. mnemonic: %v", opcode.Mnemonic)
	}

	// cpu.register.p
	// breakフラグは物理的にpには存在しない
	if err := cpu.pushStack(cpu.register.p | bFlagMask); err != nil {
		return 0, fmt.Errorf("CPU: php: failed push p to stack. err: %w", err)
	}
	cpu.incrementPC(uint16(opcode.Length))
	return opcode.Cycles, nil
}

func (cpu *CPU) pla(opcode Opcode) (cycles uint8, err error) {
	if opcode.Mnemonic != PLA {
		return 0, fmt.Errorf("invalid mnemonic was specified. mnemonic: %v", opcode.Mnemonic)
	}

	result, err := cpu.popStack()
	if err != nil {
		return 0, fmt.Errorf("CPU: pla: failed pop stack. err: %w", err)
	}

	cpu.setFlag(zeroFlag, result == 0)
	cpu.setFlag(negativeFlag, cpu.isNegative(result))

	cpu.setA(result)

	cpu.incrementPC(uint16(opcode.Length))
	return opcode.Cycles, nil
}

func (cpu *CPU) plp(opcode Opcode) (cycles uint8, err error) {
	if opcode.Mnemonic != PLP {
		return 0, fmt.Errorf("invalid mnemonic was specified. mnemonic: %v", opcode.Mnemonic)
	}

	result, err := cpu.popStack()
	if err != nil {
		return 0, fmt.Errorf("CPU: plp: failed pop stack. err: %w", err)
	}

	// 取得したresultのB4とB5を除去 | 現在のPフラグのB4とB5だけ抽出
	p := (result &^ bFlagMask) | (cpu.register.p & bFlagMask)
	cpu.setP(p)

	cpu.incrementPC(uint16(opcode.Length))
	return opcode.Cycles, nil
}

// rol Rotate Left
// value = value << 1 through C, or visually: C <- [76543210] <- C
func (cpu *CPU) rol(opcode Opcode) (cycles uint8, err error) {
	if opcode.Mnemonic != ROL {
		return 0, fmt.Errorf("invalid mnemonic was specified. mnemonic: %v", opcode.Mnemonic)
	}

	arg, additionalCycles, err := cpu.fetchArg(opcode.AddressingMode)
	if err != nil {
		return 0, fmt.Errorf("CPU: rol: failed fetchArg. err: %w", err)
	}

	result := arg << 1
	if cpu.getFlag(carryFlag) {
		// carryフラグがある場合のみ最後に1を追加
		result |= 1
	}

	// 最上位ビット(MSB)が立っているときにシフトしたらcarryする
	cpu.setFlag(carryFlag, arg&0x80 != 0)
	cpu.setFlag(zeroFlag, result == 0)
	cpu.setFlag(negativeFlag, cpu.isNegative(result))

	if opcode.AddressingMode == Accumulator {
		cpu.setA(result)
	} else {
		// メモリ書き込み時に追加サイクルを必要としないため
		addr, _, err := cpu.fetchAddr(opcode.AddressingMode)
		if err != nil {
			return 0, fmt.Errorf("CPU: rol: failed fetch addr for writing result. err: %w", err)
		}

		if err := cpu.memory.Write(addr, result); err != nil {
			return 0, fmt.Errorf("CPU: rol: failed write memory. addr: %x, err: %w", addr, err)
		}
	}

	cpu.incrementPC(uint16(opcode.Length))
	return opcode.Cycles + additionalCycles, nil
}

// ror Rotate Right
func (cpu *CPU) ror(opcode Opcode) (cycles uint8, err error) {
	if opcode.Mnemonic != ROR {
		return 0, fmt.Errorf("invalid mnemonic was specified. mnemonic: %v", opcode.Mnemonic)
	}

	arg, additionalCycles, err := cpu.fetchArg(opcode.AddressingMode)
	if err != nil {
		return 0, fmt.Errorf("CPU: rol: failed fetchArg. err: %w", err)
	}

	result := arg >> 1
	if cpu.getFlag(carryFlag) {
		// carryフラグがある場合、最上位ビットに1を設定
		result |= 0x80
	}

	// 最上位ビット(MSB)が立っているときにシフトしたらcarryする
	cpu.setFlag(carryFlag, arg&0x80 != 0)
	cpu.setFlag(zeroFlag, result == 0)
	cpu.setFlag(negativeFlag, cpu.isNegative(result))

	if opcode.AddressingMode == Accumulator {
		cpu.setA(result)
	} else {
		// メモリ書き込み時に追加サイクルを必要としないため
		addr, _, err := cpu.fetchAddr(opcode.AddressingMode)
		if err != nil {
			return 0, fmt.Errorf("CPU: ror: failed fetch addr for writing result. err: %w", err)
		}

		if err := cpu.memory.Write(addr, result); err != nil {
			return 0, fmt.Errorf("CPU: ror: failed write memory. addr: %x, err: %w", addr, err)
		}
	}

	cpu.incrementPC(uint16(opcode.Length))
	return opcode.Cycles + additionalCycles, nil
}

// rti: REturn from Interrupt
func (cpu *CPU) rti(opcode Opcode) (cycles uint8, err error) {
	if opcode.Mnemonic != RTI {
		return 0, fmt.Errorf("invalid mnemonic was specified. mnemonic: %v", opcode.Mnemonic)
	}

	result, err := cpu.popStack()
	if err != nil {
		return 0, fmt.Errorf("CPU: rti: failed pop stack. err: %w", err)
	}

	// http://wiki.nesdev.com/w/index.php/Status_flags: Pの 4bit 目と 5bit 目は更新しない
	p := (result &^ bFlagMask) | (cpu.register.p & bFlagMask)
	cpu.setP(p)

	// upper -> lowerの順にpopする
	// これはC++の参考と違ったので注意
	upper, err := cpu.popStack()
	if err != nil {
		return 0, fmt.Errorf("CPU: rti: failed pop upper stack. err: %w", err)
	}
	lower, err := cpu.popStack()
	if err != nil {
		return 0, fmt.Errorf("CPU: rti: failed pop lower stack. err: %w", err)
	}

	pc := bit_helper.BytesToUint16(lower, upper)
	cpu.setPC(pc)
	return opcode.Cycles, nil
}

// rts Return from Subroutine
func (cpu *CPU) rts(opcode Opcode) (cycles uint8, err error) {
	if opcode.Mnemonic != RTS {
		return 0, fmt.Errorf("invalid mnemonic was specified. mnemonic: %v", opcode.Mnemonic)
	}

	upper, err := cpu.popStack()
	if err != nil {
		return 0, fmt.Errorf("CPU: rts: failed pop stack. err: %w", err)
	}
	lower, err := cpu.popStack()
	if err != nil {
		return 0, fmt.Errorf("CPU: rts: failed pop stack. err: %w", err)
	}

	pc := bit_helper.BytesToUint16(lower, upper)

	// JSR でスタックにプッシュされるアドレスは JSR の最後のアドレスで、RTS 側でインクリメントされる
	pc++

	cpu.setPC(pc)
	return opcode.Cycles, nil
}

// sbc Subtract with Caryy
func (cpu *CPU) sbc(opcode Opcode) (cycles uint8, err error) {
	if opcode.Mnemonic != SBC {
		return 0, fmt.Errorf("invalid mnemonic was specified. mnemonic: %v", opcode.Mnemonic)
	}

	arg, additionalCycles, err := cpu.fetchArg(opcode.AddressingMode)
	if err != nil {
		return 0, fmt.Errorf("CPU: sbc: failed fetch arg. err: %w", err)
	}

	// 足し算に変換
	// http://www.righto.com/2012/12/the-6502-overflow-flag-explained.html#:~:text=The%20definition%20of%20the%206502,fit%20into%20a%20signed%20byte.&text=For%20each%20set%20of%20input,and%20the%20overflow%20bit%20V.
	// A - arg - borrow == A + ~arg + carry

	arg = ^arg

	isCarryFlag := cpu.getFlag(carryFlag)

	var carry byte
	if isCarryFlag {
		carry = 1
	}

	tmp := uint16(cpu.register.a) + uint16(arg) + uint16(carry)
	result := uint8(tmp & 0xFF)

	cpu.setFlag(overflowFlag, isSignedOverFlowed(cpu.register.a, arg, isCarryFlag))
	cpu.setFlag(carryFlag, tmp > 0xFF)
	cpu.setFlag(negativeFlag, cpu.isNegative(result))
	cpu.setFlag(zeroFlag, result == 0)

	cpu.setA(result)

	cpu.incrementPC(uint16(opcode.Length))
	return opcode.Cycles + additionalCycles, nil
}

func (cpu *CPU) sec(opcode Opcode) (cycles uint8, err error) {
	if opcode.Mnemonic != SEC {
		return 0, fmt.Errorf("invalid mnemonic was specified. mnemonic: %v", opcode.Mnemonic)
	}
	cpu.setFlag(carryFlag, true)
	cpu.incrementPC(uint16(opcode.Length))
	return opcode.Cycles, nil
}

func (cpu *CPU) sed(opcode Opcode) (cycles uint8, err error) {
	if opcode.Mnemonic != SED {
		return 0, fmt.Errorf("invalid mnemonic was specified. mnemonic: %v", opcode.Mnemonic)
	}

	cpu.setFlag(decimalFlag, true)
	cpu.incrementPC(uint16(opcode.Length))
	return opcode.Cycles, nil
}

func (cpu *CPU) sei(opcode Opcode) (cycles uint8, err error) {
	if opcode.Mnemonic != SEI {
		return 0, fmt.Errorf("invalid mnemonic was specified. mnemonic: %v", opcode.Mnemonic)
	}
	cpu.setFlag(interruptFlag, true)
	cpu.incrementPC(uint16(opcode.Length))
	return opcode.Cycles, nil
}

// sta Store A
func (cpu *CPU) sta(opcode Opcode) (cycles uint8, err error) {
	if opcode.Mnemonic != STA {
		return 0, fmt.Errorf("invalid mnemonic was specified. mnemonic: %v", opcode.Mnemonic)
	}

	addr, additionalCycles, err := cpu.fetchAddr(opcode.AddressingMode)
	if err != nil {
		return 0, fmt.Errorf("CPU: sta: failed fetch addr. err: %w", err)
	}

	if err := cpu.memory.Write(addr, cpu.register.a); err != nil {
		return 0, fmt.Errorf("CPU: sta: failed write memory. err: %w", err)
	}
	cpu.incrementPC(uint16(opcode.Length))
	return opcode.Cycles + additionalCycles, nil
}

func (cpu *CPU) stx(opcode Opcode) (cycles uint8, err error) {
	if opcode.Mnemonic != STX {
		return 0, fmt.Errorf("invalid mnemonic was specified. mnemonic: %v", opcode.Mnemonic)
	}

	addr, additionalCycles, err := cpu.fetchAddr(opcode.AddressingMode)
	if err != nil {
		return 0, fmt.Errorf("CPU: stx: failed fetch addr. err: %w", err)
	}

	if err := cpu.memory.Write(addr, cpu.register.x); err != nil {
		return 0, fmt.Errorf("CPU: stx: failed write memory. err: %w", err)
	}
	cpu.incrementPC(uint16(opcode.Length))
	return opcode.Cycles + additionalCycles, nil
}

func (cpu *CPU) sty(opcode Opcode) (cycles uint8, err error) {
	if opcode.Mnemonic != STY {
		return 0, fmt.Errorf("invalid mnemonic was specified. mnemonic: %v", opcode.Mnemonic)
	}

	addr, additionalCycles, err := cpu.fetchAddr(opcode.AddressingMode)
	if err != nil {
		return 0, fmt.Errorf("CPU: sty: failed fetch addr. err: %w", err)
	}

	if err := cpu.memory.Write(addr, cpu.register.y); err != nil {
		return 0, fmt.Errorf("CPU: sty: failed write memory. err: %w", err)
	}
	cpu.incrementPC(uint16(opcode.Length))
	return opcode.Cycles + additionalCycles, nil
}

// tax Transfer A to X
func (cpu *CPU) tax(opcode Opcode) (cycles uint8, err error) {
	if opcode.Mnemonic != TAX {
		return 0, fmt.Errorf("invalid mnemonic was specified. mnemonic: %v", opcode.Mnemonic)
	}

	cpu.setFlag(zeroFlag, cpu.register.a == 0)
	cpu.setFlag(negativeFlag, cpu.isNegative(cpu.register.a))

	cpu.setX(cpu.register.a)

	cpu.incrementPC(uint16(opcode.Length))
	return opcode.Cycles, nil
}

func (cpu *CPU) tay(opcode Opcode) (cycles uint8, err error) {
	if opcode.Mnemonic != TAY {
		return 0, fmt.Errorf("invalid mnemonic was specified. mnemonic: %v", opcode.Mnemonic)
	}

	cpu.setFlag(zeroFlag, cpu.register.a == 0)
	cpu.setFlag(negativeFlag, cpu.isNegative(cpu.register.a))

	cpu.setY(cpu.register.a)

	cpu.incrementPC(uint16(opcode.Length))
	return opcode.Cycles, nil
}

func (cpu *CPU) tsx(opcode Opcode) (cycles uint8, err error) {
	if opcode.Mnemonic != TSX {
		return 0, fmt.Errorf("invalid mnemonic was specified. mnemonic: %v", opcode.Mnemonic)
	}

	cpu.setFlag(zeroFlag, cpu.register.sp == 0)
	cpu.setFlag(negativeFlag, cpu.isNegative(cpu.register.sp))
	cpu.setX(cpu.register.sp)
	cpu.incrementPC(uint16(opcode.Length))
	return opcode.Cycles, nil
}

func (cpu *CPU) txa(opcode Opcode) (cycles uint8, err error) {
	if opcode.Mnemonic != TXA {
		return 0, fmt.Errorf("invalid mnemonic was specified. mnemonic: %v", opcode.Mnemonic)
	}

	cpu.setFlag(zeroFlag, cpu.register.x == 0)
	cpu.setFlag(negativeFlag, cpu.isNegative(cpu.register.x))
	cpu.setA(cpu.register.x)

	cpu.incrementPC(uint16(opcode.Length))
	return opcode.Cycles, nil
}

func (cpu *CPU) txs(opcode Opcode) (cycles uint8, err error) {
	if opcode.Mnemonic != TXS {
		return 0, fmt.Errorf("invalid mnemonic was specified. mnemonic: %v", opcode.Mnemonic)
	}
	cpu.setSP(cpu.register.x)
	cpu.incrementPC(uint16(opcode.Length))
	return opcode.Cycles, nil
}

func (cpu *CPU) tya(opcode Opcode) (cycles uint8, err error) {
	if opcode.Mnemonic != TYA {
		return 0, fmt.Errorf("invalid mnemonic was specified. mnemonic: %v", opcode.Mnemonic)
	}

	cpu.setFlag(zeroFlag, cpu.register.y == 0)
	cpu.setFlag(negativeFlag, cpu.isNegative(cpu.register.y))

	cpu.setA(cpu.register.y)

	cpu.incrementPC(uint16(opcode.Length))
	return opcode.Cycles, nil
}

// TODO 拡張命令の実装

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

func (cpu *CPU) setP(p byte) {
	cpu.register.p = p
}

func (cpu *CPU) setX(x byte) {
	cpu.register.x = x
}

func (cpu *CPU) setY(y byte) {
	cpu.register.y = y
}

func (cpu *CPU) setSP(sp byte) {
	cpu.register.sp = sp
}

func (cpu *CPU) isNegative(b byte) bool {
	return b&0x80 != 0
}

// pushStack pushes a byte onto the stack.
// NES stack resides in page 2 (0x0100 - 0x01FF).
func (cpu *CPU) pushStack(b byte) error {
	spAddr := uint16(cpu.register.sp) | uint16(0x0100)

	if err := cpu.memory.Write(spAddr, b); err != nil {
		return fmt.Errorf("CPU: pushStack: failed write memory. addr: %x, err: %w", spAddr, err)
	}
	cpu.register.sp -= 1
	return nil
}

func (cpu *CPU) popStack() (byte, error) {
	spAddr := uint16(cpu.register.sp) | uint16(0x0100)
	b, err := cpu.memory.Read(spAddr)
	if err != nil {
		return 0, fmt.Errorf("CPU: popStack: failed read memory. addr: %x, err: %w", spAddr, err)
	}
	cpu.register.sp += 1
	return b, nil
}

// isSignedOverFlowed 符号付きの計算でオーバーフローしているかどうかを判定できる
// 引数のnとmはuint8なのに注意すること
func isSignedOverFlowed(n byte, m byte, carryFlag bool) bool {
	var carry byte
	if carryFlag {
		carry = 1
	}

	result := n + m + carry
	return ((m ^ result) & (n ^ result) & 0x80) == 0x80
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
