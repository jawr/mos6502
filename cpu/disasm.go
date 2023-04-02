package cpu

import (
	"fmt"
)

type DisassembledInstruction struct {
	Address     uint16
	Opcode      OPCode
	Operand     uint16
	Mode        AddressMode
	Disassembly string
}

func (cpu *MOS6502) disassembleInstruction(address uint16) *DisassembledInstruction {
	opcode := cpu.memory.Read(address)
	instruction := cpu.instructions[opcode]

	if instruction == nil {
		return nil
	}

	var operand uint16
	var disassembly string

	if instruction.size > 1 {
		operand = cpu.memory.ReadWord(address + 1)
	}

	disassembly = fmt.Sprintf("%s ", instruction.opc)

	switch instruction.mode {
	case AM_IMPLIED:
		// No additional operands
	case AM_IMMEDIATE:
		disassembly += fmt.Sprintf("#$%02X", operand&0xFF)
	case AM_ABSOLUTE:
		disassembly += fmt.Sprintf("$%04X", operand)
	case AM_ZEROPAGE:
		disassembly += fmt.Sprintf("$%02X", operand&0xFF)
	case AM_INDEXED_X:
		disassembly += fmt.Sprintf("$%04X,X", operand)
	case AM_INDEXED_Y:
		disassembly += fmt.Sprintf("$%04X,Y", operand)
	case AM_ZEROPAGE_X:
		disassembly += fmt.Sprintf("$%02X,X", operand&0xFF)
	case AM_ZEROPAGE_Y:
		disassembly += fmt.Sprintf("$%02X,Y", operand&0xFF)
	case AM_INDIRECT:
		disassembly += fmt.Sprintf("($%04X)", operand)
	case AM_PRE_INDEXED:
		disassembly += fmt.Sprintf("($%02X,X)", operand&0xFF)
	case AM_POST_INDEXED:
		disassembly += fmt.Sprintf("($%02X),Y", operand&0xFF)
	case AM_RELATIVE:
		disassembly += fmt.Sprintf("$%04X", address+2+uint16(int8(operand&0xFF)))
	}

	return &DisassembledInstruction{
		Address:     address,
		Opcode:      instruction.opc,
		Operand:     operand,
		Mode:        instruction.mode,
		Disassembly: disassembly,
	}
}
