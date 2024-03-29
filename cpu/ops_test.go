package cpu

import (
	"testing"
)

func TestADC(t *testing.T) {
	tests := testCases{
		{
			name:        "add with carry",
			program:     []uint8{0x69, 0x02},
			setupA:      newUint8(0xff),
			expectA:     newUint8(0x01),
			expectCarry: true,
		},
		{
			name:        "add to zero",
			program:     []uint8{0x69, 0x02},
			setupA:      newUint8(0xfe),
			expectA:     newUint8(0x00),
			expectCarry: true,
			expectZero:  true,
		},
		{
			name:           "127 + 1 = 128, returns V = 1",
			program:        []uint8{0x69, 0x01},
			setupA:         newUint8(0x7f),
			expectOverflow: true,
			expectNegative: true,
		},
		{
			name:    "adds two positive numbers without carry",
			program: []uint8{0x69, 0x0f},
			expectA: newUint8(0x1f),
			setupA:  newUint8(0x10),
		},
		{
			name:    "immediate without carry",
			program: []uint8{0x69, 0x42},
			expectA: newUint8(0x43),
			setupA:  newUint8(0x01),
		},
		{
			name:           "zeropage without carry",
			program:        []uint8{0x65, 0x42},
			memory:         map[uint16]uint8{0x42: 0x80},
			expectA:        newUint8(0x81),
			setupA:         newUint8(0x01),
			expectNegative: true,
		},
		{
			name:           "absolute without carry",
			program:        []uint8{0x6d, 0x00, 0x04},
			memory:         map[uint16]uint8{0x0400: 0x42},
			expectA:        newUint8(0x43),
			expectCarry:    false,
			expectOverflow: false,
			expectNegative: false,
			setupA:         newUint8(0x01),
		},
	}
	tests.run(t)
}

func TestAND(t *testing.T) {
	tests := testCases{
		{
			name:           "immediate",
			program:        []uint8{0x29, 0xAA},
			expectA:        newUint8(0xAA),
			expectNegative: true,
			setupA:         newUint8(0xFF),
		},
		{
			name:    "zeropage",
			program: []uint8{0x25, 0x42},
			memory:  map[uint16]uint8{0x42: 0x0F},
			expectA: newUint8(0x0E),
			setupA:  newUint8(0xDE),
		},
		{
			name:           "absolute",
			program:        []uint8{0x2D, 0x00, 0x04},
			memory:         map[uint16]uint8{0x0400: 0xF0},
			expectA:        newUint8(0xC0),
			expectNegative: true,
			setupA:         newUint8(0xC0),
		},
	}
	tests.run(t)
}

func TestASL(t *testing.T) {
	tests := testCases{
		{
			name:        "accumulator",
			program:     []uint8{0x0a},
			expectA:     newUint8(0x54),
			expectCarry: true,
		},
		{
			name:       "accumulator 0",
			program:    []uint8{0x0a},
			setupA:     newUint8(0x00),
			expectA:    newUint8(0x00),
			expectZero: true,
		},
		{
			name:           "zeropage",
			program:        []uint8{0x06, 0x42},
			memory:         map[uint16]uint8{0x0042: 0x55},
			expectMemory:   map[uint16]uint8{0x0042: 0xaa},
			expectNegative: true,
		},
		{
			name:           "zeropage,x",
			program:        []uint8{0x16, 0x42},
			memory:         map[uint16]uint8{0x0047: 0x55},
			expectMemory:   map[uint16]uint8{0x0047: 0xaa},
			expectNegative: true,
			setupX:         newUint8(0x5),
		},
		{
			name:           "absolute",
			program:        []uint8{0x0e, 0x42},
			memory:         map[uint16]uint8{0x0042: 0x55},
			expectMemory:   map[uint16]uint8{0x0042: 0xaa},
			expectNegative: true,
		},
		{
			name:           "absolute,x",
			program:        []uint8{0x1e, 0x42},
			memory:         map[uint16]uint8{0x0047: 0x55},
			expectMemory:   map[uint16]uint8{0x0047: 0xaa},
			expectNegative: true,
			setupX:         newUint8(0x5),
		},
	}
	tests.run(t)
}

func TestBCC(t *testing.T) {
	tests := testCases{
		{
			// Test BCC with carry flag set; the branch should not be taken
			name:        "no branch",
			program:     []uint8{0x90, 0x10},
			setupCarry:  newBool(true),
			expectCarry: true,
			expectPC:    newUint16(ProgramStart + 0x02),
		},
		{
			// Test BCC with carry flag clear; the branch should be taken
			name:        "branch",
			program:     []uint8{0x90, 0x10},
			expectCarry: false,
			expectPC:    newUint16(ProgramStart + 0x02 + 0x10),
		},
		{
			// Test BCC with carry flag clear and crossing a page boundary
			// The branch should be taken, and an extra cycle should be consumed
			name:        "branch with page boundary crossing",
			program:     []uint8{0x90, 0xF6},
			expectCarry: false,
			expectPC:    newUint16(ProgramStart - 0x8),
		},
	}
	tests.run(t)
}

func TestBCS(t *testing.T) {
	tests := testCases{
		{
			name:        "no branch",
			program:     []uint8{0xb0, 0x02},
			expectCarry: false,
			expectPC:    newUint16(ProgramStart + 0x02),
		},
		{
			name:        "branch",
			program:     []uint8{0xb0, 0x10},
			setupCarry:  newBool(true),
			expectCarry: true,
			expectPC:    newUint16(ProgramStart + 0x02 + 0x10),
		},
	}
	tests.run(t)
}

func TestBEQ(t *testing.T) {
	tests := testCases{
		{
			name:       "no branch",
			program:    []uint8{0xf0, 0x02},
			expectZero: false,
			expectPC:   newUint16(ProgramStart + 0x02),
		},
		{
			name:       "branch",
			program:    []uint8{0xf0, 0x10},
			setupZero:  newBool(true),
			expectZero: true,
			expectPC:   newUint16(ProgramStart + 0x02 + 0x10),
		},
	}
	tests.run(t)
}

func TestBIT(t *testing.T) {
	tests := testCases{
		{
			name:       "BIT sets Z flag when zero bit is set",
			program:    []uint8{0x24, 0x10},
			memory:     map[uint16]uint8{0x0010: 0x00},
			setupA:     newUint8(0xFF),
			expectZero: true,
		},
		{
			name:       "BIT clears Z flag when zero bit is clear",
			program:    []uint8{0x24, 0x10},
			memory:     map[uint16]uint8{0x0010: 0x01},
			setupA:     newUint8(0xFF),
			expectZero: false,
		},
		{
			name:           "BIT sets N flag when negative bit is set",
			program:        []uint8{0x24, 0x10},
			memory:         map[uint16]uint8{0x0010: 0x80},
			setupA:         newUint8(0xFF),
			expectNegative: true,
		},
		{
			name:           "BIT clears N flag when negative bit is clear",
			program:        []uint8{0x24, 0x10},
			memory:         map[uint16]uint8{0x0010: 0x7F},
			setupA:         newUint8(0xFF),
			expectNegative: false,
			expectOverflow: true,
		},
		{
			name:           "BIT sets V flag when overflow bit is set",
			program:        []uint8{0x24, 0x10},
			memory:         map[uint16]uint8{0x0010: 0x40},
			setupA:         newUint8(0xFF),
			expectOverflow: true,
		},
		{
			name:           "BIT clears V flag when overflow bit is clear",
			program:        []uint8{0x24, 0x10},
			memory:         map[uint16]uint8{0x0010: 0x3F},
			setupA:         newUint8(0xFF),
			expectOverflow: false,
		},
	}
	tests.run(t)
}

