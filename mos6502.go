package main

import (
	"fmt"
)

/*
65k of memory, 256 pages with
256 bytes per page

addresses are 16 bits in length with the
first byte referencing the page (up to 255)
and the second byte referncing the offset
on the page (up to 255)

1
2 6 3 1
8 4 2 6 8 4 2 1
---------------
0 0 0 0 0 0 0 0

page 0 aka Zero Page is a special
page for quick access as accessing
it only requires 1 byte address rather than
two

the last six bytes of the last page (page 255)
have special addresses but are still considered ROM:

- interrupt handlers for (IRC and NMI)
- reset handler (starting point for the processor)

overall the lower half of memory is RAM and the
upper half is ROM
*/
type Memory [0x100 * 0x100]uint8

func (m *Memory) Read(address uint16) uint8 {
	// reads a 1 byte address
	return m[address]
}

func (m *Memory) ReadWord(address uint16) uint16 {
	// takes a 2 byte address and returns a 2 byte address
	return uint16(m[address]) + (uint16(m[address+1]) << 8)
}

type AddressMode uint8

const (
	AM_IMPLIED AddressMode = iota
	AM_IMMEDIATE
	AM_ABSOLUTE
	AM_ZEROPAGE
	AM_INDEXED_X
	AM_INDEXED_Y
	AM_ZEROPAGE_X
	AM_ZEROPAGE_Y
	AM_INDIRECT
	AM_PRE_INDEXED
	AM_POST_INDEXED
	AM_RELATIVE
)

type Instruction struct {
	name    string
	cycles  uint8
	size    uint8 // number of bytes to load
	execute func(uint16)
	mode    AddressMode
}

func (i *Instruction) parseOperand(cpu *MOS6502) uint16 {
	switch i.mode {
	case AM_IMPLIED:
		// single byte instructrions
		return 0

	case AM_IMMEDIATE:
		// literal operand loaded into memory
		// always an 8 bit value
		return cpu.pc

	case AM_ABSOLUTE:
		// full 16 bit address in LLHH format
		lo := cpu.memory.Read(cpu.pc)
		hi := cpu.memory.Read(cpu.pc + 1)

		return (uint16(hi) << 8) + uint16(lo)

	case AM_ZEROPAGE:
		// 1 byte address in the zeropage (high byte is 0x00)
		return uint16(cpu.memory.Read(cpu.pc))

	case AM_ZEROPAGE_X:
		// first byte comes from pc
		address := cpu.memory.Read(cpu.pc)
		// add contents of x register
		address += cpu.x
		// address will have wrapped around meaning we stay
		// in zp
		return uint16(address)

	case AM_ZEROPAGE_Y:
		// first byte comes from pc
		address := cpu.memory.Read(cpu.pc)
		// add contents of y register
		address += cpu.y
		// address will have wrapped around meaning we stay
		// in zp
		return uint16(address)

	case AM_INDEXED_X:
		// read 16 bit address in LLHH format
		lo := cpu.memory.Read(cpu.pc)
		hi := cpu.memory.Read(cpu.pc + 1)

		address := (uint16(hi) << 8) + uint16(lo)
		address += uint16(cpu.x)

		return uint16(address)

	case AM_INDEXED_Y:
		// read 16 bit address in LLHH format
		lo := cpu.memory.Read(cpu.pc)
		hi := cpu.memory.Read(cpu.pc + 1)

		address := (uint16(hi) << 8) + uint16(lo)
		address += uint16(cpu.y)

		return uint16(address)

	case AM_PRE_INDEXED:
		// first byte comes from pc
		address := cpu.memory.Read(cpu.pc)

		// add contents of x register
		address += cpu.x

		// get the lookup from this address
		lookup := cpu.memory.ReadWord(uint16(address))

		// resolve the lookup
		return lookup

	case AM_POST_INDEXED:
		// first byte comes from pc
		address := cpu.memory.Read(cpu.pc)

		// get the lookup from zeropage
		lookup := cpu.memory.ReadWord(uint16(address))

		// add contents of y register
		lookup += uint16(cpu.y)

		// resolve the lookup
		return lookup

	case AM_INDIRECT:
		// get the indirect address
		lo := cpu.memory.Read(cpu.pc)
		hi := cpu.memory.Read(cpu.pc + 1)

		address := (uint16(hi) << 8) + uint16(lo)

		// read the address from the indirect address
		return cpu.memory.ReadWord(address)

	default:
		panic("unsupported address mode")
	}
}

