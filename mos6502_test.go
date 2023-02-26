package main

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

// cycle a cpu n cycles
func cycle(t *testing.T, cpu *MOS6502, n uint8) {
	var i uint8
	for i = 0; i < n; i++ {
		cpu.Cycle()
	}
	if cpu.wait != 0 {
		t.Errorf("expected cycles to be 0 got %d", cpu.wait)
	}
}

func expect8(t *testing.T, a, b uint8) {
	if a != b {
		t.Errorf("expected a:%02x to be b:%02x", a, b)
	}
}

func expect16(t *testing.T, a, b uint16) {
	if a != b {
		t.Errorf("expected a:%04x to be b:%04x", a, b)
	}
}

func TestLDA(t *testing.T) {
	cases := []struct {
		name    string
		program []uint8
		// initialise bootstrap with memory
		bootstrap map[uint16]uint8
		cycles    uint8
		expect    uint8
		// initialise registers
		x uint8
		y uint8
	}{
		{"immediate", []uint8{0xa9, 0x42}, nil, 2, 0x42, 0x0, 0x0},
		{"zeropage", []uint8{0xa5, 0x01}, map[uint16]uint8{0x0001: 0x99}, 3, 0x99, 0x0, 0x0},
		{"zeropage,x(x=0)", []uint8{0xb5, 0x80}, map[uint16]uint8{0x0080: 0x40}, 4, 0x40, 0x0, 0x0},
		{"zeropage,x(x=0x02)", []uint8{0xb5, 0x80}, map[uint16]uint8{0x0080: 0x40, 0x0082: 0xaa}, 4, 0xaa, 0x2, 0x0},
		{"absolute", []uint8{0xad, 0x10, 0x30}, map[uint16]uint8{0x3010: 0x22}, 4, 0x22, 0x0, 0x0},
		{"absolute,x(x=0)", []uint8{0xbd, 0x20, 0x31}, map[uint16]uint8{0x3120: 0x72}, 4, 0x72, 0x0, 0x0},
		{"absolute,x(x=0x02)", []uint8{0xbd, 0x20, 0x31}, map[uint16]uint8{0x3132: 0x72}, 4, 0x72, 0x12, 0x0},
		{"absolute,y(y=0)", []uint8{0xb9, 0x20, 0x31}, map[uint16]uint8{0x3120: 0x72}, 4, 0x72, 0x0, 0x0},
		{"absolute,y(y=0x02)", []uint8{0xb9, 0x20, 0x31}, map[uint16]uint8{0x3132: 0x72}, 4, 0x72, 0x0, 0x12},
		{"(indirect,x)(x=0x05)", []uint8{0xa1, 0x70}, map[uint16]uint8{0x0075: 0x32, 0x0076: 0x30, 0x3032: 0xa5}, 6, 0xa5, 0x05, 0x0},
		{"(indirect),y(y=0x10)", []uint8{0xb1, 0x70}, map[uint16]uint8{0x0070: 0x43, 0x0071: 0x35, 0x3553: 0x23}, 5, 0x23, 0x0, 0x10},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			cpu := setup(test.program, test.bootstrap)
			// initialise registers
			cpu.x = test.x
			cpu.y = test.y
			// run program
			cycle(t, cpu, test.cycles)
			// assertions
			expect8(t, cpu.a, test.expect)
		})
	}
}

func TestLDX(t *testing.T) {
	cases := []struct {
		name    string
		program []uint8
		// initialise bootstrap with memory
		bootstrap map[uint16]uint8
		cycles    uint8
		expect    uint8
		// initialise registers
		y uint8
	}{
		{"immediate", []uint8{0xa2, 0x42}, nil, 2, 0x42, 0},
		{"zeropage", []uint8{0xa6, 0x42}, map[uint16]uint8{0x0042: 0x1}, 3, 0x1, 0},
		{"zeropage,y", []uint8{0xb6, 0x42}, map[uint16]uint8{0x0043: 0x1}, 4, 0x1, 0x1},
		{"absolute", []uint8{0xae, 0x42, 0xaa}, map[uint16]uint8{0xaa42: 0x1}, 4, 0x1, 0},
		{"absolute,y", []uint8{0xbe, 0x42, 0xaa}, map[uint16]uint8{0xaa43: 0x1}, 4, 0x1, 0x1},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			cpu := setup(test.program, test.bootstrap)
			// initialise registers
			cpu.y = test.y
			// run program
			cycle(t, cpu, test.cycles)
			// assertions
			expect8(t, cpu.x, test.expect)
		})
	}
}

