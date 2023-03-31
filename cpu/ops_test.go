package cpu

import (
	"testing"
)

func TestADC(t *testing.T) {
	tests := testCases{
		{
			name:        "add with carry",
			program:     []uint8{0x69, 0x02},
			cycles:      2,
			setupA:      newUint8(0xff),
			expectA:     newUint8(0x01),
			expectCarry: true,
		},
		{
			name:        "add to zero",
			program:     []uint8{0x69, 0x02},
			cycles:      2,
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
			cycles:         2,
		},
		{
			name:    "adds two positive numbers without carry",
			program: []uint8{0x69, 0x0f},
			cycles:  2,
			expectA: newUint8(0x1f),
			setupA:  newUint8(0x10),
		},
		{
			name:    "immediate without carry",
			program: []uint8{0x69, 0x42},
			cycles:  2,
			expectA: newUint8(0x43),
			setupA:  newUint8(0x01),
		},
		{
			name:           "zero page without carry",
			program:        []uint8{0x65, 0x42},
			memory:         map[uint16]uint8{0x42: 0x80},
			cycles:         3,
			expectA:        newUint8(0x81),
			setupA:         newUint8(0x01),
			expectNegative: true,
		},
		{
			name:           "absolute without carry",
			program:        []uint8{0x6d, 0x00, 0x04},
			memory:         map[uint16]uint8{0x0400: 0x42},
			cycles:         4,
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
			cycles:         2,
			expectA:        newUint8(0xAA),
			expectNegative: true,
			setupA:         newUint8(0xFF),
		},
		{
			name:    "zeropage",
			program: []uint8{0x25, 0x42},
			memory:  map[uint16]uint8{0x42: 0x0F},
			cycles:  3,
			expectA: newUint8(0x0E),
			setupA:  newUint8(0xDE),
		},
		{
			name:           "absolute",
			program:        []uint8{0x2D, 0x00, 0x04},
			memory:         map[uint16]uint8{0x0400: 0xF0},
			cycles:         4,
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
			cycles:      2,
		},
		{
			name:       "accumulator 0",
			program:    []uint8{0x0a},
			setupA:     newUint8(0x00),
			expectA:    newUint8(0x00),
			cycles:     2,
			expectZero: true,
		},
		{
			name:           "zeropage",
			program:        []uint8{0x06, 0x42},
			memory:         map[uint16]uint8{0x0042: 0x55},
			cycles:         5,
			expectMemory:   map[uint16]uint8{0x0042: 0xaa},
			expectNegative: true,
		},
		{
			name:           "zeropage,x",
			program:        []uint8{0x16, 0x42},
			memory:         map[uint16]uint8{0x0047: 0x55},
			cycles:         6,
			expectMemory:   map[uint16]uint8{0x0047: 0xaa},
			expectNegative: true,
			setupX:         newUint8(0x5),
		},
		{
			name:           "absolute",
			program:        []uint8{0x0e, 0x42},
			memory:         map[uint16]uint8{0x0042: 0x55},
			cycles:         6,
			expectMemory:   map[uint16]uint8{0x0042: 0xaa},
			expectNegative: true,
		},
		{
			name:           "absolute,x",
			program:        []uint8{0x1e, 0x42},
			memory:         map[uint16]uint8{0x0047: 0x55},
			cycles:         7,
			expectMemory:   map[uint16]uint8{0x0047: 0xaa},
			expectNegative: true,
			setupX:         newUint8(0x5),
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
			cycles:      2,
		},
		{
			name:        "clear unset carry",
			program:     []uint8{0x18},
			expectCarry: false,
			cycles:      2,
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
			cycles:        2,
		},
		{
			name:          "clear unset decimal",
			program:       []uint8{0xd8},
			expectDecimal: newBool(false),
			cycles:        2,
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
			cycles:                 2,
		},
		{
			name:                   "clear unset interrupt",
			program:                []uint8{0x58},
			expectInterruptDisable: newBool(false),
			cycles:                 2,
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
			cycles:         2,
		},
		{
			name:           "clear unset overflow",
			program:        []uint8{0xb8},
			expectOverflow: false,
			cycles:         2,
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
			cycles:  2,
		},
		{
			name:    "inx 0aa",
			program: []uint8{0xe8},
			setupX:  newUint8(0x0a),
			expectX: newUint8(0x0b),
			cycles:  2,
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
			cycles:  2,
		},
		{
			name:    "iny 0aa",
			program: []uint8{0xc8},
			expectY: newUint8(0x0b),
			setupY:  newUint8(0x0a),
			cycles:  2,
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
			cycles:       5,
			expectMemory: map[uint16]uint8{0x0042: 0x0a},
		},
		{
			name:         "zeropage,x",
			program:      []uint8{0xf6, 0x42},
			memory:       map[uint16]uint8{0x0043: 0x09},
			cycles:       6,
			expectMemory: map[uint16]uint8{0x0043: 0x0a},
			setupX:       newUint8(0x1),
		},
		{
			name:         "absolute",
			program:      []uint8{0xee, 0x42, 0xaa},
			memory:       map[uint16]uint8{0xaa42: 0x09},
			cycles:       6,
			expectMemory: map[uint16]uint8{0xaa42: 0x0a},
		},
		{
			name:         "absolute,x",
			program:      []uint8{0xfe, 0x42, 0xaa},
			memory:       map[uint16]uint8{0xaa43: 0x09},
			cycles:       7,
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
			cycles:   3,
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
			cycles:   5,
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
				0x01fd: 0x01,
				0x01fc: 0x04,
			},
			cycles: 6,
		},
	}

	tests.run(t)
}

