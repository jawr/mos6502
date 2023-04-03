package cpu

import (
	"fmt"
	"log"
)

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
	StackOffset uint16 = 0x0100
	StackBottom uint8  = 0x00
	StackTop    uint8  = 0xff
)

// describe the kind of halt received
type HaltType uint8

const (
	Continue HaltType = iota
	HaltSuccess
	HaltTrap
	HaltUnknownInstruction
)

type MOS6502 struct {
	// main register
	a uint8
	// index registers
	x uint8
	y uint8

	sp uint8

	// program counter
	pc uint16

	p flags

	// operations take a predetermined amount of time
	wait uint8

	// instruction table
	instructions [0x100]*instruction

	// memory thats set on reset
	memory *Memory

	// halt the cpu
	halt HaltType

	// print out step debug information
	Debug bool
	// detect if we are in a trap loop
	TrapDetector bool
	trapDetector trapDetector

	// catpure the number of additional cycles
	additionalCycles uint8

	// total cycle count
	TotalCycles uint64

	// last test
	StopOnPC uint16
}

func NewMOS6502() *MOS6502 {
	cpu := MOS6502{}

	// setup the instruction table
	cpu.setupInstructions()

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

func (cpu *MOS6502) SetPC(pc uint16) {
	cpu.pc = pc
}

func (cpu *MOS6502) Halt() HaltType {
	return cpu.halt
}

func (cpu *MOS6502) Cycle() {
	if cpu.pc == uint16(cpu.StopOnPC) {
		cpu.halt = HaltSuccess
		return
	}

	// reset state
	cpu.additionalCycles = 0

	// pop the 8bit opcode and progress the pc
	opcode := cpu.memory.Read(cpu.pc)

	// read the instruction from the table halting if not found
	instruction := cpu.instructions[opcode]
	if instruction == nil {
		cpu.halt = HaltUnknownInstruction
		log.Printf("no instruction found for opcode %02x at %04x", opcode, opcode)
		return
	}

	// increment the pc by the number of bytes read for the operand
	address := instruction.load(cpu)

	if cpu.Debug {
		disasm := cpu.disassembleInstruction(cpu.pc)
		log.Printf(
			"%04x : %02x\t%-30s\t%s\tA:%02x X:%02x Y:%02x\tSP:%04x",
			cpu.pc,
			opcode,
			disasm.Disassembly,
			cpu.p.String(),
			cpu.a,
			cpu.x,
			cpu.y,
			cpu.sp,
		)
	}

	if cpu.TrapDetector {
		cpu.trapDetector.push(cpu.pc)
		if cpu.trapDetector.hastrap() {
			cpu.halt = HaltTrap
			log.Printf("trap detected at %04x", cpu.pc)
			return
		}
	}

	// increment the pc by the size of the instruction
	cpu.pc += uint16(instruction.size)

	// mark the cpu busy for the number of cycles the instruction takes (- this cycle)
	cpu.TotalCycles += uint64(instruction.cycles + cpu.additionalCycles)

	instruction.execute(address)
}

func stackAddress(sp uint8) uint16 {
	return (StackOffset | uint16(sp))
}

// push a byte onto the stack if we overflow wrap around to the top of the stack
func (cpu *MOS6502) push(b uint8) {
	cpu.memory[stackAddress(cpu.sp)] = b
	cpu.sp--
}

// pop a byte off the stack. if we overflow wrap around to the bottom of the stack
func (cpu *MOS6502) pop() uint8 {
	cpu.sp++
	b := cpu.memory[stackAddress(cpu.sp)]
	return b
}

func fmt8(n string, b uint8) string {
	return fmt.Sprintf("%s\t%08b\t%02x\n", n, b, b)
}

func fmt16(n string, b uint16) string {
	return fmt.Sprintf("%s\t%08b\t%04x\n", n, b, b)
}