func TestBMI(t *testing.T) {
	tests := testCases{
		{
			name:           "no branch",
			program:        []uint8{0x30, 0x02},
			expectNegative: false,
			expectPC:       newUint16(ProgramStart + 0x02),
		},
		{
			name:           "branch",
			program:        []uint8{0x30, 0x10},
			setupNegative:  newBool(true),
			expectNegative: true,
			expectPC:       newUint16(ProgramStart + 0x02 + 0x10),
		},
	}
	tests.run(t)
}

func TestBNE(t *testing.T) {
	tests := testCases{
		{
			name:       "no branch",
			program:    []uint8{0xd0, 0x02},
			setupZero:  newBool(true),
			expectZero: true,
			expectPC:   newUint16(ProgramStart + 0x02),
		},
		{
			name:       "branch",
			program:    []uint8{0xd0, 0x10},
			expectZero: false,
			expectPC:   newUint16(ProgramStart + 0x02 + 0x10),
		},
	}

	tests.run(t)
}

func TestBPL(t *testing.T) {
	tests := testCases{
		{
			name:           "no branch",
			program:        []uint8{0x10, 0x02},
			setupNegative:  newBool(true),
			expectNegative: true,
			expectPC:       newUint16(ProgramStart + 0x02),
		},
		{
			name:           "branch",
			program:        []uint8{0x10, 0x10},
			expectNegative: false,
			expectPC:       newUint16(ProgramStart + 0x02 + 0x10),
		},
	}
	tests.run(t)
}

func TestBRK(t *testing.T) {
	tests := testCases{
		{
			name: "BRK sets B flag and pushes PC and status to stack",
			program: []uint8{
				0x00, // BRK
			},
			memory: map[uint16]uint8{
				IRQVectorLow:  0x10,
				IRQVectorHigh: 0x10,
			},
			expectPC: newUint16(0x1010),
			expectSP: newUint8(StackTop - 0x03), // lo, hi, pc
			expectMemory: map[uint16]uint8{
				stackAddress(StackTop):       0xdd, // push PC high byte
				stackAddress(StackTop - 0x1): 0x02, // push PC low byte
				stackAddress(StackTop - 0x2): 0x34, // push status with B flag set
			},
			expectBreak:            newBool(true),
			expectInterruptDisable: newBool(true),
		},
		{
			name: "BRK with other flags set",
			program: []uint8{
				0x00, // BRK
			},
			memory: map[uint16]uint8{
				IRQVectorLow:  0x10,
				IRQVectorHigh: 0x10,
			},
			setupCarry:    newBool(true),
			setupZero:     newBool(true),
			setupOverflow: newBool(true),
			setupNegative: newBool(true),
			expectPC:      newUint16(0x1010),
			expectSP:      newUint8(StackTop - 0x03),
			expectMemory: map[uint16]uint8{
				stackAddress(StackTop):       0xdd,
				stackAddress(StackTop - 0x1): 0x02,
				stackAddress(StackTop - 0x2): 0xF7, // All flags set except zero
			},
			expectBreak:            newBool(true),
			expectInterruptDisable: newBool(true),
			expectCarry:            true,
			expectZero:             true,
			expectOverflow:         true,
			expectNegative:         true,
		},
		{
			name: "BRK with no other flags set",
			program: []uint8{
				0x00, // BRK
			},
			memory: map[uint16]uint8{
				IRQVectorLow:  0x10,
				IRQVectorHigh: 0x10,
			},
			expectPC: newUint16(0x1010),
			expectSP: newUint8(StackTop - 0x03),
			expectMemory: map[uint16]uint8{
				stackAddress(StackTop):       0xdd,
				stackAddress(StackTop - 0x1): 0x02,
				stackAddress(StackTop - 0x2): 0x34, // Only B flag and reserved flag set
			},
			expectBreak:            newBool(true),
			expectInterruptDisable: newBool(true),
		},
	}
	tests.run(t)
}

func TestBVC(t *testing.T) {
	tests := testCases{
		{
			name:           "no branch",
			program:        []uint8{0x50, 0x02},
			setupOverflow:  newBool(true),
			expectOverflow: true,
			expectPC:       newUint16(ProgramStart + 0x02),
		},
		{
			name:           "branch",
			program:        []uint8{0x50, 0x10},
			expectOverflow: false,
			expectPC:       newUint16(ProgramStart + 0x02 + 0x10),
		},
	}
	tests.run(t)
}

func TestBVS(t *testing.T) {
	tests := testCases{
		{
			name:           "no branch",
			program:        []uint8{0x70, 0x02},
			expectOverflow: false,
			expectPC:       newUint16(ProgramStart + 0x02),
		},
		{
			name:           "branch",
			program:        []uint8{0x70, 0x10},
			setupOverflow:  newBool(true),
			expectOverflow: true,
			expectPC:       newUint16(ProgramStart + 0x02 + 0x10),
		},
	}
	tests.run(t)
}

func TestCLC(t *testing.T) {
	tests := testCases{
		{
			name:        "clear carry",
			program:     []uint8{0x18},
			setupCarry:  newBool(true),
			expectCarry: false,
		},
		{
			name:        "clear unset carry",
			program:     []uint8{0x18},
			expectCarry: false,
		},
	}
	tests.run(t)
}

func TestCLD(t *testing.T) {
	tests := testCases{
		{
			name:          "clear decimal",
			program:       []uint8{0xd8},
			setupDecimal:  newBool(true),
			expectDecimal: newBool(false),
		},
		{
			name:          "clear unset decimal",
			program:       []uint8{0xd8},
			expectDecimal: newBool(false),
		},
	}
	tests.run(t)
}

func TestCLI(t *testing.T) {
	tests := testCases{
		{
			name:                   "clear interrupt",
			program:                []uint8{0x58},
			setupInterruptDisable:  newBool(true),
			expectInterruptDisable: newBool(false),
		},
		{
			name:                   "clear unset interrupt",
			program:                []uint8{0x58},
			expectInterruptDisable: newBool(false),
		},
	}
	tests.run(t)
}

func TestCLV(t *testing.T) {
	tests := testCases{
		{
			name:           "clear overflow",
			program:        []uint8{0xb8},
			setupOverflow:  newBool(true),
			expectOverflow: false,
		},
		{
			name:           "clear unset overflow",
			program:        []uint8{0xb8},
			expectOverflow: false,
		},
	}
	tests.run(t)
}