func TestLDA(t *testing.T) {
	tests := testCases{
		{
			name:    "immediate",
			program: []uint8{0xa9, 0x42},
			cycles:  2,
			expectA: newUint8(0x42),
		},
		{
			name:       "immediate, with zero",
			program:    []uint8{0xa9, 0x00},
			cycles:     2,
			expectA:    newUint8(0x00),
			expectZero: true,
		},
		{
			name:           "zeropage",
			program:        []uint8{0xa5, 0x01},
			memory:         map[uint16]uint8{0x01: 0x99},
			cycles:         3,
			expectA:        newUint8(0x99),
			expectNegative: true,
		},
		{
			name:    "zeropage,x(x=0)",
			program: []uint8{0xb5, 0x80},
			memory:  map[uint16]uint8{0x0080: 0x40},
			cycles:  4,
			expectA: newUint8(0x40),
		},
		{
			name:    "zeropage,x(x=0x02)",
			program: []uint8{0xb5, 0x80},
			memory:  map[uint16]uint8{0x82: 0x40},
			cycles:  4,
			setupX:  newUint8(0x02),
			expectA: newUint8(0x40),
		},
		{
			name:    "absolute",
			program: []uint8{0xad, 0x10, 0x30},
			memory:  map[uint16]uint8{0x3010: 0x22},
			cycles:  4,
			expectA: newUint8(0x22),
		},
		{
			name:    "absolute,x(x=0)",
			program: []uint8{0xbd, 0x10, 0x30},
			memory:  map[uint16]uint8{0x3010: 0x22},
			cycles:  4,
			expectA: newUint8(0x22),
		},
		{
			name:    "absolute,x(x=2)",
			program: []uint8{0xbd, 0x10, 0x30},
			memory:  map[uint16]uint8{0x3012: 0x22},
			cycles:  4,
			setupX:  newUint8(0x02),
			expectA: newUint8(0x22),
		},
		{
			name:    "absolute,y(y=0)",
			program: []uint8{0xb9, 0x10, 0x30},
			memory:  map[uint16]uint8{0x3010: 0x22},
			cycles:  4,
			expectA: newUint8(0x22),
		},
		{
			name:    "absolute,y(y=2)",
			program: []uint8{0xb9, 0x10, 0x30},
			memory:  map[uint16]uint8{0x3012: 0x22},
			cycles:  4,
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
			cycles:         6,
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
			cycles:  6,
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
			cycles:  2,
			expectX: newUint8(0x42),
		},
		{
			name:    "zeropage",
			program: []uint8{0xa6, 0x42},
			memory:  map[uint16]uint8{0x0042: 0x1},
			cycles:  3,
			expectX: newUint8(0x1),
		},
		{
			name:    "zeropage,y",
			program: []uint8{0xb6, 0x42},
			memory:  map[uint16]uint8{0x0043: 0x1},
			cycles:  4,
			expectX: newUint8(0x1),
			setupY:  newUint8(0x1),
		},
		{
			name:    "absolute",
			program: []uint8{0xae, 0x42, 0xaa},
			memory:  map[uint16]uint8{0xaa42: 0x1},
			cycles:  4,
			expectX: newUint8(0x1),
		},
		{
			name:    "absolute,y",
			program: []uint8{0xbe, 0x42, 0xaa},
			memory:  map[uint16]uint8{0xaa43: 0x1},
			cycles:  4,
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
			cycles:  2,
			expectY: newUint8(0x42),
		},
		{
			name:    "zeropage",
			program: []uint8{0xa4, 0x42},
			memory:  map[uint16]uint8{0x0042: 0x1},
			cycles:  3,
			expectY: newUint8(0x1),
		},
		{
			name:    "zeropage,y",
			program: []uint8{0xb4, 0x42},
			memory:  map[uint16]uint8{0x0043: 0x1},
			cycles:  4,
			setupY:  newUint8(0x1),
			expectY: newUint8(0x1),
		},
		{
			name:    "absolute",
			program: []uint8{0xac, 0x42, 0xaa},
			memory:  map[uint16]uint8{0xaa42: 0x1},
			cycles:  4,
			expectY: newUint8(0x1),
		},
		{
			name:    "absolute,y",
			program: []uint8{0xbc, 0x42, 0xaa},
			memory:  map[uint16]uint8{0xaa43: 0x1},
			cycles:  4,
			expectY: newUint8(0x1),
			setupY:  newUint8(0x1),
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
			cycles:  2,
		},
		{
			name:       "accumulator 0",
			program:    []uint8{0x4a},
			setupA:     newUint8(0x00),
			expectA:    newUint8(0x00),
			cycles:     2,
			expectZero: true,
		},
		{
			name:         "zeropage",
			program:      []uint8{0x46, 0x42},
			memory:       map[uint16]uint8{0x0042: 0x55},
			cycles:       5,
			expectMemory: map[uint16]uint8{0x0042: 0x2a},
		},
		{
			name:         "zeropage,x",
			program:      []uint8{0x56, 0x42},
			memory:       map[uint16]uint8{0x0047: 0x55},
			cycles:       6,
			expectMemory: map[uint16]uint8{0x0047: 0x2a},
			setupX:       newUint8(0x5),
		},
		{
			name:         "absolute",
			program:      []uint8{0x4e, 0x42},
			memory:       map[uint16]uint8{0x0042: 0x55},
			cycles:       6,
			expectMemory: map[uint16]uint8{0x0042: 0x2a},
		},
		{
			name:         "absolute,x",
			program:      []uint8{0x5e, 0x42},
			memory:       map[uint16]uint8{0x0047: 0x55},
			cycles:       7,
			expectMemory: map[uint16]uint8{0x0047: 0x2a},
			setupX:       newUint8(0x5),
		},
	}

	tests.run(t)
}