func TestLDY(t *testing.T) {
	cases := []struct {
		name    string
		program []uint8
		// initialise bootstrap with memory
		bootstrap map[uint16]uint8
		cycles    uint8
		expect    uint8
		// initialise registers
		y uint8
	}{
		{"immediate", []uint8{0xa2, 0x42}, nil, 2, 0x42, 0},
		{"zeropage", []uint8{0xa6, 0x42}, map[uint16]uint8{0x0042: 0x1}, 3, 0x1, 0},
		{"zeropage,y", []uint8{0xb6, 0x42}, map[uint16]uint8{0x0043: 0x1}, 4, 0x1, 0x1},
		{"absolute", []uint8{0xae, 0x42, 0xaa}, map[uint16]uint8{0xaa42: 0x1}, 4, 0x1, 0},
		{"absolute,y", []uint8{0xbe, 0x42, 0xaa}, map[uint16]uint8{0xaa43: 0x1}, 4, 0x1, 0x1},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			cpu := setup(test.program, test.bootstrap)
			// initialise registers
			cpu.y = test.y
			// run program
			cycle(t, cpu, test.cycles)
			// assertions
			expect8(t, cpu.x, test.expect)
		})
	}
}

func TestSTA(t *testing.T) {
	cases := []struct {
		name    string
		program []uint8
		// initialise bootstrap with memory
		bootstrap     map[uint16]uint8
		cycles        uint8
		expectAddress uint16
		expect        uint16
		// initialise registers
		a uint8
		x uint8
		y uint8
	}{
		{"zeropage", []uint8{0x85, 0x01}, nil, 3, 0x0001, 0x12, 0x12, 0, 0},
		{"zeropage,x", []uint8{0x95, 0x01}, nil, 4, 0x0002, 0x12, 0x12, 0x1, 0},
		{"absolute", []uint8{0x8d, 0xaa, 0xbb}, nil, 4, 0xbbaa, 0x12, 0x12, 0, 0},
		{"absolute,x", []uint8{0x9d, 0xaa, 0xbb}, nil, 5, 0xbbab, 0x12, 0x12, 0x1, 0},
		{"absolute,y", []uint8{0x99, 0xaa, 0xbb}, nil, 5, 0xbbab, 0x12, 0x12, 0, 0x1},
		{"(indirect,x)", []uint8{0x81, 0x70}, map[uint16]uint8{0x0071: 0x0012}, 6, 0x0012, 0x12, 0x12, 0x1, 0},
		{"(indirect),y", []uint8{0x91, 0x70}, map[uint16]uint8{0x0070: 0x0012}, 6, 0x0013, 0x12, 0x12, 0, 0x1},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			cpu := setup(test.program, test.bootstrap)
			// initialise registers
			cpu.a = test.a
			cpu.x = test.x
			cpu.y = test.y
			// run program
			cycle(t, cpu, test.cycles)
			// assertions
			expect16(t, cpu.memory.ReadWord(test.expectAddress), test.expect)
		})
	}
}

func TestCLC(t *testing.T) {
	cases := []struct {
		name   string
		start  bool
		expect bool
	}{
		{"cleared to cleared", false, false},
		{"set to cleared", true, false},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			cpu := setup([]uint8{0x18}, nil)
			if test.start {
				cpu.p.set(P_C)
			} else {
				cpu.p.clear(P_C)
			}
			cycle(t, cpu, 2)
			if cpu.p.isSet(P_C) != test.expect {
				t.Errorf("expected P_C to be %t got %t", test.expect, cpu.p.isSet(P_C))
			}
		})
	}
}

func TestCLD(t *testing.T) {
	cases := []struct {
		name   string
		start  bool
		expect bool
	}{
		{"cleared to cleared", false, false},
		{"set to cleared", true, false},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			cpu := setup([]uint8{0xd8}, nil)
			if test.start {
				cpu.p.set(P_D)
			} else {
				cpu.p.clear(P_D)
			}
			cycle(t, cpu, 2)
			if cpu.p.isSet(P_D) != test.expect {
				t.Errorf("expected P_D to be %t got %t", test.expect, cpu.p.isSet(P_D))
			}
		})
	}
}

func TestCLI(t *testing.T) {
	cases := []struct {
		name   string
		start  bool
		expect bool
	}{
		{"cleared to cleared", false, false},
		{"set to cleared", true, false},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			cpu := setup([]uint8{0x58}, nil)
			if test.start {
				cpu.p.set(P_I)
			} else {
				cpu.p.clear(P_I)
			}
			cycle(t, cpu, 2)
			if cpu.p.isSet(P_I) != test.expect {
				t.Errorf("expected P_I to be %t got %t", test.expect, cpu.p.isSet(P_I))
			}
		})
	}
}

func TestCLV(t *testing.T) {
	cases := []struct {
		name   string
		start  bool
		expect bool
	}{
		{"cleared to cleared", false, false},
		{"set to cleared", true, false},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			cpu := setup([]uint8{0xb8}, nil)
			if test.start {
				cpu.p.set(P_V)
			} else {
				cpu.p.clear(P_V)
			}
			cycle(t, cpu, 2)
			if cpu.p.isSet(P_V) != test.expect {
				t.Errorf("expected P_V to be %t got %t", test.expect, cpu.p.isSet(P_V))
			}
		})
	}
}