func TestCMP(t *testing.T) {
	tests := testCases{
		{
			name: "Immediate, equal",
			program: []uint8{
				0xC9, // CMP
				0x0A, // Immediate value
			},
			setupA:                 newUint8(0x0A),
			expectA:                newUint8(0x0A),
			expectPC:               newUint16(ProgramStart + 2),
			expectZero:             true,
			expectCarry:            true,
			expectNegative:         false,
			expectInterruptDisable: nil,
			expectDecimal:          nil,
			expectBreak:            nil,
			expectMemory:           nil,
		},
		{
			name: "Immediate, greater",
			program: []uint8{
				0xC9, // CMP
				0x05, // Immediate value
			},
			setupA:                 newUint8(0x0A),
			expectA:                newUint8(0x0A),
			expectPC:               newUint16(ProgramStart + 2),
			expectZero:             false,
			expectCarry:            true,
			expectNegative:         false,
			expectInterruptDisable: nil,
			expectDecimal:          nil,
			expectBreak:            nil,
			expectMemory:           nil,
		},
		{
			name: "Immediate, less",
			program: []uint8{
				0xC9, // CMP
				0x0F, // Immediate value
			},
			setupA:                 newUint8(0x0A),
			expectA:                newUint8(0x0A),
			expectPC:               newUint16(ProgramStart + 2),
			expectZero:             false,
			expectCarry:            false,
			expectNegative:         true,
			expectInterruptDisable: nil,
			expectDecimal:          nil,
			expectBreak:            nil,
			expectMemory:           nil,
		},
		// Add more test cases here for other addressing modes and scenarios
	}
	tests.run(t)
}

func TestCPX(t *testing.T) {
	tests := testCases{
		{
			name: "Immediate, equal",
			// Load the CPX immediate instruction (0xE0) followed by the value 0x42
			program: []uint8{0xE0, 0x42},
			// Set the X register to 0x42 before executing the instruction
			setupX: newUint8(0x42),
			// Run the instruction for 2 cycles
			// Expect the Zero flag to be true after executing the instruction
			expectZero: true,
			// Expect the Carry flag to be true after executing the instruction
			expectCarry: true,
			// Expect the Negative flag to be false after executing the instruction
			expectNegative: false,
			// Expect the program counter to be incremented by 2 after executing the instruction
			expectPC: newUint16(ProgramStart + 2),
		},
		{
			name: "Immediate, less",
			// Load the CPX immediate instruction (0xE0) followed by the value 0x42
			program: []uint8{0xE0, 0x42},
			// Set the X register to 0x40 before executing the instruction
			setupX: newUint8(0x40),
			// Run the instruction for 2 cycles
			// Expect the Zero flag to be false after executing the instruction
			expectZero: false,
			// Expect the Carry flag to be false after executing the instruction
			expectCarry: false,
			// Expect the Negative flag to be true after executing the instruction
			expectNegative: true,
			// Expect the program counter to be incremented by 2 after executing the instruction
			expectPC: newUint16(ProgramStart + 2),
		},
		{
			name: "Immediate, greater",
			// Load the CPX immediate instruction (0xE0) followed by the value 0x42
			program: []uint8{0xE0, 0x42},
			// Set the X register to 0x44 before executing the instruction
			setupX: newUint8(0x44),
			// Run the instruction for 2 cycles
			// Expect the Zero flag to be false after executing the instruction
			expectZero: false,
			// Expect the Carry flag to be true after executing the instruction
			expectCarry: true,
			// Expect the Negative flag to be false after executing the instruction
			expectNegative: false,
			// Expect the program counter to be incremented by 2 after executing the instruction
			expectPC: newUint16(ProgramStart + 2),
		},
		{
			name: "zeropage, equal",
			// Load the CPX zeropage instruction (0xE4) followed by the zeropage address 0x10
			program: []uint8{0xE4, 0x10},
			// Set the X register to 0x42 before executing the instruction
			setupX: newUint8(0x42),
			// Set the value at zeropage address 0x10 to 0x42
			memory: map[uint16]uint8{0x10: 0x42},
			// Run the instruction for 3 cycles
			// Expect the Zero flag to be true after executing the instruction
			expectZero: true,
			// Expect the Carry flag to be true after executing the instruction
			expectCarry: true,
			// Expect the Negative flag to be false after executing the instruction
			expectNegative: false,
			// Expect the program counter to be incremented by 2 after executing the instruction
			expectPC: newUint16(ProgramStart + 2),
		},
		{
			name: "Absolute, equal",
			// Load the CPX absolute instruction (0xEC) followed by the absolute address 0x1234
			program: []uint8{0xEC, 0x34, 0x12},
			// Set the X register to 0x42 before executing the instruction
			setupX: newUint8(0x42),
			// Set the value at absolute address 0x1234 to 0x42
			memory: map[uint16]uint8{0x1234: 0x42},
			// Run the instruction for 4 cycles
			// Expect the Zero flag to be true after executing the instruction
			expectZero: true,
			// Expect the Carry flag to be true after executing the instruction
			expectCarry: true,
			// Expect the Negative flag to be false after executing the instruction
			expectNegative: false,
			// Expect the program counter to be incremented by 3 after executing the instruction
			expectPC: newUint16(ProgramStart + 3),
		},
	}
	tests.run(t)
}

func TestCPY(t *testing.T) {
	tests := testCases{
		{
			name: "Immediate, equal",
			// Load the CPY immediate instruction (0xE0) followed by the value 0x42
			program: []uint8{0xc0, 0x42},
			// Set the Y register to 0x42 before executing the instruction
			setupY: newUint8(0x42),
			// Run the instruction for 2 cycles
			// Expect the Zero flag to be true after executing the instruction
			expectZero: true,
			// Expect the Carry flag to be true after executing the instruction
			expectCarry: true,
			// Expect the Negative flag to be false after executing the instruction
			expectNegative: false,
			// Expect the program counter to be incremented by 2 after executing the instruction
			expectPC: newUint16(ProgramStart + 2),
		},
		{
			name: "Immediate, less",
			// Load the CPY immediate instruction (0xE0) followed by the value 0x42
			program: []uint8{0xc0, 0x42},
			// Set the Y register to 0x40 before executing the instruction
			setupY: newUint8(0x40),
			// Run the instruction for 2 cycles
			// Expect the Zero flag to be false after executing the instruction
			expectZero: false,
			// Expect the Carry flag to be false after executing the instruction
			expectCarry: false,
			// Expect the Negative flag to be true after executing the instruction
			expectNegative: true,
			// Expect the program counter to be incremented by 2 after executing the instruction
			expectPC: newUint16(ProgramStart + 2),
		},
		{
			name: "Immediate, greater",
			// Load the CPY immediate instruction (0xE0) followed by the value 0x42
			program: []uint8{0xc0, 0x42},
			// Set the Y register to 0x44 before executing the instruction
			setupY: newUint8(0x44),
			// Run the instruction for 2 cycles
			// Expect the Zero flag to be false after executing the instruction
			expectZero: false,
			// Expect the Carry flag to be true after executing the instruction
			expectCarry: true,
			// Expect the Negative flag to be false after executing the instruction
			expectNegative: false,
			// Expect the program counter to be incremented by 2 after executing the instruction
			expectPC: newUint16(ProgramStart + 2),
		},
		{
			name: "zeropage, equal",
			// Load the CPY zeropage instruction (0xE4) followed by the zeropage address 0x10
			program: []uint8{0xc4, 0x10},
			// Set the Y register to 0x42 before executing the instruction
			setupY: newUint8(0x42),
			// Set the value at zeropage address 0x10 to 0x42
			memory: map[uint16]uint8{0x10: 0x42},
			// Run the instruction for 3 cycles
			// Expect the Zero flag to be true after executing the instruction
			expectZero: true,
			// Expect the Carry flag to be true after executing the instruction
			expectCarry: true,
			// Expect the Negative flag to be false after executing the instruction
			expectNegative: false,
			// Expect the program counter to be incremented by 2 after executing the instruction
			expectPC: newUint16(ProgramStart + 2),
		},
		{
			name: "Absolute, equal",
			// Load the CPY absolute instruction (0xEC) followed by the absolute address 0x1234
			program: []uint8{0xcC, 0x34, 0x12},
			// Set the Y register to 0x42 before executing the instruction
			setupY: newUint8(0x42),
			// Set the value at absolute address 0x1234 to 0x42
			memory: map[uint16]uint8{0x1234: 0x42},
			// Run the instruction for 4 cycles
			// Expect the Zero flag to be true after executing the instruction
			expectZero: true,
			// Expect the Carry flag to be true after executing the instruction
			expectCarry: true,
			// Expect the Negative flag to be false after executing the instruction
			expectNegative: false,
			// Expect the program counter to be incremented by 3 after executing the instruction
			expectPC: newUint16(ProgramStart + 3),
		},
	}
	tests.run(t)
}