func TestNOP(t *testing.T) {
	tests := testCases{
		{
			name:    "implied",
			program: []uint8{0xea},
			cycles:  2,
		},
	}

	tests.run(t)
}

func TestORA(t *testing.T) {
	tests := testCases{
		{
			name:    "immediate",
			program: []uint8{0x09, 0x42},
			cycles:  2,
			setupA:  newUint8(0x10),
			expectA: newUint8(0x52),
		},
		{
			name:    "zeropage",
			program: []uint8{0x05, 0x42},
			cycles:  3,
			memory:  map[uint16]uint8{0x0042: 0x42},
			setupA:  newUint8(0x10),
			expectA: newUint8(0x52),
		},
		{
			name:    "zeropage,x",
			program: []uint8{0x15, 0x42},
			cycles:  4,
			memory:  map[uint16]uint8{0x0043: 0x42},
			setupA:  newUint8(0x10),
			setupX:  newUint8(0x01),
			expectA: newUint8(0x52),
		},
		{
			name:    "absolute",
			program: []uint8{0x0d, 0x42, 0xaa},
			cycles:  4,
			memory:  map[uint16]uint8{0xaa42: 0x42},
			setupA:  newUint8(0x10),
			expectA: newUint8(0x52),
		},
		{
			name:    "absolute,x",
			program: []uint8{0x1d, 0x42, 0xaa},
			cycles:  4,
			memory:  map[uint16]uint8{0xaa43: 0x42},
			setupA:  newUint8(0x10),
			setupX:  newUint8(0x01),
			expectA: newUint8(0x52),
		},
		{
			name:    "absolute,y",
			program: []uint8{0x19, 0x42, 0xaa},
			cycles:  4,
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
			cycles:  6,
			setupA:  newUint8(0x10),
			setupX:  newUint8(0x01),
			expectA: newUint8(0x52),
		},
		{
			name:    "(indirect),y",
			program: []uint8{0x11, 0xaa},
			cycles:  6,
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

func TestSTA(t *testing.T) {
	tests := testCases{
		{
			name:         "zeropage",
			program:      []uint8{0x85, 0x01},
			cycles:       3,
			setupA:       newUint8(0x12),
			expectMemory: map[uint16]uint8{0x0001: 0x12},
		},
		{
			name:         "zeropage,x",
			program:      []uint8{0x95, 0x01},
			cycles:       4,
			setupA:       newUint8(0x12),
			setupX:       newUint8(0x1),
			expectMemory: map[uint16]uint8{0x0002: 0x12},
		},
		{
			name:         "absolute",
			program:      []uint8{0x8d, 0xaa, 0xbb},
			cycles:       4,
			setupA:       newUint8(0x12),
			expectMemory: map[uint16]uint8{0xbbaa: 0x12},
		},
		{
			name:         "absolute,x",
			program:      []uint8{0x9d, 0xaa, 0xbb},
			cycles:       5,
			setupA:       newUint8(0x12),
			setupX:       newUint8(0x1),
			expectMemory: map[uint16]uint8{0xbbab: 0x12},
		},
		{
			name:         "absolute,y",
			program:      []uint8{0x99, 0xaa, 0xbb},
			cycles:       5,
			setupA:       newUint8(0x12),
			setupY:       newUint8(0x1),
			expectMemory: map[uint16]uint8{0xbbab: 0x12},
		},
		{
			name:         "(indirect,x)",
			program:      []uint8{0x81, 0x70},
			memory:       map[uint16]uint8{0x0071: 0x0012},
			cycles:       6,
			setupA:       newUint8(0x12),
			setupX:       newUint8(0x1),
			expectMemory: map[uint16]uint8{0x0012: 0x12},
		},
		{
			name:         "(indirect),y",
			program:      []uint8{0x91, 0x70},
			memory:       map[uint16]uint8{0x0070: 0x0012},
			cycles:       6,
			setupA:       newUint8(0x12),
			setupY:       newUint8(0x1),
			expectMemory: map[uint16]uint8{0x0013: 0x12},
		},
	}

	tests.run(t)
}
