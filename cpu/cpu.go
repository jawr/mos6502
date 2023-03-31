package cpu

import "fmt"

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
