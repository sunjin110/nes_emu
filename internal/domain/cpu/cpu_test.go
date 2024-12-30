package cpu

import "testing"

func Test_CPU_abc(t *testing.T) {
	tests := []struct {
		name             string
		a                byte
		operand          byte
		carryIn          bool
		expectedA        byte
		expectedCarry    bool
		expectedZero     bool
		expectedNegative bool
	}{
		{
			name:             "Simple addition without carry",
			a:                0x10,
			operand:          0x20,
			carryIn:          false,
			expectedA:        0x30,
			expectedCarry:    false,
			expectedZero:     false,
			expectedNegative: false,
		},
		{
			name:             "Addition with carry",
			a:                0xFF,
			operand:          0x01,
			carryIn:          false,
			expectedA:        0x00,
			expectedCarry:    true,
			expectedZero:     true,
			expectedNegative: false,
		},
		{
			name:             "Addition with carry flag input",
			a:                0x10,
			operand:          0x20,
			carryIn:          true,
			expectedA:        0x31,
			expectedCarry:    false,
			expectedZero:     false,
			expectedNegative: false,
		},
		{
			name:             "Negative result",
			a:                0x50,
			operand:          0xB0,
			carryIn:          false,
			expectedA:        0x00,
			expectedCarry:    true,
			expectedZero:     true,
			expectedNegative: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cpu := &CPU{
				register: Register{
					a: tt.a,
					p: 0,
				},
			}

			if tt.carryIn {
				cpu.setFlag(carryFlag, true)
			}

			cpu.adc(tt.operand)

			if cpu.register.a != tt.expectedA {
				t.Errorf("got A = 0x%02X, want 0x%02X", cpu.register.a, tt.expectedA)
			}

			if cpu.getFlag(carryFlag) != tt.expectedCarry {
				t.Errorf("got Carry = %v, want %v", cpu.getFlag(carryFlag), tt.expectedCarry)
			}

			if cpu.getFlag(zeroFlag) != tt.expectedZero {
				t.Errorf("got Zero = %v, want %v", cpu.getFlag(zeroFlag), tt.expectedZero)
			}

			if cpu.getFlag(negativeFlag) != tt.expectedNegative {
				t.Errorf("got Negative = %v, want %v", cpu.getFlag(negativeFlag), tt.expectedNegative)
			}
		})
	}
}