type MOS6502 struct {
	// main register
	a uint8
	// index registers
	x uint8
	y uint8

	// stack pointer
	sp uint8

	// program counter
	pc uint16

	// status register (https://www.masswerk.at/6502/6502_instruction_set.html)
	// N -> Sign
	// V -> Overflow
	// - -> Reserved
	// B -> Break
	// D -> Decimal
	// I -> Interrupt Disable
	// Z -> Zero
	// C -> Carry
	p flags

	// operations take a predetermined amount of time
	wait uint8

	// instruction table
	instructions [0x100]*Instruction

	// memory thats set on reset
	memory *Memory
}

func NewMOS6502() *MOS6502 {
	cpu := MOS6502{}

	// ADC
	cpu.instructions[0x69] = &Instruction{name: "ADC", cycles: 2, execute: cpu.adc, size: 2, mode: AM_IMMEDIATE}
	cpu.instructions[0x65] = &Instruction{name: "ADC", cycles: 3, execute: cpu.adc, size: 2, mode: AM_ZEROPAGE}
	cpu.instructions[0x75] = &Instruction{name: "ADC", cycles: 4, execute: cpu.adc, size: 2, mode: AM_ZEROPAGE_X}
	cpu.instructions[0x6d] = &Instruction{name: "ADC", cycles: 4, execute: cpu.adc, size: 3, mode: AM_ABSOLUTE}
	cpu.instructions[0x7d] = &Instruction{name: "ADC", cycles: 4, execute: cpu.adc, size: 3, mode: AM_INDEXED_X}
	cpu.instructions[0x79] = &Instruction{name: "ADC", cycles: 4, execute: cpu.adc, size: 3, mode: AM_INDEXED_Y}
	cpu.instructions[0x61] = &Instruction{name: "ADC", cycles: 6, execute: cpu.adc, size: 2, mode: AM_PRE_INDEXED}
	cpu.instructions[0x71] = &Instruction{name: "ADC", cycles: 5, execute: cpu.adc, size: 2, mode: AM_POST_INDEXED}

	// AND
	cpu.instructions[0x29] = &Instruction{name: "AND", cycles: 2, execute: cpu.and, size: 2, mode: AM_IMMEDIATE}
	cpu.instructions[0x25] = &Instruction{name: "AND", cycles: 3, execute: cpu.and, size: 2, mode: AM_ZEROPAGE}
	cpu.instructions[0x35] = &Instruction{name: "AND", cycles: 4, execute: cpu.and, size: 2, mode: AM_ZEROPAGE_X}
	cpu.instructions[0x2d] = &Instruction{name: "AND", cycles: 4, execute: cpu.and, size: 3, mode: AM_ABSOLUTE}
	cpu.instructions[0x3d] = &Instruction{name: "AND", cycles: 4, execute: cpu.and, size: 3, mode: AM_INDEXED_X}
	cpu.instructions[0x39] = &Instruction{name: "AND", cycles: 4, execute: cpu.and, size: 3, mode: AM_INDEXED_Y}
	cpu.instructions[0x21] = &Instruction{name: "AND", cycles: 6, execute: cpu.and, size: 2, mode: AM_PRE_INDEXED}
	cpu.instructions[0x31] = &Instruction{name: "AND", cycles: 5, execute: cpu.and, size: 2, mode: AM_POST_INDEXED}

	// ASL

	// CLC
	cpu.instructions[0x18] = &Instruction{name: "CLC", cycles: 2, execute: cpu.clc, size: 1, mode: AM_IMPLIED}

	// CLD
	cpu.instructions[0xd8] = &Instruction{name: "CLD", cycles: 2, execute: cpu.cld, size: 1, mode: AM_IMPLIED}

	// CLI
	cpu.instructions[0x58] = &Instruction{name: "CLI", cycles: 2, execute: cpu.cli, size: 1, mode: AM_IMPLIED}

	// CLV
	cpu.instructions[0xb8] = &Instruction{name: "CLV", cycles: 2, execute: cpu.clv, size: 1, mode: AM_IMPLIED}

	// INC
	cpu.instructions[0xe6] = &Instruction{name: "INC", cycles: 5, execute: cpu.inc, size: 2, mode: AM_ZEROPAGE}
	cpu.instructions[0xf6] = &Instruction{name: "INC", cycles: 6, execute: cpu.inc, size: 2, mode: AM_ZEROPAGE_X}
	cpu.instructions[0xee] = &Instruction{name: "INC", cycles: 6, execute: cpu.inc, size: 3, mode: AM_ABSOLUTE}
	cpu.instructions[0xfe] = &Instruction{name: "INC", cycles: 7, execute: cpu.inc, size: 3, mode: AM_INDEXED_X}

	// INX
	cpu.instructions[0xe8] = &Instruction{name: "INX", cycles: 2, execute: cpu.inx, size: 1, mode: AM_IMPLIED}

	// INY
	cpu.instructions[0xc8] = &Instruction{name: "INY", cycles: 2, execute: cpu.iny, size: 1, mode: AM_IMPLIED}

	// JMP
	cpu.instructions[0x4c] = &Instruction{name: "JMP", cycles: 3, execute: cpu.jmp, size: 3, mode: AM_ABSOLUTE}
	cpu.instructions[0x6c] = &Instruction{name: "JMP", cycles: 5, execute: cpu.jmp, size: 3, mode: AM_INDIRECT}

	// JSR
	cpu.instructions[0x20] = &Instruction{name: "JSR", cycles: 6, execute: cpu.jsr, size: 3, mode: AM_ABSOLUTE}

	// NOP
	cpu.instructions[0xea] = &Instruction{
		name:    "NOP",
		cycles:  2,
		execute: cpu.nop,
		size:    1,
		mode:    AM_IMPLIED,
	}

	// LDA
	cpu.instructions[0xa9] = &Instruction{
		name:    "LDA",
		cycles:  2,
		execute: cpu.lda,
		size:    2,
		mode:    AM_IMMEDIATE,
	}
	cpu.instructions[0xa5] = &Instruction{
		name:    "LDA",
		cycles:  3,
		execute: cpu.lda,
		size:    2,
		mode:    AM_ZEROPAGE,
	}
	cpu.instructions[0xb5] = &Instruction{
		name:    "LDA",
		cycles:  4,
		execute: cpu.lda,
		size:    2,
		mode:    AM_ZEROPAGE_X,
	}
	cpu.instructions[0xad] = &Instruction{
		name:    "LDA",
		cycles:  4,
		execute: cpu.lda,
		size:    3,
		mode:    AM_ABSOLUTE,
	}
	cpu.instructions[0xbd] = &Instruction{
		name:    "LDA",
		cycles:  4,
		execute: cpu.lda,
		size:    3,
		mode:    AM_INDEXED_X,
	}
	cpu.instructions[0xb9] = &Instruction{
		name:    "LDA",
		cycles:  4,
		execute: cpu.lda,
		size:    3,
		mode:    AM_INDEXED_Y,
	}
	cpu.instructions[0xa1] = &Instruction{
		name:    "LDA",
		cycles:  6,
		execute: cpu.lda,
		size:    2,
		mode:    AM_PRE_INDEXED,
	}
	cpu.instructions[0xb1] = &Instruction{
		name:    "LDA",
		cycles:  5,
		execute: cpu.lda,
		size:    2,
		mode:    AM_POST_INDEXED,
	}

	// LDX
	cpu.instructions[0xa2] = &Instruction{name: "LDX", cycles: 2, execute: cpu.ldx, size: 2, mode: AM_IMMEDIATE}
	cpu.instructions[0xa6] = &Instruction{name: "LDX", cycles: 3, execute: cpu.ldx, size: 2, mode: AM_ZEROPAGE}
	cpu.instructions[0xb6] = &Instruction{name: "LDX", cycles: 4, execute: cpu.ldx, size: 2, mode: AM_ZEROPAGE_Y}
	cpu.instructions[0xae] = &Instruction{name: "LDX", cycles: 4, execute: cpu.ldx, size: 3, mode: AM_ABSOLUTE}
	cpu.instructions[0xbe] = &Instruction{name: "LDX", cycles: 4, execute: cpu.ldx, size: 3, mode: AM_INDEXED_Y}

	// LDY
	cpu.instructions[0xa0] = &Instruction{name: "LDY", cycles: 2, execute: cpu.ldy, size: 2, mode: AM_IMMEDIATE}
	cpu.instructions[0xa4] = &Instruction{name: "LDY", cycles: 3, execute: cpu.ldy, size: 2, mode: AM_ZEROPAGE}
	cpu.instructions[0xb4] = &Instruction{name: "LDY", cycles: 4, execute: cpu.ldy, size: 2, mode: AM_ZEROPAGE_Y}
	cpu.instructions[0xac] = &Instruction{name: "LDY", cycles: 4, execute: cpu.ldy, size: 3, mode: AM_ABSOLUTE}
	cpu.instructions[0xbc] = &Instruction{name: "LDY", cycles: 4, execute: cpu.ldy, size: 3, mode: AM_INDEXED_Y}

	// STA
	cpu.instructions[0x85] = &Instruction{name: "STA", cycles: 3, execute: cpu.sta, size: 2, mode: AM_ZEROPAGE}
	cpu.instructions[0x95] = &Instruction{name: "STA", cycles: 4, execute: cpu.sta, size: 2, mode: AM_ZEROPAGE_X}
	cpu.instructions[0x8d] = &Instruction{name: "STA", cycles: 4, execute: cpu.sta, size: 3, mode: AM_ABSOLUTE}
	cpu.instructions[0x9d] = &Instruction{name: "STA", cycles: 5, execute: cpu.sta, size: 3, mode: AM_INDEXED_X}
	cpu.instructions[0x99] = &Instruction{name: "STA", cycles: 5, execute: cpu.sta, size: 3, mode: AM_INDEXED_Y}
	cpu.instructions[0x81] = &Instruction{name: "STA", cycles: 6, execute: cpu.sta, size: 2, mode: AM_PRE_INDEXED}
	cpu.instructions[0x91] = &Instruction{name: "STA", cycles: 6, execute: cpu.sta, size: 2, mode: AM_POST_INDEXED}

	return &cpu
}

