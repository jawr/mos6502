package cpu

import "fmt"

const (
	// NMI (Non-Maskable Interrupt) vector
	NMIVectorLow  uint16 = 0xfffa
	NMIVectorHigh uint16 = 0xfffb
	// RES (Reset) vector
	RESVectorLow  uint16 = 0xfffc
	RESVectorHigh uint16 = 0xfffd
	// IRQ (Interrupt Request) vector
	IRQVectorLow  uint16 = 0xfffe
	IRQVectorHigh uint16 = 0xffff
	// Stack pointer start
	StackBottom uint16 = 0x0100
	StackTop    uint16 = 0x01ff
)

type MOS6502 struct {
	// main register
	a uint8
	// index registers
	x uint8
	y uint8

	// stack pointer
	// this is actullay an 8 bit register masked to
	// 0x0100 but we use a 16 bit for convenience
	sp uint16

	// program counter
	pc uint16

	// status register (https://www.masswerk.at/6502/6502_instruction_set.html)
	// N -> Sign/Negative
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
	instructions [0x100]*instruction

	// memory thats set on reset
	memory *Memory
}

func NewMOS6502() *MOS6502 {
	cpu := MOS6502{}

	// ADC
	cpu.instructions[0x69] = NewInstruction(OPC_ADC, 2, 2, cpu.adc, AM_IMMEDIATE)
	cpu.instructions[0x65] = NewInstruction(OPC_ADC, 3, 2, cpu.adc, AM_ZEROPAGE)
	cpu.instructions[0x75] = NewInstruction(OPC_ADC, 4, 2, cpu.adc, AM_ZEROPAGE_X)
	cpu.instructions[0x6d] = NewInstruction(OPC_ADC, 4, 3, cpu.adc, AM_ABSOLUTE)
	cpu.instructions[0x7d] = NewInstruction(OPC_ADC, 4, 3, cpu.adc, AM_INDEXED_X)
	cpu.instructions[0x79] = NewInstruction(OPC_ADC, 4, 3, cpu.adc, AM_INDEXED_Y)
	cpu.instructions[0x61] = NewInstruction(OPC_ADC, 6, 2, cpu.adc, AM_PRE_INDEXED)
	cpu.instructions[0x71] = NewInstruction(OPC_ADC, 5, 2, cpu.adc, AM_POST_INDEXED)

	// AND
	cpu.instructions[0x29] = NewInstruction(OPC_AND, 2, 2, cpu.and, AM_IMMEDIATE)
	cpu.instructions[0x25] = NewInstruction(OPC_AND, 3, 2, cpu.and, AM_ZEROPAGE)
	cpu.instructions[0x35] = NewInstruction(OPC_AND, 4, 2, cpu.and, AM_ZEROPAGE_X)
	cpu.instructions[0x2d] = NewInstruction(OPC_AND, 4, 3, cpu.and, AM_ABSOLUTE)
	cpu.instructions[0x3d] = NewInstruction(OPC_AND, 4, 3, cpu.and, AM_INDEXED_X)
	cpu.instructions[0x39] = NewInstruction(OPC_AND, 4, 3, cpu.and, AM_INDEXED_Y)
	cpu.instructions[0x21] = NewInstruction(OPC_AND, 6, 2, cpu.and, AM_PRE_INDEXED)
	cpu.instructions[0x31] = NewInstruction(OPC_AND, 5, 2, cpu.and, AM_POST_INDEXED)

	// ASL
	cpu.instructions[0x0a] = NewInstruction(OPC_ASL, 2, 1, cpu.asl, AM_IMPLIED)
	cpu.instructions[0x06] = NewInstruction(OPC_ASL, 5, 2, cpu.asl, AM_ZEROPAGE)
	cpu.instructions[0x16] = NewInstruction(OPC_ASL, 6, 2, cpu.asl, AM_ZEROPAGE_X)
	cpu.instructions[0x0e] = NewInstruction(OPC_ASL, 6, 3, cpu.asl, AM_ABSOLUTE)
	cpu.instructions[0x1e] = NewInstruction(OPC_ASL, 7, 3, cpu.asl, AM_INDEXED_X)

	// BCC
	cpu.instructions[0x90] = NewInstruction(OPC_BCC, 2, 2, cpu.bcc, AM_RELATIVE)

	// BCS
	cpu.instructions[0xb0] = NewInstruction(OPC_BCS, 2, 2, cpu.bcs, AM_RELATIVE)

	// BEQ
	cpu.instructions[0xf0] = NewInstruction(OPC_BEQ, 2, 2, cpu.beq, AM_RELATIVE)

	// BIT
	cpu.instructions[0x24] = NewInstruction(OPC_BIT, 3, 2, cpu.bit, AM_ZEROPAGE)
	cpu.instructions[0x2c] = NewInstruction(OPC_BIT, 4, 3, cpu.bit, AM_ABSOLUTE)

	// BMI
	cpu.instructions[0x30] = NewInstruction(OPC_BMI, 2, 2, cpu.bmi, AM_RELATIVE)

	// BNE
	cpu.instructions[0xd0] = NewInstruction(OPC_BNE, 2, 2, cpu.bne, AM_RELATIVE)

	// BPL
	cpu.instructions[0x10] = NewInstruction(OPC_BPL, 2, 2, cpu.bpl, AM_RELATIVE)

	// BRK
	cpu.instructions[0x00] = NewInstruction(OPC_BRK, 7, 1, cpu.brk, AM_IMPLIED)

	// BVC
	cpu.instructions[0x50] = NewInstruction(OPC_BVC, 2, 2, cpu.bvc, AM_RELATIVE)

	// BVS
	cpu.instructions[0x70] = NewInstruction(OPC_BVS, 2, 2, cpu.bvs, AM_RELATIVE)

	// CLC
	cpu.instructions[0x18] = NewInstruction(OPC_CLC, 2, 1, cpu.clc, AM_IMPLIED)

	// CLD
	cpu.instructions[0xd8] = NewInstruction(OPC_CLD, 2, 1, cpu.cld, AM_IMPLIED)

	// CLI
	cpu.instructions[0x58] = NewInstruction(OPC_CLI, 2, 1, cpu.cli, AM_IMPLIED)

	// CLV
	cpu.instructions[0xb8] = NewInstruction(OPC_CLV, 2, 1, cpu.clv, AM_IMPLIED)

	// CMP
	cpu.instructions[0xc9] = NewInstruction(OPC_CMP, 2, 2, cpu.cmp, AM_IMMEDIATE)
	cpu.instructions[0xc5] = NewInstruction(OPC_CMP, 3, 2, cpu.cmp, AM_ZEROPAGE)
	cpu.instructions[0xd5] = NewInstruction(OPC_CMP, 4, 2, cpu.cmp, AM_ZEROPAGE_X)
	cpu.instructions[0xcd] = NewInstruction(OPC_CMP, 4, 3, cpu.cmp, AM_ABSOLUTE)
	cpu.instructions[0xdd] = NewInstruction(OPC_CMP, 4, 3, cpu.cmp, AM_INDEXED_X)
	cpu.instructions[0xd9] = NewInstruction(OPC_CMP, 4, 3, cpu.cmp, AM_INDEXED_Y)
	cpu.instructions[0xc1] = NewInstruction(OPC_CMP, 6, 2, cpu.cmp, AM_PRE_INDEXED)
	cpu.instructions[0xd1] = NewInstruction(OPC_CMP, 5, 2, cpu.cmp, AM_POST_INDEXED)

	// CPX
	cpu.instructions[0xe0] = NewInstruction(OPC_CPX, 2, 2, cpu.cpx, AM_IMMEDIATE)
	cpu.instructions[0xe4] = NewInstruction(OPC_CPX, 3, 2, cpu.cpx, AM_ZEROPAGE)
	cpu.instructions[0xec] = NewInstruction(OPC_CPX, 4, 3, cpu.cpx, AM_ABSOLUTE)

	// CPY
	cpu.instructions[0xc0] = NewInstruction(OPC_CPY, 2, 2, cpu.cpy, AM_IMMEDIATE)
	cpu.instructions[0xc4] = NewInstruction(OPC_CPY, 3, 2, cpu.cpy, AM_ZEROPAGE)
	cpu.instructions[0xcc] = NewInstruction(OPC_CPY, 4, 3, cpu.cpy, AM_ABSOLUTE)

	// DEC
	cpu.instructions[0xc6] = NewInstruction(OPC_DEC, 5, 2, cpu.dec, AM_ZEROPAGE)
	cpu.instructions[0xd6] = NewInstruction(OPC_DEC, 6, 2, cpu.dec, AM_ZEROPAGE_X)
	cpu.instructions[0xce] = NewInstruction(OPC_DEC, 6, 3, cpu.dec, AM_ABSOLUTE)

	// DEX
	cpu.instructions[0xca] = NewInstruction(OPC_DEX, 2, 1, cpu.dex, AM_IMPLIED)

	// DEY
	cpu.instructions[0x88] = NewInstruction(OPC_DEY, 2, 1, cpu.dey, AM_IMPLIED)

	// EOR
	cpu.instructions[0x49] = NewInstruction(OPC_EOR, 2, 2, cpu.eor, AM_IMMEDIATE)
	cpu.instructions[0x45] = NewInstruction(OPC_EOR, 3, 2, cpu.eor, AM_ZEROPAGE)
	cpu.instructions[0x55] = NewInstruction(OPC_EOR, 4, 2, cpu.eor, AM_ZEROPAGE_X)
	cpu.instructions[0x4d] = NewInstruction(OPC_EOR, 4, 3, cpu.eor, AM_ABSOLUTE)
	cpu.instructions[0x5d] = NewInstruction(OPC_EOR, 4, 3, cpu.eor, AM_INDEXED_X)
	cpu.instructions[0x59] = NewInstruction(OPC_EOR, 4, 3, cpu.eor, AM_INDEXED_Y)
	cpu.instructions[0x41] = NewInstruction(OPC_EOR, 6, 2, cpu.eor, AM_PRE_INDEXED)
	cpu.instructions[0x51] = NewInstruction(OPC_EOR, 5, 2, cpu.eor, AM_POST_INDEXED)

	// INC
	cpu.instructions[0xe6] = NewInstruction(OPC_INC, 5, 2, cpu.inc, AM_ZEROPAGE)
	cpu.instructions[0xf6] = NewInstruction(OPC_INC, 6, 2, cpu.inc, AM_ZEROPAGE_X)
	cpu.instructions[0xee] = NewInstruction(OPC_INC, 6, 3, cpu.inc, AM_ABSOLUTE)
	cpu.instructions[0xfe] = NewInstruction(OPC_INC, 7, 3, cpu.inc, AM_INDEXED_X)

	// INX
	cpu.instructions[0xe8] = NewInstruction(OPC_INX, 2, 1, cpu.inx, AM_IMPLIED)

	// INY
	cpu.instructions[0xc8] = NewInstruction(OPC_INY, 2, 1, cpu.iny, AM_IMPLIED)

	// JMP
	cpu.instructions[0x4c] = NewInstruction(OPC_JMP, 3, 3, cpu.jmp, AM_ABSOLUTE)
	cpu.instructions[0x6c] = NewInstruction(OPC_JMP, 5, 3, cpu.jmp, AM_INDIRECT)

	// JSR
	cpu.instructions[0x20] = NewInstruction(OPC_JSR, 6, 3, cpu.jsr, AM_ABSOLUTE)

	// LDA
	cpu.instructions[0xa9] = NewInstruction(OPC_LDA, 2, 2, cpu.lda, AM_IMMEDIATE)
	cpu.instructions[0xa5] = NewInstruction(OPC_LDA, 3, 2, cpu.lda, AM_ZEROPAGE)
	cpu.instructions[0xb5] = NewInstruction(OPC_LDA, 4, 2, cpu.lda, AM_ZEROPAGE_X)
	cpu.instructions[0xad] = NewInstruction(OPC_LDA, 4, 3, cpu.lda, AM_ABSOLUTE)
	cpu.instructions[0xbd] = NewInstruction(OPC_LDA, 4, 3, cpu.lda, AM_INDEXED_X)
	cpu.instructions[0xb9] = NewInstruction(OPC_LDA, 4, 3, cpu.lda, AM_INDEXED_Y)
	cpu.instructions[0xa1] = NewInstruction(OPC_LDA, 6, 2, cpu.lda, AM_PRE_INDEXED)
	cpu.instructions[0xb1] = NewInstruction(OPC_LDA, 5, 2, cpu.lda, AM_POST_INDEXED)

	// LDX
	cpu.instructions[0xa2] = NewInstruction(OPC_LDX, 2, 2, cpu.ldx, AM_IMMEDIATE)
	cpu.instructions[0xa6] = NewInstruction(OPC_LDX, 3, 2, cpu.ldx, AM_ZEROPAGE)
	cpu.instructions[0xb6] = NewInstruction(OPC_LDX, 4, 2, cpu.ldx, AM_ZEROPAGE_Y)
	cpu.instructions[0xae] = NewInstruction(OPC_LDX, 4, 3, cpu.ldx, AM_ABSOLUTE)
	cpu.instructions[0xbe] = NewInstruction(OPC_LDX, 4, 3, cpu.ldx, AM_INDEXED_Y)

	// LDY
	cpu.instructions[0xa0] = NewInstruction(OPC_LDY, 2, 2, cpu.ldy, AM_IMMEDIATE)
	cpu.instructions[0xa4] = NewInstruction(OPC_LDY, 3, 2, cpu.ldy, AM_ZEROPAGE)
	cpu.instructions[0xb4] = NewInstruction(OPC_LDY, 4, 2, cpu.ldy, AM_ZEROPAGE_Y)
	cpu.instructions[0xac] = NewInstruction(OPC_LDY, 4, 3, cpu.ldy, AM_ABSOLUTE)
	cpu.instructions[0xbc] = NewInstruction(OPC_LDY, 4, 3, cpu.ldy, AM_INDEXED_Y)

	// LSR
	cpu.instructions[0x4a] = NewInstruction(OPC_LSR, 2, 1, cpu.lsr, AM_IMPLIED)
	cpu.instructions[0x46] = NewInstruction(OPC_LSR, 5, 2, cpu.lsr, AM_ZEROPAGE)
	cpu.instructions[0x56] = NewInstruction(OPC_LSR, 6, 2, cpu.lsr, AM_ZEROPAGE_X)
	cpu.instructions[0x4e] = NewInstruction(OPC_LSR, 6, 3, cpu.lsr, AM_ABSOLUTE)
	cpu.instructions[0x5e] = NewInstruction(OPC_LSR, 7, 3, cpu.lsr, AM_INDEXED_X)

	// NOP
	cpu.instructions[0xea] = NewInstruction(OPC_NOP, 2, 1, cpu.nop, AM_IMPLIED)

	// ORA
	cpu.instructions[0x09] = NewInstruction(OPC_ORA, 2, 2, cpu.ora, AM_IMMEDIATE)
	cpu.instructions[0x05] = NewInstruction(OPC_ORA, 3, 2, cpu.ora, AM_ZEROPAGE)
	cpu.instructions[0x15] = NewInstruction(OPC_ORA, 4, 2, cpu.ora, AM_ZEROPAGE_X)
	cpu.instructions[0x0d] = NewInstruction(OPC_ORA, 4, 3, cpu.ora, AM_ABSOLUTE)
	cpu.instructions[0x1d] = NewInstruction(OPC_ORA, 4, 3, cpu.ora, AM_INDEXED_X)
	cpu.instructions[0x19] = NewInstruction(OPC_ORA, 4, 3, cpu.ora, AM_INDEXED_Y)
	cpu.instructions[0x01] = NewInstruction(OPC_ORA, 6, 2, cpu.ora, AM_PRE_INDEXED)
	cpu.instructions[0x11] = NewInstruction(OPC_ORA, 5, 2, cpu.ora, AM_POST_INDEXED)

	// PHA
	cpu.instructions[0x48] = NewInstruction(OPC_PHA, 3, 1, cpu.pha, AM_IMPLIED)

	// STA
	cpu.instructions[0x85] = NewInstruction(OPC_STA, 3, 2, cpu.sta, AM_ZEROPAGE)
	cpu.instructions[0x95] = NewInstruction(OPC_STA, 4, 2, cpu.sta, AM_ZEROPAGE_X)
	cpu.instructions[0x8d] = NewInstruction(OPC_STA, 4, 3, cpu.sta, AM_ABSOLUTE)
	cpu.instructions[0x9d] = NewInstruction(OPC_STA, 5, 3, cpu.sta, AM_INDEXED_X)
	cpu.instructions[0x99] = NewInstruction(OPC_STA, 5, 3, cpu.sta, AM_INDEXED_Y)
	cpu.instructions[0x81] = NewInstruction(OPC_STA, 6, 2, cpu.sta, AM_PRE_INDEXED)
	cpu.instructions[0x91] = NewInstruction(OPC_STA, 6, 2, cpu.sta, AM_POST_INDEXED)

	return &cpu
}

func (cpu *MOS6502) Reset(memory *Memory) {
	// reset registers
	cpu.a = 0xaa
	cpu.x = 0x0
	cpu.y = 0x0
	// reset stack pointer
	cpu.sp = StackTop
	// reset flags  http://forum.6502.org/viewtopic.php?t=829
	//    7   6   5   4   3   2   1   0
	//    N   V       B   D   I   Z   C
	//    *   *   1   1   0   1   *   *
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

	// increment the pc by the number of bytes read for the operand
	cpu.pc += uint16(instruction.size - 1)
	// mark the cpu busy for the number of cycles the instruction takes
	cpu.wait = instruction.cycles - 1

	instruction.execute(operand)
}

// push a byte onto the stack if we go exceed the stack
// wrap around to the top of the stack
func (cpu *MOS6502) push(b uint8) {
	cpu.memory[StackBottom|cpu.sp] = b
	cpu.sp--
	if cpu.sp < StackBottom {
		cpu.sp = StackTop
	}
}

func fmt8(n string, b uint8) {
	fmt.Printf("%s\t%08b\t%02x\n", n, b, b)
}

func fmt16(n string, b uint16) {
	fmt.Printf("%s\t%08b\t%04x\n", n, b, b)
}
