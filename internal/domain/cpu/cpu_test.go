package cpu

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/sunjin110/nes_emu/internal/domain/prgrom"
	"github.com/sunjin110/nes_emu/pkg/bit_helper"
)

type dummyMemory struct {
	data map[uint16]byte
}

func (m *dummyMemory) Read(addr uint16) (byte, error) {
	return m.data[addr], nil
}

func (m *dummyMemory) Write(addr uint16, value byte) error {
	m.data[addr] = value
	return nil
}

func (m *dummyMemory) GetPRGROM() prgrom.PRGROM {
	return &dummyPRGROM{
		data: m.data,
	}
}

type dummyPRGROM struct {
	data map[uint16]byte
}

func (rom *dummyPRGROM) Read(addr uint16) byte {
	return rom.data[addr]
}

func (rom *dummyPRGROM) InitPC() uint16 {
	pcLower := rom.data[0xFFFC]
	pcUpper := rom.data[0xFFFD]
	pc := bit_helper.BytesToUint16(pcLower, pcUpper)
	return pc
}

// go test -v -count=1 -timeout 30s -run ^Test_CPU_Run$ github.com/sunjin110/nes_emu/internal/domain/cpu
func Test_CPU_Run(t *testing.T) {
	Convey("Test_CPU_Run", t, func() {
		type test struct {
			name           string
			initialMemory  map[uint16]byte
			initalRegs     Register
			expectedMemory map[uint16]byte
			expectedRegs   Register
			expectedCycles uint8
		}

		tests := []test{
			{
				name: "ADC Immediate - A = A + 5",
				initialMemory: map[uint16]byte{
					0x8000: 0x69, // ADC Immediate opcode
					0x8001: 0x05, // Operand
				},
				initalRegs: Register{
					a:  0x03,   // A = 3
					pc: 0x8000, // Program Counter starts here
					p:  0x00,   // Processor flags clear
				},
				expectedMemory: map[uint16]byte{}, // No memory changes
				expectedRegs: Register{
					a:  0x08,   // A = 3 + 5
					pc: 0x8002, // Program Counter after instruction
					p:  0x00,   // Flags remain unchanged
				},
				expectedCycles: 2, // ADC Immediate takes 2 cycles
			},
			{
				name: "ADC Zeropage - A = A + M[0x10]",
				initialMemory: map[uint16]byte{
					0x8000: 0x65, // ADC Zeropage opcode
					0x8001: 0x10, // Address
					0x0010: 0x05, // Value at address 0x0010
				},
				initalRegs: Register{
					a:  0x03,   // A = 3
					pc: 0x8000, // Program Counter starts here
					p:  0x00,   // Processor flags clear
				},
				expectedMemory: map[uint16]byte{}, // No memory changes
				expectedRegs: Register{
					a:  0x08,   // A = 3 + 5
					pc: 0x8002, // Program Counter after instruction
					p:  0x00,   // Flags remain unchanged
				},
				expectedCycles: 3, // ADC Zeropage takes 3 cycles
			},
			{
				name: "ADC ZeropageX - A = A + M[0x10 + X]",
				initialMemory: map[uint16]byte{
					0x8000: 0x75, // ADC ZeropageX opcode
					0x8001: 0x10, // Base address
					0x0015: 0x07, // Value at address 0x0010 + X (X = 5)
				},
				initalRegs: Register{
					a:  0x03,   // A = 3
					x:  0x05,   // X = 5
					pc: 0x8000, // Program Counter starts here
					p:  0x00,   // Processor flags clear
				},
				expectedMemory: map[uint16]byte{}, // No memory changes
				expectedRegs: Register{
					a:  0x0A, // A = 3 + 7
					x:  0x05,
					pc: 0x8002, // Program Counter after instruction
					p:  0x00,   // Flags remain unchanged
				},
				expectedCycles: 4, // ADC ZeropageX takes 4 cycles
			},
			{
				name: "ADC Absolute - A = A + M[0x1000]",
				initialMemory: map[uint16]byte{
					0x8000: 0x6D, // ADC Absolute opcode
					0x8001: 0x00, // Lower byte of address
					0x8002: 0x10, // Upper byte of address
					0x1000: 0x0A, // Value at address 0x1000
				},
				initalRegs: Register{
					a:  0x05,   // A = 5
					pc: 0x8000, // Program Counter starts here
					p:  0x00,   // Processor flags clear
				},
				expectedMemory: map[uint16]byte{}, // No memory changes
				expectedRegs: Register{
					a:  0x0F,   // A = 5 + 10
					pc: 0x8003, // Program Counter after instruction
					p:  0x00,   // Flags remain unchanged
				},
				expectedCycles: 4, // ADC Absolute takes 4 cycles
			},
			{
				name: "ADC AbsoluteX - A = A + M[0x1000 + X]",
				initialMemory: map[uint16]byte{
					0x8000: 0x7D, // ADC AbsoluteX opcode
					0x8001: 0x00, // Lower byte of address
					0x8002: 0x10, // Upper byte of address
					0x1005: 0x05, // Value at address 0x1000 + X (X = 5)
				},
				initalRegs: Register{
					a:  0x03,   // A = 3
					x:  0x05,   // X = 5
					pc: 0x8000, // Program Counter starts here
					p:  0x00,   // Processor flags clear
				},
				expectedMemory: map[uint16]byte{}, // No memory changes
				expectedRegs: Register{
					a:  0x08,   // A = 3 + 5
					x:  0x05,   // X remains unchanged
					pc: 0x8003, // Program Counter after instruction
					p:  0x00,   // Flags remain unchanged
				},
				expectedCycles: 4, // ADC AbsoluteX takes 4 cycles
			},
			{
				name: "ADC AbsoluteY - A = A + M[0x1000 + Y]",
				initialMemory: map[uint16]byte{
					0x8000: 0x79, // ADC AbsoluteY opcode
					0x8001: 0x00, // Lower byte of address
					0x8002: 0x10, // Upper byte of address
					0x1007: 0x07, // Value at address 0x1000 + Y (Y = 7)
				},
				initalRegs: Register{
					a:  0x03,   // A = 3
					y:  0x07,   // Y = 7
					pc: 0x8000, // Program Counter starts here
					p:  0x00,   // Processor flags clear
				},
				expectedMemory: map[uint16]byte{}, // No memory changes
				expectedRegs: Register{
					a:  0x0A,   // A = 3 + 7
					y:  0x07,   // Y remains unchanged
					pc: 0x8003, // Program Counter after instruction
					p:  0x00,   // Flags remain unchanged
				},
				expectedCycles: 4, // ADC AbsoluteY takes 4 cycles
			},
			{
				name: "ADC IndirectX - A = A + M[*(0x10 + X)]",
				initialMemory: map[uint16]byte{
					0x8000: 0x61, // ADC IndirectX opcode
					0x8001: 0x10, // Base address
					0x0015: 0x00, // Indirect lower byte
					0x0016: 0x10, // Indirect upper byte
					0x1000: 0x09, // Value at address *(0x10 + X)
				},
				initalRegs: Register{
					a:  0x03,   // A = 3
					x:  0x05,   // X = 5
					pc: 0x8000, // Program Counter starts here
					p:  0x00,   // Processor flags clear
				},
				expectedMemory: map[uint16]byte{}, // No memory changes
				expectedRegs: Register{
					a:  0x0C,   // A = 3 + 9
					x:  0x05,   // X remains unchanged
					pc: 0x8002, // Program Counter after instruction
					p:  0x00,   // Flags remain unchanged
				},
				expectedCycles: 6, // ADC IndirectX takes 6 cycles
			},
			{
				name: "ADC IndirectY - A = A + M[*(0x10) + Y]",
				initialMemory: map[uint16]byte{
					0x8000: 0x71, // ADC IndirectY opcode
					0x8001: 0x10, // Base address
					0x0010: 0x00, // Indirect lower byte
					0x0011: 0x10, // Indirect upper byte
					0x1005: 0x04, // Value at address *(0x10) + Y (Y = 5)
				},
				initalRegs: Register{
					a:  0x02,   // A = 2
					y:  0x05,   // Y = 5
					pc: 0x8000, // Program Counter starts here
					p:  0x00,   // Processor flags clear
				},
				expectedMemory: map[uint16]byte{}, // No memory changes
				expectedRegs: Register{
					a:  0x06,   // A = 2 + 4
					y:  0x05,   // Y remains unchanged
					pc: 0x8002, // Program Counter after instruction
					p:  0x00,   // Flags remain unchanged
				},
				expectedCycles: 5, // ADC IndirectY takes 5 cycles
			},
			{
				name: "AND Immediate - A = A & Operand",
				initialMemory: map[uint16]byte{
					0x8000: 0x29, // AND Immediate opcode
					0x8001: 0x0F, // Operand
				},
				initalRegs: Register{
					a:  0x3F,   // A = 0b00111111
					pc: 0x8000, // Program Counter starts here
					p:  0x00,   // Processor flags clear
				},
				expectedMemory: map[uint16]byte{}, // No memory changes
				expectedRegs: Register{
					a:  0x0F,   // A = 0b00111111 & 0b00001111
					pc: 0x8002, // Program Counter after instruction
					p:  0x00,   // Flags remain unchanged
				},
				expectedCycles: 2, // AND Immediate takes 2 cycles
			},
			{
				name: "AND Zeropage - A = A & M[0x10]",
				initialMemory: map[uint16]byte{
					0x8000: 0x25, // AND Zeropage opcode
					0x8001: 0x10, // Address
					0x0010: 0x0F, // Value at address 0x10
				},
				initalRegs: Register{
					a:  0x3F,   // A = 0b00111111
					pc: 0x8000, // Program Counter starts here
					p:  0x00,   // Processor flags clear
				},
				expectedMemory: map[uint16]byte{}, // No memory changes
				expectedRegs: Register{
					a:  0x0F,   // A = 0b00111111 & 0b00001111
					pc: 0x8002, // Program Counter after instruction
					p:  0x00,   // Flags remain unchanged
				},
				expectedCycles: 3, // AND Zeropage takes 3 cycles
			},
			{
				name: "AND ZeropageX - A = A & M[0x10 + X]",
				initialMemory: map[uint16]byte{
					0x8000: 0x35, // AND ZeropageX opcode
					0x8001: 0x10, // Base address
					0x0015: 0x0F, // Value at address 0x10 + X (X = 5)
				},
				initalRegs: Register{
					a:  0x3F,   // A = 0b00111111
					x:  0x05,   // X = 5
					pc: 0x8000, // Program Counter starts here
					p:  0x00,   // Processor flags clear
				},
				expectedMemory: map[uint16]byte{}, // No memory changes
				expectedRegs: Register{
					a:  0x0F,   // A = 0b00111111 & 0b00001111
					x:  0x05,   // X remains unchanged
					pc: 0x8002, // Program Counter after instruction
					p:  0x00,   // Flags remain unchanged
				},
				expectedCycles: 4, // AND ZeropageX takes 4 cycles
			},
			{
				name: "AND Absolute - A = A & M[0x1000]",
				initialMemory: map[uint16]byte{
					0x8000: 0x2D, // AND Absolute opcode
					0x8001: 0x00, // Lower byte of address
					0x8002: 0x10, // Upper byte of address
					0x1000: 0x0F, // Value at address 0x1000
				},
				initalRegs: Register{
					a:  0x3F,   // A = 0b00111111
					pc: 0x8000, // Program Counter starts here
					p:  0x00,   // Processor flags clear
				},
				expectedMemory: map[uint16]byte{}, // No memory changes
				expectedRegs: Register{
					a:  0x0F,   // A = 0b00111111 & 0b00001111
					pc: 0x8003, // Program Counter after instruction
					p:  0x00,   // Flags remain unchanged
				},
				expectedCycles: 4, // AND Absolute takes 4 cycles
			},
			{
				name: "AND AbsoluteX - A = A & M[0x1000 + X]",
				initialMemory: map[uint16]byte{
					0x8000: 0x3D, // AND AbsoluteX opcode
					0x8001: 0x00, // Lower byte of address
					0x8002: 0x10, // Upper byte of address
					0x1005: 0x0F, // Value at address 0x1000 + X (X = 5)
				},
				initalRegs: Register{
					a:  0x3F,   // A = 0b00111111
					x:  0x05,   // X = 5
					pc: 0x8000, // Program Counter starts here
					p:  0x00,   // Processor flags clear
				},
				expectedMemory: map[uint16]byte{}, // No memory changes
				expectedRegs: Register{
					a:  0x0F,   // A = 0b00111111 & 0b00001111
					x:  0x05,   // X remains unchanged
					pc: 0x8003, // Program Counter after instruction
					p:  0x00,   // Flags remain unchanged
				},
				expectedCycles: 4, // AND AbsoluteX takes 4 cycles
			},
			{
				name: "AND AbsoluteY - A = A & M[0x1000 + Y]",
				initialMemory: map[uint16]byte{
					0x8000: 0x39, // AND AbsoluteY opcode
					0x8001: 0x00, // Lower byte of address
					0x8002: 0x10, // Upper byte of address
					0x1007: 0x0F, // Value at address 0x1000 + Y (Y = 7)
				},
				initalRegs: Register{
					a:  0x3F,   // A = 0b00111111
					y:  0x07,   // Y = 7
					pc: 0x8000, // Program Counter starts here
					p:  0x00,   // Processor flags clear
				},
				expectedMemory: map[uint16]byte{}, // No memory changes
				expectedRegs: Register{
					a:  0x0F,   // A = 0b00111111 & 0b00001111
					y:  0x07,   // Y remains unchanged
					pc: 0x8003, // Program Counter after instruction
					p:  0x00,   // Flags remain unchanged
				},
				expectedCycles: 4, // AND AbsoluteY takes 4 cycles
			},
			{
				name: "AND IndirectX - A = A & M[*(0x10 + X)]",
				initialMemory: map[uint16]byte{
					0x8000: 0x21, // AND IndirectX opcode
					0x8001: 0x10, // Base address
					0x0015: 0x00, // Indirect lower byte
					0x0016: 0x10, // Indirect upper byte
					0x1000: 0x0F, // Value at address *(0x10 + X)
				},
				initalRegs: Register{
					a:  0x3F,   // A = 0b00111111
					x:  0x05,   // X = 5
					pc: 0x8000, // Program Counter starts here
					p:  0x00,   // Processor flags clear
				},
				expectedMemory: map[uint16]byte{}, // No memory changes
				expectedRegs: Register{
					a:  0x0F,   // A = 0b00111111 & 0b00001111
					x:  0x05,   // X remains unchanged
					pc: 0x8002, // Program Counter after instruction
					p:  0x00,   // Flags remain unchanged
				},
				expectedCycles: 6, // AND IndirectX takes 6 cycles
			},
			{
				name: "AND IndirectY - A = A & M[*(0x10) + Y]",
				initialMemory: map[uint16]byte{
					0x8000: 0x31, // AND IndirectY opcode
					0x8001: 0x10, // Base address
					0x0010: 0x00, // Indirect lower byte
					0x0011: 0x10, // Indirect upper byte
					0x1005: 0x0F, // Value at address *(0x10) + Y (Y = 5)
				},
				initalRegs: Register{
					a:  0x3F,   // A = 0b00111111
					y:  0x05,   // Y = 5
					pc: 0x8000, // Program Counter starts here
					p:  0x00,   // Processor flags clear
				},
				expectedMemory: map[uint16]byte{}, // No memory changes
				expectedRegs: Register{
					a:  0x0F,   // A = 0b00111111 & 0b00001111
					y:  0x05,   // Y remains unchanged
					pc: 0x8002, // Program Counter after instruction
					p:  0x00,   // Flags remain unchanged
				},
				expectedCycles: 5, // AND IndirectY takes 5 cycles
			},
			{
				name: "ASL Accumulator - A = A << 1",
				initialMemory: map[uint16]byte{
					0x8000: 0x0A, // ASL Accumulator opcode
				},
				initalRegs: Register{
					a:  0x40,   // A = 0b01000000
					pc: 0x8000, // Program Counter starts here
					p:  0x00,   // Processor flags clear
				},
				expectedMemory: map[uint16]byte{}, // No memory changes
				expectedRegs: Register{
					a:  0x80,   // A = 0b10000000
					pc: 0x8001, // Program Counter after instruction
					p:  0x80,   // Negative flag set
				},
				expectedCycles: 2, // ASL Accumulator takes 2 cycles
			},
			{
				name: "BCC - Branch if Carry Clear",
				initialMemory: map[uint16]byte{
					0x8000: 0x90, // BCC opcode
					0x8001: 0x02, // Offset = 2
				},
				initalRegs: Register{
					pc: 0x8000, // Program Counter starts here
					p:  0x00,   // Carry flag clear
				},
				expectedMemory: map[uint16]byte{}, // No memory changes
				expectedRegs: Register{
					pc: 0x8004, // Program Counter jumps by offset (0x8002 + 2)
					p:  0x00,   // Flags remain unchanged
				},
				expectedCycles: 3, // BCC takes 2 cycles + 1 for branch taken
			},
			{
				name: "BCS - Branch if Carry Set, no page crossing",
				initialMemory: map[uint16]byte{
					0x8000: 0xB0, // BCS opcode
					0x8001: 0x02, // Offset = +2
				},
				initalRegs: Register{
					pc: 0x8000, // Program Counter starts here
					p:  0x01,   // Carry flag set
				},
				expectedMemory: map[uint16]byte{}, // No memory changes
				expectedRegs: Register{
					pc: 0x8004, // Program Counter jumps by offset (0x8002 + 2)
					p:  0x01,   // Flags remain unchanged
				},
				expectedCycles: 3, // No page crossing, so +1 cycle
			},
			{
				name: "BCS - Branch if Carry Set, page crossing",
				initialMemory: map[uint16]byte{
					0x80FD: 0xB0, // BCS opcode
					0x80FE: 0x03, // Offset = +3
				},
				initalRegs: Register{
					pc: 0x80FD, // Program Counter starts here
					p:  0x01,   // Carry flag set
				},
				expectedMemory: map[uint16]byte{}, // No memory changes
				expectedRegs: Register{
					pc: 0x8102, // Program Counter jumps by offset (0x8100 + 1)
					p:  0x01,   // Flags remain unchanged
				},
				expectedCycles: 4, // Page crossing, so +2 cycles
			},
			{
				name: "BEQ - Branch if Equal",
				initialMemory: map[uint16]byte{
					0x8000: 0xF0, // BEQ opcode
					0x8001: 0x05, // Offset = 5
				},
				initalRegs: Register{
					pc: 0x8000, // Program Counter starts here
					p:  0x02,   // Zero flag set
				},
				expectedMemory: map[uint16]byte{}, // No memory changes
				expectedRegs: Register{
					pc: 0x8007, // Program Counter jumps by offset (0x8002 + 5)
					p:  0x02,   // Flags remain unchanged
				},
				expectedCycles: 3, // BEQ takes 2 cycles + 1 for branch taken
			},
			{
				name: "BIT Zeropage",
				initialMemory: map[uint16]byte{
					0x8000: 0x24, // BIT Zeropage opcode
					0x8001: 0x10, // Address
					0x0010: 0x80, // Value at address 0x0010
				},
				initalRegs: Register{
					a:  0x80,
					pc: 0x8000,
					p:  0x00,
				},
				expectedMemory: map[uint16]byte{},
				expectedRegs: Register{
					a:  0x80,
					pc: 0x8002,
					p:  0x80, // Negative flag set
				},
				expectedCycles: 3,
			},
			{
				name: "BIT Absolute",
				initialMemory: map[uint16]byte{
					0x8000: 0x2C, // BIT Absolute opcode
					0x8001: 0x10, // Low byte of address
					0x8002: 0x00, // High byte of address
					0x0010: 0xC0, // Value at address 0x0010
				},
				initalRegs: Register{
					a:  0xC0,
					pc: 0x8000,
					p:  0x00,
				},
				expectedMemory: map[uint16]byte{},
				expectedRegs: Register{
					a:  0xC0,
					pc: 0x8003,
					p:  0xC0, // Negative and Overflow flags set
				},
				expectedCycles: 4,
			},
			{
				name: "BMI - Branch if Minus",
				initialMemory: map[uint16]byte{
					0x8000: 0x30, // BMI opcode
					0x8001: 0x02, // Offset
				},
				initalRegs: Register{
					pc: 0x8000,
					p:  0x80, // Negative flag set
				},
				expectedMemory: map[uint16]byte{},
				expectedRegs: Register{
					pc: 0x8004, // Branch taken
					p:  0x80,
				},
				expectedCycles: 3,
			},
			{
				name: "BNE - Branch if Not Equal",
				initialMemory: map[uint16]byte{
					0x8000: 0xD0, // BNE opcode
					0x8001: 0x02, // Offset
				},
				initalRegs: Register{
					pc: 0x8000,
					p:  0x00, // Zero flag not set
				},
				expectedMemory: map[uint16]byte{},
				expectedRegs: Register{
					pc: 0x8004, // Branch taken
					p:  0x00,
				},
				expectedCycles: 3,
			},
			{
				name: "BPL - Branch if Positive",
				initialMemory: map[uint16]byte{
					0x8000: 0x10, // BPL opcode
					0x8001: 0x02, // Offset
				},
				initalRegs: Register{
					pc: 0x8000,
					p:  0x00, // Negative flag not set
				},
				expectedMemory: map[uint16]byte{},
				expectedRegs: Register{
					pc: 0x8004, // Branch taken
					p:  0x00,
				},
				expectedCycles: 3,
			},
			{
				name: "BRK - Force Interrupt",
				initialMemory: map[uint16]byte{
					0x8000: 0x00, // BRK opcode
					0xFFFE: 0x10, // 下
					0xFFFF: 0x01, // 上
				},
				initalRegs: Register{
					pc: 0x8000,
					sp: 0xFD, // Stack pointer
				},
				expectedMemory: map[uint16]byte{
					0x01FD: 0x01, // PC low byte pushed
					0x01FC: 0x80, // PC high byte pushed
					0x01FB: 0b110100,
				},
				expectedRegs: Register{
					pc: 0x0110, // Jump to IRQ/BRK vector
					sp: 0xFA,   // Stack pointer after pushes
					p:  0b00010100,
				},
				expectedCycles: 7,
			},
			{
				name: "BVC - Branch if Overflow Clear",
				initialMemory: map[uint16]byte{
					0x8000: 0x50, // BVC opcode
					0x8001: 0x02, // Offset
				},
				initalRegs: Register{
					pc: 0x8000,
					p:  0x00, // Overflow flag not set
				},
				expectedMemory: map[uint16]byte{},
				expectedRegs: Register{
					pc: 0x8004, // Branch taken
					p:  0x00,
				},
				expectedCycles: 3,
			},
			{
				name: "BVS - Branch if Overflow Set",
				initialMemory: map[uint16]byte{
					0x8000: 0x70, // BVS opcode
					0x8001: 0x02, // Offset
				},
				initalRegs: Register{
					pc: 0x8000,
					p:  0x40, // Overflow flag set
				},
				expectedMemory: map[uint16]byte{},
				expectedRegs: Register{
					pc: 0x8004, // Branch taken
					p:  0x40,
				},
				expectedCycles: 3,
			},
		}

		for _, tt := range tests {
			Convey(tt.name, func() {

				m := &dummyMemory{
					data: tt.initialMemory,
				}

				cpu, err := NewCPU(m, m.GetPRGROM())
				So(err, ShouldBeNil)

				cpu.register = tt.initalRegs

				cycles, err := cpu.Run()
				So(err, ShouldBeNil)
				So(cycles, ShouldEqual, tt.expectedCycles)

				// メモリ検証
				for addr, expected := range tt.expectedMemory {
					actual, _ := cpu.memory.Read(addr)

					if actual != expected {
						fmt.Printf("======= addr: %xが一致しません. expected: %x, actual: %x", addr, expected, actual)
					}

					So(actual, ShouldEqual, expected)
				}
				So(cpu.register, ShouldResemble, tt.expectedRegs)
			})
		}
	})
}