func TestDEC(t *testing.T) {
	tests := testCases{
		// Test DEC with zeropage addressing
		{
			name: "DEC zeropage",
			program: []uint8{
				0xc6, 0x10, // DEC $10
			},
			memory: map[uint16]uint8{
				0x0010: 0x02, // memory location $10 contains 0x02
			},
			expectMemory: map[uint16]uint8{
				0x0010: 0x01, // memory location $10 should be decremented to 0x01
			},
		},
		// Test DEC with zeropage, X addressing
		{
			name: "DEC zeropage, X",
			program: []uint8{
				0xd6, 0x10, // DEC $10,X
			},
			setupX: newUint8(0x01),
			memory: map[uint16]uint8{
				0x0011: 0x03, // memory location $11 ($10 + X) contains 0x03
			},
			expectMemory: map[uint16]uint8{
				0x0011: 0x02, // memory location $11 should be decremented to 0x02
			},
		},
		// Test DEC with Absolute addressing
		{
			name: "DEC Absolute",
			program: []uint8{
				0xce, 0x01, 0x20, // DEC $2001
			},
			memory: map[uint16]uint8{
				0x2001: 0x04, // memory location $2001 contains 0x04
			},
			expectMemory: map[uint16]uint8{
				0x2001: 0x03, // memory location $2001 should be decremented to 0x03
			},
		},
	}
	tests.run(t)
}

func TestDEX(t *testing.T) {
	tests := testCases{
		{
			name:           "DEX - Zero flag set",
			program:        []uint8{0xca},  // DEX opcode
			setupX:         newUint8(0x01), // Initial value of X register
			expectX:        newUint8(0x00), // Expect X register to be decremented by 1
			expectPC:       newUint16(0xdd01),
			expectZero:     true,
			expectNegative: false,
		},
		{
			name:           "DEX - Negative flag set",
			program:        []uint8{0xca},  // DEX opcode
			setupX:         newUint8(0x00), // Initial value of X register
			expectX:        newUint8(0xff), // Expect X register to wrap around and become 0xFF
			expectPC:       newUint16(0xdd01),
			expectZero:     false,
			expectNegative: true,
		},
		{
			name:           "DEX - No flags set",
			program:        []uint8{0xca},  // DEX opcode
			setupX:         newUint8(0x02), // Initial value of X register
			expectX:        newUint8(0x01), // Expect X register to be decremented by 1
			expectPC:       newUint16(0xdd01),
			expectZero:     false,
			expectNegative: false,
		},
	}
	tests.run(t)
}

func TestDEY(t *testing.T) {
	tests := testCases{
		{
			name:           "DEY - Zero flag set",
			program:        []uint8{0x88},  // DEY opcode
			setupY:         newUint8(0x01), // Initial value of Y register
			expectY:        newUint8(0x00), // Expect Y register to be decremented by 1
			expectPC:       newUint16(0xdd01),
			expectZero:     true,
			expectNegative: false,
		},
		{
			name:           "DEY - Negative flag set",
			program:        []uint8{0x88},  // DEY opcode
			setupY:         newUint8(0x00), // Initial value of Y register
			expectY:        newUint8(0xff), // Expect Y register to wrap around and become 0xFF
			expectPC:       newUint16(0xdd01),
			expectZero:     false,
			expectNegative: true,
		},
		{
			name:           "DEY - No flags set",
			program:        []uint8{0x88},  // DEY opcode
			setupY:         newUint8(0x02), // Initial value of Y register
			expectY:        newUint8(0x01), // Expect Y register to be decremented by 1
			expectPC:       newUint16(0xdd01),
			expectZero:     false,
			expectNegative: false,
		},
	}
	tests.run(t)
}

func TestEOR(t *testing.T) {
	tests := testCases{
		// Test EOR Immediate mode
		{
			name:           "EOR immediate mode, no carry",
			program:        []uint8{0x49, 0x0F}, // EOR #$0F
			memory:         make(map[uint16]uint8),
			setupA:         newUint8(0xF0),
			expectA:        newUint8(0xFF),
			expectPC:       newUint16(ProgramStart + 2),
			expectCarry:    false,
			expectZero:     false,
			expectOverflow: false,
			expectNegative: true,
		},
		// Test EOR zeropage mode
		{
			name:           "EOR zeropage mode",
			program:        []uint8{0x45, 0x10}, // EOR $10
			memory:         map[uint16]uint8{0x0010: 0x0F},
			setupA:         newUint8(0xF0),
			expectA:        newUint8(0xFF),
			expectPC:       newUint16(ProgramStart + 2),
			expectCarry:    false,
			expectZero:     false,
			expectOverflow: false,
			expectNegative: true,
		},
	}
	tests.run(t)
}

func TestINX(t *testing.T) {
	tests := testCases{
		{
			name:    "inx 0x0",
			program: []uint8{0xe8},
			expectX: newUint8(0x1),
		},
		{
			name:    "inx 0aa",
			program: []uint8{0xe8},
			setupX:  newUint8(0x0a),
			expectX: newUint8(0x0b),
		},
	}
	tests.run(t)
}

func TestINY(t *testing.T) {
	tests := testCases{
		{
			name:    "iny 0x0",
			program: []uint8{0xc8},
			expectY: newUint8(0x1),
		},
		{
			name:    "iny 0aa",
			program: []uint8{0xc8},
			expectY: newUint8(0x0b),
			setupY:  newUint8(0x0a),
		},
	}
	tests.run(t)
}

func TestINC(t *testing.T) {
	tests := testCases{
		{
			name:         "zeropage",
			program:      []uint8{0xe6, 0x42},
			memory:       map[uint16]uint8{0x0042: 0x09},
			expectMemory: map[uint16]uint8{0x0042: 0x0a},
		},
		{
			name:         "zeropage,x",
			program:      []uint8{0xf6, 0x42},
			memory:       map[uint16]uint8{0x0043: 0x09},
			expectMemory: map[uint16]uint8{0x0043: 0x0a},
			setupX:       newUint8(0x1),
		},
		{
			name:         "absolute",
			program:      []uint8{0xee, 0x42, 0xaa},
			memory:       map[uint16]uint8{0xaa42: 0x09},
			expectMemory: map[uint16]uint8{0xaa42: 0x0a},
		},
		{
			name:         "absolute,x",
			program:      []uint8{0xfe, 0x42, 0xaa},
			memory:       map[uint16]uint8{0xaa43: 0x09},
			expectMemory: map[uint16]uint8{0xaa43: 0x0a},
			setupX:       newUint8(0x1),
		},
	}
	tests.run(t)
}

