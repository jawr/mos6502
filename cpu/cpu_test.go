package cpu

import (
	"log"
	"testing"
)

const (
	ProgramStart uint16 = 0xdd00
)

const DebugTests = false

// setup a program within a cpu and return it
func setup(program []uint8, bootstrap map[uint16]uint8) *MOS6502 {
	memory := &Memory{}

	// Reset vector
	memory[RESVectorLow] = uint8(ProgramStart & 0xff)
	memory[RESVectorHigh] = uint8(ProgramStart >> 8)

	for i := 0; i < len(program); i++ {
		memory[ProgramStart+uint16(i)] = program[i]
	}

	// map any memory over
	for address, v := range bootstrap {
		memory[address] = v
	}

	cpu := NewMOS6502()
	cpu.Reset(memory)
	cpu.Debug = DebugTests

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
	for i = 1; i < n; i++ {
		cpu.Cycle()
	}

	if cpu.wait != 0 {
		t.Logf("expected wait to be 0 got %d cycles should be: %d", cpu.wait, n+cpu.wait)
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
	setupZero             *bool
	setupDecimal          *bool
	setupInterruptDisable *bool
	setupOverflow         *bool
	setupNegative         *bool

	// expected number of cycles to run
	cycles uint8
	// expect flags
	expectCarry            bool
	expectZero             bool
	expectBreak            *bool
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

	// expectMemory to look like this
	expectMemory map[uint16]uint8
}

// run a test case setting up state and then asserting
// all registers and flags
func (tc *testCase) setup(t *testing.T) *MOS6502 {
	t.Helper()

	if tc.cycles == 0 {
		tc.cycles = 2
	}

	// setup state
	cpu := setup(tc.program, tc.memory)

	// setup program expected memory
	if len(tc.expectMemory) > 0 {
		for i := 0; i < len(tc.program); i++ {
			tc.expectMemory[uint16(0xdd00+i)] = tc.program[i]
		}
		// play memory over the top
		for address, b := range cpu.memory {
			if b == 0 {
				continue
			}
			// if we have a value already, prefer expected memory
			if _, ok := tc.expectMemory[uint16(address)]; ok {
				continue
			}
			tc.expectMemory[uint16(address)] = b
		}
		// add the reset vector
		tc.expectMemory[0xfffc] = 0x00
		tc.expectMemory[0xfffd] = 0xdd
	}

	setupUint8(&cpu.a, tc.setupA)
	setupUint8(&cpu.x, tc.setupX)
	setupUint8(&cpu.y, tc.setupY)
	setupUint8(&cpu.sp, tc.setupSP)
	setupUint16(&cpu.pc, tc.setupPC)

	// setup flags
	if tc.setupInterruptDisable != nil {
		cpu.p.set(P_InterruptDisable, *tc.setupInterruptDisable)
	}
	if tc.setupDecimal != nil {
		cpu.p.set(P_Decimal, *tc.setupDecimal)
	}
	if tc.setupOverflow != nil {
		cpu.p.set(P_Overflow, *tc.setupOverflow)
	}
	if tc.setupNegative != nil {
		cpu.p.set(P_Negative, *tc.setupNegative)
	}
	if tc.setupCarry != nil {
		cpu.p.set(P_Carry, *tc.setupCarry)
	}
	if tc.setupZero != nil {
		cpu.p.set(P_Zero, *tc.setupZero)
	}

	return cpu
}

// run a test case setting up state and then asserting
// all registers and flags
func (tc *testCase) run(t *testing.T, cpu *MOS6502) {
	// run
	cycle(t, cpu, tc.cycles)

	if DebugTests {
		log.Printf("Memory...")
		for address, value := range cpu.memory {
			if value == 0 {
				continue
			}
			log.Printf("\t%04x : %02x", address, value)
		}
		log.Println("--------------------")
	}

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

	if tc.expectBreak != nil {
		expectFlag(t, cpu, P_Break, *tc.expectBreak)
	}
	expectFlag(t, cpu, P_Reserved, true)

	if tc.expectMemory != nil {
		for address := range cpu.memory {
			expected := tc.expectMemory[uint16(address)]
			if cpu.memory[address] != expected {
				t.Errorf("expected memory %04x to be %02x got %02x", address, expected, cpu.memory[address])
			}
		}
	}
}

// helper type for running multiple testCases
type testCases []testCase

// run all testCases
func (tcs testCases) run(t *testing.T) {
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			cpu := tc.setup(t)
			tc.run(t, cpu)
		})
	}
}