func (cpu *MOS6502) Reset(memory *Memory) {
	// reset registers
	cpu.a = 0xaa
	cpu.x = 0x0
	cpu.y = 0x0
	// reset stack pointer
	cpu.sp = 0xfd
	// reset flags
	cpu.p = 0b00110100

	cpu.pc = memory.ReadWord(0xfffc)

	cpu.memory = memory
	cpu.wait = 0
}

func (cpu *MOS6502) Cycle() {
	if cpu.wait > 0 {
		cpu.wait--
		return
	}

	// pop the 8bit opcode and progress the pc
	opcode := cpu.memory.Read(cpu.pc)
	cpu.pc++

	instruction := cpu.instructions[opcode]
	if instruction == nil {
		return
	}

	operand := instruction.parseOperand(cpu)

	fmt.Printf("opcode=%02x operand=%04x pc=%04x a=%02x x=%02x y=%02x\n", opcode, operand, cpu.pc, cpu.a, cpu.x, cpu.y)

	// increment the pc by the number of bytes read for the operand
	cpu.pc += uint16(instruction.size - 1)
	// mark the cpu busy for the number of cycles the instruction takes
	cpu.wait = instruction.cycles - 1

	instruction.execute(operand)
}

// flags
const (
	P_C flag = 1 << iota
	P_Z
	P_I
	P_D
	P_B
	P_
	P_V
	P_N
)