func TestJMP(t *testing.T) {
	tests := testCases{
		{
			name:     "absolute",
			program:  []uint8{0x4c, 0x00, 0x04},
			expectPC: newUint16(0x0400),
		},
		{
			name:    "indirect",
			program: []uint8{0x6c, 0x00, 0x04},
			memory: map[uint16]uint8{
				0x0400: 0x42,
				0x0401: 0x23,
				0x042:  0x23,
				0x043:  0x42,
			},
			expectPC: newUint16(0x2342),
		},
	}
	tests.run(t)
}

func TestJSR(t *testing.T) {
	tests := testCases{
		{
			name:    "jsr",
			program: []uint8{0x20, 0x01, 0x04},
			expectMemory: map[uint16]uint8{
				stackAddress(StackTop):        0xdd,
				stackAddress(StackTop - 0x01): 0x02,
			},
		},
	}
	tests.run(t)
}

func TestLDA(t *testing.T) {
	tests := testCases{
		{
			name:    "immediate",
			program: []uint8{0xa9, 0x42},
			expectA: newUint8(0x42),
		},
		{
			name:       "immediate, with zero",
			program:    []uint8{0xa9, 0x00},
			expectA:    newUint8(0x00),
			expectZero: true,
		},
		{
			name:           "zeropage",
			program:        []uint8{0xa5, 0x01},
			memory:         map[uint16]uint8{0x01: 0x99},
			expectA:        newUint8(0x99),
			expectNegative: true,
		},
		{
			name:    "zeropage,x(x=0)",
			program: []uint8{0xb5, 0x80},
			memory:  map[uint16]uint8{0x0080: 0x40},
			expectA: newUint8(0x40),
		},
		{
			name:    "zeropage,x(x=0x02)",
			program: []uint8{0xb5, 0x80},
			memory:  map[uint16]uint8{0x82: 0x40},
			setupX:  newUint8(0x02),
			expectA: newUint8(0x40),
		},
		{
			name:    "absolute",
			program: []uint8{0xad, 0x10, 0x30},
			memory:  map[uint16]uint8{0x3010: 0x22},
			expectA: newUint8(0x22),
		},
		{
			name:    "absolute,x(x=0)",
			program: []uint8{0xbd, 0x10, 0x30},
			memory:  map[uint16]uint8{0x3010: 0x22},
			expectA: newUint8(0x22),
		},
		{
			name:    "absolute,x(x=2)",
			program: []uint8{0xbd, 0x10, 0x30},
			memory:  map[uint16]uint8{0x3012: 0x22},
			setupX:  newUint8(0x02),
			expectA: newUint8(0x22),
		},
		{
			name:    "absolute,y(y=0)",
			program: []uint8{0xb9, 0x10, 0x30},
			memory:  map[uint16]uint8{0x3010: 0x22},
			expectA: newUint8(0x22),
		},
		{
			name:    "absolute,y(y=2)",
			program: []uint8{0xb9, 0x10, 0x30},
			memory:  map[uint16]uint8{0x3012: 0x22},
			setupY:  newUint8(0x02),
			expectA: newUint8(0x22),
		},
		{
			name:    "(indirect,x)(x=0x05)",
			program: []uint8{0xa1, 0x70},
			memory: map[uint16]uint8{
				0x0075: 0x32,
				0x0076: 0x30,
				0x3032: 0xa5,
			},
			setupX:         newUint8(0x05),
			expectA:        newUint8(0xa5),
			expectNegative: true,
		},
		{
			name:    "(indirect,y)(y=0x10)",
			program: []uint8{0xb1, 0x70},
			memory: map[uint16]uint8{
				0x0070: 0x43,
				0x53:   0x23,
			},
			setupY:  newUint8(0x10),
			expectA: newUint8(0x23),
		},
	}
	tests.run(t)
}

func TestLDX(t *testing.T) {
	tests := testCases{
		{
			name:    "immediate",
			program: []uint8{0xa2, 0x42},
			expectX: newUint8(0x42),
		},
		{
			name:    "zeropage",
			program: []uint8{0xa6, 0x42},
			memory:  map[uint16]uint8{0x0042: 0x1},
			expectX: newUint8(0x1),
		},
		{
			name:    "zeropage,y",
			program: []uint8{0xb6, 0x42},
			memory:  map[uint16]uint8{0x0043: 0x1},
			expectX: newUint8(0x1),
			setupY:  newUint8(0x1),
		},
		{
			name:    "absolute",
			program: []uint8{0xae, 0x42, 0xaa},
			memory:  map[uint16]uint8{0xaa42: 0x1},
			expectX: newUint8(0x1),
		},
		{
			name:    "absolute,y",
			program: []uint8{0xbe, 0x42, 0xaa},
			memory:  map[uint16]uint8{0xaa43: 0x1},
			expectX: newUint8(0x1),
			setupY:  newUint8(0x1),
		},
	}
	tests.run(t)
}

func TestLDY(t *testing.T) {
	tests := testCases{
		{
			name:    "immediate",
			program: []uint8{0xa0, 0x42},
			expectY: newUint8(0x42),
		},
		{
			name:    "zeropage",
			program: []uint8{0xa4, 0x42},
			memory:  map[uint16]uint8{0x0042: 0x1},
			expectY: newUint8(0x1),
		},
		{
			name:    "zeropage,x",
			program: []uint8{0xb4, 0x42},
			memory:  map[uint16]uint8{0x0043: 0x1},
			setupX:  newUint8(0x1),
			expectY: newUint8(0x1),
		},
		{
			name:    "absolute",
			program: []uint8{0xac, 0x42, 0xaa},
			memory:  map[uint16]uint8{0xaa42: 0x1},
			expectY: newUint8(0x1),
		},
		{
			name:    "absolute,x",
			program: []uint8{0xbc, 0x42, 0xaa},
			memory:  map[uint16]uint8{0xaa43: 0x1},
			setupX:  newUint8(0x1),
			expectY: newUint8(0x1),
		},
	}
	tests.run(t)
}

func TestLSR(t *testing.T) {
	tests := testCases{
		{
			name:    "accumulator",
			program: []uint8{0x4a},
			expectA: newUint8(0x55),
		},
		{
			name:       "accumulator 0",
			program:    []uint8{0x4a},
			setupA:     newUint8(0x00),
			expectA:    newUint8(0x00),
			expectZero: true,
		},
		{
			name:         "zeropage",
			program:      []uint8{0x46, 0x42},
			memory:       map[uint16]uint8{0x0042: 0x55},
			expectMemory: map[uint16]uint8{0x0042: 0x2a},
			expectCarry:  true,
		},
		{
			name:         "zeropage,x",
			program:      []uint8{0x56, 0x42},
			memory:       map[uint16]uint8{0x0047: 0x55},
			expectMemory: map[uint16]uint8{0x0047: 0x2a},
			setupX:       newUint8(0x5),
			expectCarry:  true,
		},
		{
			name:         "absolute",
			program:      []uint8{0x4e, 0x42},
			memory:       map[uint16]uint8{0x0042: 0x55},
			expectMemory: map[uint16]uint8{0x0042: 0x2a},
			expectCarry:  true,
		},
		{
			name:         "absolute,x",
			program:      []uint8{0x5e, 0x42},
			memory:       map[uint16]uint8{0x0047: 0x55},
			expectMemory: map[uint16]uint8{0x0047: 0x2a},
			setupX:       newUint8(0x5),
			expectCarry:  true,
		},
	}
	tests.run(t)
}

