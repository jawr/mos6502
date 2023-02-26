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

type InstructionName string

const (
	LDA InstructionName = "LDA"
	NOP InstructionName = "NOP"
)

type Instruction struct {
	name    InstructionName
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
	// setup instructions table
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

	// CLC
	cpu.instructions[0x18] = &Instruction{name: "CLC", cycles: 2, execute: cpu.clc, size: 1, mode: AM_IMPLIED}

	// CLD
	cpu.instructions[0xd8] = &Instruction{name: "CLD", cycles: 2, execute: cpu.cld, size: 1, mode: AM_IMPLIED}

	// CLI
	cpu.instructions[0x58] = &Instruction{name: "CLI", cycles: 2, execute: cpu.cli, size: 1, mode: AM_IMPLIED}

	// CLV
	cpu.instructions[0xb8] = &Instruction{name: "CLV", cycles: 2, execute: cpu.clv, size: 1, mode: AM_IMPLIED}

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
		fmt.Printf("unknown opcode %02x\n", opcode)
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
	if b&0b10000000 == 0b10000000 {
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

// operations
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
	cpu.testAndSetNegative(cpu.a)
	cpu.testAndSetZero(cpu.a)
}

func (cpu *MOS6502) ldy(address uint16) {
	// Load Index X with Memory
	value := cpu.memory.ReadWord(address)
	cpu.y = uint8(value)
	cpu.testAndSetNegative(cpu.a)
	cpu.testAndSetZero(cpu.a)
}

func (cpu *MOS6502) sta(address uint16) {
	// Store Accumulator in Memory
	cpu.memory[address] = cpu.a
}

func (cpu *MOS6502) nop(address uint16) {}

func (cpu *MOS6502) clc(address uint16) {
	cpu.p.clear(P_C)
}

func (cpu *MOS6502) cld(address uint16) {
	cpu.p.clear(P_D)
}

func (cpu *MOS6502) cli(address uint16) {
	cpu.p.clear(P_I)
}

func (cpu *MOS6502) clv(address uint16) {
	cpu.p.clear(P_V)
}
