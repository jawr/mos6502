package cpu

import (
	"testing"
)

// setup a program within a cpu and return it
func setup(program []uint8, bootstrap map[uint16]uint8) *MOS6502 {
	memory := &Memory{}

	// Reset vector
	memory[0xfffc] = 0x00
	memory[0xfffd] = 0xdd

	j := 0xdd00
	for i := 0; i < len(program); i++ {
		memory[j] = program[i]
		j++
	}

	// map any memory over
	for address, v := range bootstrap {
		memory[address] = v
	}

	cpu := NewMOS6502()
	cpu.Reset(memory)

	return cpu
}

// helper function to setup 1 byte registers
func setupUint8(register *uint8, v *uint8) {
	if v == nil {
		return
	}
	*register = *v
}

// helper function to setup 2 byte registers
func setupUint16(register *uint16, v *uint16) {
	if v == nil {
		return
	}
	*register = *v
}

// cycle a cpu n cycles
func cycle(t *testing.T, cpu *MOS6502, n uint8) {
	t.Helper()

	var i uint8
	for i = 0; i < n; i++ {
		cpu.Cycle()
	}
	if cpu.wait != 0 {
		t.Errorf("expected cycles to be 0 got %d", cpu.wait)
	}
}

// helper function to setup a uint8 pointer
func newUint8(v uint8) *uint8 {
	return &v
}

// helper function to setup a uint16 pointer
func newUint16(v uint16) *uint16 {
	return &v
}

// helper function to setup a uint16 pointer
func newBool(b bool) *bool {
	return &b
}

// common expect assert functions
func expect8(t *testing.T, a uint8, b *uint8) {
	t.Helper()
	if b == nil {
		return
	}
	if a != *b {
		t.Errorf("expected: %02x got: %02x", *b, a)
	}
}

func expect16(t *testing.T, a uint16, b *uint16) {
	t.Helper()
	if b == nil {
		return
	}
	if a != *b {
		t.Errorf("expected: %04x got: %04x", *b, a)
	}
}

// test case
type testCase struct {
	name string
	// program to load into memory
	program []uint8
	// setup memory with any bootstrap values
	memory map[uint16]uint8

	// setup registers (nil means we do not want to set)
	setupA  *uint8
	setupX  *uint8
	setupY  *uint8
	setupSP *uint8
	setupPC *uint16

	// setup flags
	setupCarry            *bool
	setupDecimal          *bool
	setupInterruptDisable *bool
	setupOverflow         *bool

	// expectations
	cycles uint8
	// expect flags
	expectCarry            bool
	expectZero             bool
	expectBreak            bool
	expectReserved         bool // should always be false
	expectOverflow         bool
	expectNegative         bool
	expectDecimal          *bool
	expectInterruptDisable *bool

	// expect registers (nil means we do not want to check)
	expectA  *uint8
	expectX  *uint8
	expectY  *uint8
	expectSP *uint8
	expectPC *uint16

	// expectMemory
	expectMemory map[uint16]uint8
}

// run a test case setting up state and then asserting
// all registers and flags
func (tc *testCase) setup(t *testing.T) *MOS6502 {
	if tc.cycles == 0 {
		t.Fatal("provided 0 cycles")
	}

	// setup state
	cpu := setup(tc.program, tc.memory)

	setupUint8(&cpu.a, tc.setupA)
	setupUint8(&cpu.x, tc.setupX)
	setupUint8(&cpu.y, tc.setupY)
	setupUint8(&cpu.sp, tc.setupSP)
	setupUint16(&cpu.pc, tc.setupPC)

	// setup flags
	if tc.setupInterruptDisable != nil {
		cpu.p.set(P_InterruptDisable)
	}
	if tc.setupDecimal != nil {
		cpu.p.set(P_Decimal)
	}
	if tc.setupOverflow != nil {
		cpu.p.set(P_Overflow)
	}
	if tc.setupCarry != nil {
		cpu.p.set(P_Carry)
	}

	return cpu
}

// run a test case setting up state and then asserting
// all registers and flags
func (tc *testCase) run(t *testing.T, cpu *MOS6502) {
	t.Run(tc.name, func(t *testing.T) {
		// run
		cycle(t, cpu, tc.cycles)

		// assert registers
		expect8(t, cpu.a, tc.expectA)
		expect8(t, cpu.x, tc.expectX)
		expect8(t, cpu.y, tc.expectY)
		expect8(t, cpu.sp, tc.expectSP)
		expect16(t, cpu.pc, tc.expectPC)

		// assert flags
		expectFlag(t, cpu, P_Carry, tc.expectCarry)
		expectFlag(t, cpu, P_Zero, tc.expectZero)
		expectFlag(t, cpu, P_Overflow, tc.expectOverflow)
		expectFlag(t, cpu, P_Negative, tc.expectNegative)

		if tc.expectInterruptDisable != nil {
			expectFlag(t, cpu, P_InterruptDisable, *tc.expectInterruptDisable)
		}
		if tc.expectDecimal != nil {
			expectFlag(t, cpu, P_Decimal, *tc.expectDecimal)
		}

		// TODO
		// expectFlag(t, cpu, P_Break, tc.expectBreak)
		// expectFlag(t, cpu, P_Decimal, tc.expectDecimal)
		// expectFlag(t, cpu, P_Reserved, tc.expectReserved)

		if tc.expectMemory != nil {
			for address, expected := range tc.expectMemory {
				if cpu.memory[address] != expected {
					t.Errorf("expected memory %04x to be %02x got %02x", address, expected, cpu.memory[address])
				}
			}
		}
	})
}

// helper type for running multiple testCases
type testCases []testCase

// run all testCases
func (tcs testCases) run(t *testing.T) {
	for _, tc := range tcs {
		cpu := tc.setup(t)
		tc.run(t, cpu)
	}
}