type flag uint8
type flags uint8

func (a flags) isSet(b flag) bool {
	return uint8(a)&uint8(b) != 0x0
}

func (a *flags) set(b flag) {
	*a = flags(uint8(*a) | uint8(b))
}

func (a *flags) clear(b flag) {
	*a = flags(uint8(*a) &^ uint8(b))
}

func (cpu *MOS6502) testAndSetNegative(b uint8) {
	if b&0x80 == 0x80 {
		cpu.p.set(P_N)
	} else {
		cpu.p.clear(P_N)
	}
}

func (cpu *MOS6502) testAndSetZero(b uint8) {
	if b == 0x0 {
		cpu.p.set(P_Z)
	} else {
		cpu.p.clear(P_Z)
	}
}

func (cpu *MOS6502) testAndSetCarry(b uint16) {
	// if b is bigger than an 8bit set the carry
	if b > 0xff {
		cpu.p.set(P_C)
	} else {
		cpu.p.clear(P_C)
	}
}

func (cpu *MOS6502) testAndSetOverflow(a, b, sum uint16) {
	// Calculate the overflow by checking if the sign bit of the operands and
	// the result differ (which indicates a signed overflow has occurred).
	aSign := a & 0x80
	bSign := b & 0x80
	sumSign := sum & 0x80
	overflow := (aSign == bSign) && (aSign != sumSign)

	if overflow {
		cpu.p.set(P_N)
	} else {
		cpu.p.clear(P_N)
	}
}