func TestNOP(t *testing.T) {
	tests := testCases{
		{
			name:    "implied",
			program: []uint8{0xea},
		},
	}
	tests.run(t)
}

func TestORA(t *testing.T) {
	tests := testCases{
		{
			name:    "immediate",
			program: []uint8{0x09, 0x42},
			setupA:  newUint8(0x10),
			expectA: newUint8(0x52),
		},
		{
			name:    "zeropage",
			program: []uint8{0x05, 0x42},
			memory:  map[uint16]uint8{0x0042: 0x42},
			setupA:  newUint8(0x10),
			expectA: newUint8(0x52),
		},
		{
			name:    "zeropage,x",
			program: []uint8{0x15, 0x42},
			memory:  map[uint16]uint8{0x0043: 0x42},
			setupA:  newUint8(0x10),
			setupX:  newUint8(0x01),
			expectA: newUint8(0x52),
		},
		{
			name:    "absolute",
			program: []uint8{0x0d, 0x42, 0xaa},
			memory:  map[uint16]uint8{0xaa42: 0x42},
			setupA:  newUint8(0x10),
			expectA: newUint8(0x52),
		},
		{
			name:    "absolute,x",
			program: []uint8{0x1d, 0x42, 0xaa},
			memory:  map[uint16]uint8{0xaa43: 0x42},
			setupA:  newUint8(0x10),
			setupX:  newUint8(0x01),
			expectA: newUint8(0x52),
		},
		{
			name:    "absolute,y",
			program: []uint8{0x19, 0x42, 0xaa},
			memory:  map[uint16]uint8{0xaa43: 0x42},
			setupA:  newUint8(0x10),
			setupY:  newUint8(0x01),
			expectA: newUint8(0x52),
		},
		{
			name:    "(indirect,x)",
			program: []uint8{0x01, 0xaa},
			memory: map[uint16]uint8{
				0x00ab: 0xcc,
				0x00cc: 0x42,
			},
			setupA:  newUint8(0x10),
			setupX:  newUint8(0x01),
			expectA: newUint8(0x52),
		},
		{
			name:    "(indirect),y",
			program: []uint8{0x11, 0xaa},
			memory: map[uint16]uint8{
				0xaa: 0xcc,
				0xcd: 0x42,
			},
			setupA:  newUint8(0x10),
			setupY:  newUint8(0x01),
			expectA: newUint8(0x52),
		},
	}
	tests.run(t)
}

func TestPHA(t *testing.T) {
	tests := testCases{
		{
			name:     "PHA basic",
			program:  []uint8{0x48}, // PHA
			setupA:   newUint8(0x42),
			expectA:  newUint8(0x42),
			expectSP: newUint8(0xfe),
			expectMemory: map[uint16]uint8{
				stackAddress(StackTop): 0x42,
			},
		},
		{
			name:     "PHA with wraparound",
			program:  []uint8{0x48}, // PHA
			setupA:   newUint8(0x42),
			setupSP:  newUint8(StackBottom),
			expectA:  newUint8(0x42),
			expectSP: newUint8(StackTop),
			expectMemory: map[uint16]uint8{
				stackAddress(StackBottom): 0x42,
			},
		},
	}
	tests.run(t)
}

func TestPHP(t *testing.T) {
	tests := testCases{
		{
			name: "push processor status with zero flag set",
			program: []uint8{
				0x08, // PHP
			},
			setupCarry: newBool(false),
			setupZero:  newBool(true),
			expectMemory: map[uint16]uint8{
				stackAddress(StackTop): 0x36,
			},
			expectSP:    newUint8(StackTop - 0x01),
			expectZero:  true,
			expectCarry: false,
		},
		{
			name: "push processor status with zero flag and carry set",
			program: []uint8{
				0x08, // PHP
			},
			setupCarry: newBool(true),
			setupZero:  newBool(true),
			expectMemory: map[uint16]uint8{
				stackAddress(StackTop): 0x37},
			expectSP:    newUint8(StackTop - 0x01),
			expectZero:  true,
			expectCarry: true,
		},
		{
			name: "push processor status with negative flag set",
			program: []uint8{
				0x08, // PHP
			},
			setupNegative:  newBool(true),
			expectMemory:   map[uint16]uint8{stackAddress(StackTop): 0xb4},
			expectSP:       newUint8(StackTop - 0x01),
			expectNegative: true,
		},
	}
	tests.run(t)
}

func TestPLA(t *testing.T) {
	tests := testCases{
		{
			name:     "pull from stack + 1",
			program:  []uint8{0x68}, // PLA
			setupSP:  newUint8(StackTop - 0x01),
			memory:   map[uint16]uint8{stackAddress(StackTop): 0x42},
			setupA:   newUint8(0x7f),
			expectA:  newUint8(0x42),
			expectSP: newUint8(StackTop),
		},
		{
			name:    "pull from stack wrap to bottom",
			program: []uint8{0x68}, // PLA
			setupSP: newUint8(StackTop),
			memory: map[uint16]uint8{
				stackAddress(StackBottom): 0x42,
			},
			setupA:   newUint8(0x7f),
			expectA:  newUint8(0x42),
			expectSP: newUint8(StackBottom),
		},
	}
	tests.run(t)
}

func TestPLP(t *testing.T) {
	tests := testCases{
		{
			name:                   "PLP sets all flags",
			program:                []uint8{0x28}, // PLP
			expectCarry:            true,
			expectZero:             true,
			expectDecimal:          newBool(true),
			expectInterruptDisable: newBool(true),
			expectOverflow:         true,
			expectNegative:         true,
			expectBreak:            newBool(false),
			expectReserved:         true,
			setupSP:                newUint8(StackTop - 0x01),
			memory:                 map[uint16]uint8{stackAddress(StackTop): 0xff},
		},
		{
			name:                   "PLP sets no flags",
			program:                []uint8{0x28}, // PLP
			expectCarry:            false,
			expectZero:             false,
			expectDecimal:          newBool(false),
			expectInterruptDisable: newBool(false),
			expectOverflow:         false,
			expectNegative:         false,
			setupSP:                newUint8(StackTop - 0x01),
			memory:                 map[uint16]uint8{stackAddress(StackTop): 0x00},
		},
		{
			name:                   "PLP sets some flags",
			program:                []uint8{0x28}, // PLP
			expectDecimal:          newBool(true),
			expectInterruptDisable: newBool(true),
			expectNegative:         true,
			setupSP:                newUint8(StackTop - 0x01),
			memory:                 map[uint16]uint8{stackAddress(StackTop): 0x8c},
		},
	}
	tests.run(t)
}

func TestROL(t *testing.T) {
	tests := testCases{
		{
			name: "ROL accumulator, no carry",
			program: []uint8{
				0x2A, // ROL A
			},
			setupA:  newUint8(0b01010101),
			expectA: newUint8(0b10101010),
			// Flags
			expectCarry:    false,
			expectZero:     false,
			expectOverflow: false,
			expectNegative: true,
		},
		{
			name: "ROL accumulator, with carry",
			program: []uint8{
				0x2A, // ROL A
			},
			setupA:  newUint8(0b10000001),
			expectA: newUint8(0b00000010),
			// Flags
			expectCarry:    true,
			expectZero:     false,
			expectOverflow: false,
			expectNegative: false,
		},
		{
			name: "ROL zero page",
			program: []uint8{
				0x26, 0x10, // ROL $10
			},
			memory: map[uint16]uint8{
				0x0010: 0b01010101,
			},
			expectMemory: map[uint16]uint8{
				0x0010: 0b10101010,
			},
			// Flags
			expectCarry:    false,
			expectZero:     false,
			expectOverflow: false,
			expectNegative: true,
		},
		// Add more test cases for ROL zero page, X, absolute, absolute, X as needed
	}
	tests.run(t)
}

func TestROR(t *testing.T) {
	testCases := testCases{
		{
			name: "ROR accumulator, carry unset",
			program: []uint8{
				0x6a,
			},
			setupA:         newUint8(0x02),
			setupCarry:     newBool(false),
			expectA:        newUint8(0x01),
			expectCarry:    false,
			expectZero:     false,
			expectNegative: false,
		},
		{
			name: "ROR accumulator, carry set",
			program: []uint8{
				0x6a,
			},
			setupA:         newUint8(0x01),
			setupCarry:     newBool(true),
			expectA:        newUint8(0x80),
			expectCarry:    true,
			expectZero:     false,
			expectNegative: true,
		},
		{
			name: "ROR zero page",
			program: []uint8{
				0x66, 0x10,
			},
			memory: map[uint16]uint8{
				0x0010: 0x04,
			},
			setupCarry:     newBool(false),
			expectCarry:    false,
			expectZero:     false,
			expectNegative: false,
			expectMemory: map[uint16]uint8{
				0x0010: 0x02,
			},
		},
		// Add more test cases as needed
	}
	testCases.run(t)
}

func TestRTI(t *testing.T) {
	tests := testCases{
		{
			name:    "RTI - Basic",
			program: []uint8{0x40},
			memory: map[uint16]uint8{
				stackAddress(StackTop):     0x12, // Stack: PC High
				stackAddress(StackTop - 1): 0x34, // Stack: PC Low
				stackAddress(StackTop - 2): 0x20, // Stack: P
			},
			setupSP:  newUint8(StackTop - 3),
			expectSP: newUint8(StackTop),
			expectPC: newUint16(0x1234),
		},
		{
			name:    "RTI - Flags",
			program: []uint8{0x40},
			memory: map[uint16]uint8{
				stackAddress(StackTop):     0x12,       // Stack: PC High
				stackAddress(StackTop - 1): 0x34,       // Stack: PC Low
				stackAddress(StackTop - 2): 0b11111111, // Stack: P
			},
			setupSP:                newUint8(StackTop - 3),
			expectSP:               newUint8(StackTop),
			expectPC:               newUint16(0x1234),
			expectCarry:            true,
			expectZero:             true,
			expectInterruptDisable: newBool(true),
			expectDecimal:          newBool(true),
			expectOverflow:         true,
			expectNegative:         true,
		},
	}
	tests.run(t)
}

func TestRTS(t *testing.T) {
	tests := testCases{
		{
			name:     "RTS - Basic",
			program:  []uint8{0x20, 0x05, 0xaa}, // JSR $aa05
			memory:   map[uint16]uint8{0xaa05: 0x60},
			cycles:   3,                 // JSR takes 6 cycles, RTS takes 6 cycles
			expectPC: newUint16(0xdd03), // RTS will set the PC to the return address + 1, which is 0xdd03
		},
		{
			name: "RTS - With Stack Operations",
			program: []uint8{
				// PHA
				0x48,
				// JSR $aa05
				0x20, 0x05, 0xaa,
				// PLA
				0x68,
			},
			memory: map[uint16]uint8{
				// PHA
				0xaa05: 0x48,
				// PLA
				0xaa06: 0x68,
				// RTS
				0xaa07: 0x60,
			},
			cycles:   7,
			setupA:   newUint8(0x42),
			expectA:  newUint8(0x42),
			expectPC: newUint16(0xdd05),
		},
	}
	tests.run(t)
}

func TestSBC(t *testing.T) {
	tests := testCases{
		// SBC immediate mode, no borrow
		{
			name:        "SBC immediate mode, no borrow",
			program:     []uint8{0xE9, 0x01}, // SBC #$01
			setupA:      newUint8(0x03),
			expectA:     newUint8(0x01),
			expectCarry: true,
		},
		// SBC immediate mode, with borrow
		{
			name:        "SBC immediate mode, with borrow",
			program:     []uint8{0xE9, 0x01}, // SBC #$03
			setupCarry:  newBool(true),
			setupA:      newUint8(0x03),
			expectA:     newUint8(0x02),
			expectCarry: true,
		},
	}
	tests.run(t)
}

func TestSEC(t *testing.T) {
	// test cases
	tests := testCases{
		{
			name:        "carry flag clear",
			program:     []uint8{0x38}, // SEC
			setupCarry:  newBool(false),
			expectCarry: true,
		},
		{
			name:        "carry flag set",
			program:     []uint8{0x38}, // SEC
			setupCarry:  newBool(true),
			expectCarry: true,
		},
	}
	// run test cases
	tests.run(t)
}

func TestSED(t *testing.T) {
	// SED should set decimal flag to true
	testCases{
		{
			name:          "Set decimal flag",
			program:       []uint8{0xF8},
			expectDecimal: newBool(true),
		},
	}.run(t)
}

func TestSEI(t *testing.T) {
	testCases{
		{
			name:                   "Set interrupt disable flag",
			program:                []uint8{0x78},
			expectInterruptDisable: newBool(true),
		},
	}.run(t)
}

func TestSTA(t *testing.T) {
	tests := testCases{
		{
			name:         "zeropage",
			program:      []uint8{0x85, 0x01},
			setupA:       newUint8(0x12),
			expectMemory: map[uint16]uint8{0x0001: 0x12},
		},
		{
			name:         "zeropage,x",
			program:      []uint8{0x95, 0x01},
			setupA:       newUint8(0x12),
			setupX:       newUint8(0x1),
			expectMemory: map[uint16]uint8{0x0002: 0x12},
		},
		{
			name:         "absolute",
			program:      []uint8{0x8d, 0xaa, 0xbb},
			setupA:       newUint8(0x12),
			expectMemory: map[uint16]uint8{0xbbaa: 0x12},
		},
		{
			name:         "absolute,x",
			program:      []uint8{0x9d, 0xaa, 0xbb},
			setupA:       newUint8(0x12),
			setupX:       newUint8(0x1),
			expectMemory: map[uint16]uint8{0xbbab: 0x12},
		},
		{
			name:         "absolute,y",
			program:      []uint8{0x99, 0xaa, 0xbb},
			setupA:       newUint8(0x12),
			setupY:       newUint8(0x1),
			expectMemory: map[uint16]uint8{0xbbab: 0x12},
		},
		{
			name:         "(indirect,x)",
			program:      []uint8{0x81, 0x70},
			memory:       map[uint16]uint8{0x0071: 0x0012},
			setupA:       newUint8(0x12),
			setupX:       newUint8(0x1),
			expectMemory: map[uint16]uint8{0x0012: 0x12},
		},
		{
			name:         "(indirect),y",
			program:      []uint8{0x91, 0x70},
			memory:       map[uint16]uint8{0x0070: 0x0012},
			setupA:       newUint8(0x12),
			setupY:       newUint8(0x1),
			expectMemory: map[uint16]uint8{0x0013: 0x12},
		},
	}
	tests.run(t)
}