// operations

func (cpu *MOS6502) adc(address uint16) {
	// Add Memory to Accumulator with Carry
	// A + M + C -> A, C
	var c uint16 = 0
	if cpu.p.isSet(P_C) {
		c = 1
	}

	a := uint16(cpu.a)
	m := uint16(cpu.memory.Read(address))

	// sum in uint16 to catch overflow
	sum := a + m + c

	cpu.a = uint8(sum & 0xff)
	cpu.testAndSetCarry(sum)
	cpu.testAndSetNegative(cpu.a)
	cpu.testAndSetZero(cpu.a)
	cpu.testAndSetOverflow(a, m, sum)
}

func (cpu *MOS6502) and(address uint16) {
	b := cpu.memory.Read(address)
	fmt.Printf("and %04x v=%02x & a=%02x\n", address, b, cpu.a)
	cpu.a = cpu.a & b
	cpu.testAndSetNegative(cpu.a)
	cpu.testAndSetZero(cpu.a)
}

func (cpu *MOS6502) clc(address uint16) {
	// Clear Carry Flag
	cpu.p.clear(P_C)
}

func (cpu *MOS6502) cld(address uint16) {
	// Clear Decimal Mode
	cpu.p.clear(P_D)
}

func (cpu *MOS6502) cli(address uint16) {
	// Clear Interrupt Disable Bit
	cpu.p.clear(P_I)
}

func (cpu *MOS6502) clv(address uint16) {
	// Clear Overflow Flag
	cpu.p.clear(P_V)
}

func (cpu *MOS6502) inx(address uint16) {
	// Increment Index X by One
	cpu.x++
	cpu.testAndSetNegative(cpu.x)
	cpu.testAndSetZero(cpu.x)
}

func (cpu *MOS6502) iny(address uint16) {
	// Increment Index Y by One
	cpu.y++
	cpu.testAndSetNegative(cpu.y)
	cpu.testAndSetZero(cpu.y)
}

func (cpu *MOS6502) inc(address uint16) {
	// Increment Memory by One
	cpu.memory[address]++
	value := cpu.memory.Read(address)
	cpu.testAndSetNegative(value)
	cpu.testAndSetZero(value)
}

func (cpu *MOS6502) jmp(address uint16) {
	// Jump to New Location
	cpu.pc = address
}

func (cpu *MOS6502) jsr(address uint16) {
	panic("not implemented")
}

func (cpu *MOS6502) nop(address uint16) {
	// No Operation
}

func (cpu *MOS6502) lda(address uint16) {
	// Load Accumulator with Memory
	value := cpu.memory.ReadWord(address)
	cpu.a = uint8(value)
	cpu.testAndSetNegative(cpu.a)
	cpu.testAndSetZero(cpu.a)
}

func (cpu *MOS6502) ldx(address uint16) {
	// Load Index X with Memory
	value := cpu.memory.ReadWord(address)
	cpu.x = uint8(value)
	cpu.testAndSetNegative(cpu.x)
	cpu.testAndSetZero(cpu.x)
}

func (cpu *MOS6502) ldy(address uint16) {
	// Load Index X with Memory
	value := cpu.memory.ReadWord(address)
	cpu.y = uint8(value)
	cpu.testAndSetNegative(cpu.y)
	cpu.testAndSetZero(cpu.y)
}

func (cpu *MOS6502) sta(address uint16) {
	// Store Accumulator in Memory
	cpu.memory[address] = cpu.a
}