func TestSTX(t *testing.T) {
	// test cases
	tests := testCases{
		{
			name: "STX zeropage",
			program: []uint8{
				0x86, 0x10,
			},
			setupX: newUint8(0x42),
			expectMemory: map[uint16]uint8{
				0x0010: 0x42,
			},
			expectX: newUint8(0x42),
		},
		{
			name: "STX zeropage, Y",
			program: []uint8{
				0x96, 0x10,
			},
			setupX: newUint8(0x42),
			setupY: newUint8(0x04),
			expectMemory: map[uint16]uint8{
				0x0014: 0x42,
			},
			expectX: newUint8(0x42),
		},
		{
			name: "STX absolute",
			program: []uint8{
				0x8e, 0x34, 0x12, // STX $1234
			},
			setupX: newUint8(0x42),
			expectMemory: map[uint16]uint8{
				ProgramStart + 0: 0x8e, // STX $1234
				ProgramStart + 1: 0x34, // address low byte
				ProgramStart + 2: 0x12, // address high byte
				0x1234:           0x42, // X value
			},
			expectX: newUint8(0x42),
		},
	}
	// run tests
	tests.run(t)
}

func TestSTY(t *testing.T) {
	// test cases
	testCases := testCases{
		{
			name: "STY zeropage",
			program: []uint8{
				0x84, 0x10, // STY $10
			},
			setupY: newUint8(0xab),
			expectMemory: map[uint16]uint8{
				0x0010: 0xab,
			},
		},
		{
			name: "STY zeropage X",
			program: []uint8{
				0x94, 0x10, // STY $10,X
			},
			setupY: newUint8(0xcd),
			setupX: newUint8(0x03),
			expectMemory: map[uint16]uint8{
				0x0013: 0xcd,
			},
		},
		{
			name: "STY absolute",
			program: []uint8{
				0x8c, 0x34, 0x12, // STY $1234
			},
			setupY: newUint8(0xef),
			expectMemory: map[uint16]uint8{
				0x1234: 0xef,
			},
		},
	}
	// run test cases
	testCases.run(t)
}

func TestTAX(t *testing.T) {
	tests := testCases{
		{
			name:           "transfer a to x",
			program:        []uint8{0xaa},
			setupA:         newUint8(0x42),
			expectA:        newUint8(0x42),
			expectX:        newUint8(0x42),
			expectNegative: false,
			expectZero:     false,
		},
		{
			name:           "transfer zero to x",
			program:        []uint8{0xaa},
			setupA:         newUint8(0x00),
			expectA:        newUint8(0x00),
			expectX:        newUint8(0x00),
			expectNegative: false,
			expectZero:     true,
		},
		{
			name:           "transfer negative to x",
			program:        []uint8{0xaa},
			setupA:         newUint8(0xff),
			expectA:        newUint8(0xff),
			expectX:        newUint8(0xff),
			expectNegative: true,
			expectZero:     false,
		},
	}
	tests.run(t)
}

func TestTAY(t *testing.T) {
	tests := testCases{
		{
			name:           "transfer a to y",
			program:        []uint8{0xa8},
			setupA:         newUint8(0x42),
			expectA:        newUint8(0x42),
			expectY:        newUint8(0x42),
			expectNegative: false,
			expectZero:     false,
		},
		{
			name:           "transfer zero to y",
			program:        []uint8{0xa8},
			setupA:         newUint8(0x00),
			expectA:        newUint8(0x00),
			expectY:        newUint8(0x00),
			expectNegative: false,
			expectZero:     true,
		},
		{
			name:           "transfer negative to y",
			program:        []uint8{0xa8},
			setupA:         newUint8(0xff),
			expectA:        newUint8(0xff),
			expectY:        newUint8(0xff),
			expectNegative: true,
			expectZero:     false,
		},
	}
	tests.run(t)
}

func TestTSX(t *testing.T) {
	// initialize test cases
	tests := testCases{
		{
			name:     "positive",
			program:  []uint8{0xba}, // TSX
			setupSP:  newUint8(0x01),
			expectX:  newUint8(0x01),
			expectSP: newUint8(0x01),
		},
		{
			name:           "negative",
			program:        []uint8{0xba}, // TSX
			setupSP:        newUint8(0xfe),
			expectX:        newUint8(0xfe),
			expectSP:       newUint8(0xfe),
			expectNegative: true,
		},
		{
			name:       "zero",
			program:    []uint8{0xba}, // TSX
			setupSP:    newUint8(0x0),
			expectX:    newUint8(0x0),
			expectSP:   newUint8(0x0),
			expectZero: true,
		},
	}
	tests.run(t)
}

func TestTXA(t *testing.T) {
	tests := testCases{
		{
			name:           "transfer X to A",
			program:        []uint8{0x8a},
			setupX:         newUint8(0x42),
			expectA:        newUint8(0x42),
			expectX:        newUint8(0x42),
			expectNegative: false,
			expectZero:     false,
		},
		{
			name:           "transfer zero to X",
			program:        []uint8{0x8a},
			setupA:         newUint8(0x01),
			expectA:        newUint8(0x00),
			expectX:        newUint8(0x00),
			expectNegative: false,
			expectZero:     true,
		},
		{
			name:           "transfer negative to X",
			program:        []uint8{0x8a},
			setupX:         newUint8(0xff),
			expectA:        newUint8(0xff),
			expectX:        newUint8(0xff),
			expectNegative: true,
			expectZero:     false,
		},
	}
	tests.run(t)
}

func TestTXS(t *testing.T) {
	// TXS
	tests := testCases{
		{
			name:     "TXS with positive value",
			program:  []uint8{0x9a},
			setupX:   newUint8(0x05),
			expectSP: newUint8(0x05),
		},
		{
			name:     "TXS with zero value",
			program:  []uint8{0x9a},
			setupX:   newUint8(0x00),
			expectSP: newUint8(0x00),
		},
		{
			name:     "TXS with negative value",
			program:  []uint8{0x9a},
			setupX:   newUint8(0xf0),
			expectSP: newUint8(0xf0),
		},
	}
	tests.run(t)
}

func TestTYA(t *testing.T) {
	tests := testCases{
		{
			name:           "transfer Y to A",
			program:        []uint8{0x98},
			setupY:         newUint8(0x42),
			expectA:        newUint8(0x42),
			expectY:        newUint8(0x42),
			expectNegative: false,
			expectZero:     false,
		},
		{
			name:           "transfer zero to Y",
			program:        []uint8{0x98},
			setupA:         newUint8(0x01),
			expectA:        newUint8(0x00),
			expectY:        newUint8(0x00),
			expectNegative: false,
			expectZero:     true,
		},
		{
			name:           "transfer negative to Y",
			program:        []uint8{0x98},
			setupY:         newUint8(0xff),
			expectA:        newUint8(0xff),
			expectY:        newUint8(0xff),
			expectNegative: true,
			expectZero:     false,
		},
	}
	tests.run(t)
}
