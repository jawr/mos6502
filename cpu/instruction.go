package cpu

import "fmt"

// the address mode of the instruction determines how the
// operand interacts with the registers and memory
type AddressMode uint8

// notes borrowed from https://www.masswerk.at/6502/6502_instruction_set.htm
const (
	// operand implied
	AM_IMPLIED AddressMode = iota
	// operand is byte BB (in OPC #$BB)
	AM_IMMEDIATE
	// operand is address $HHLL (in OPC $LLHH)
	AM_ABSOLUTE
	// operand is zeropage address (hi-byte is zero, address = $00LL) (in OPC $LL)
	AM_ZEROPAGE
	// operand is address; effective address is address incremented by X with carry
	AM_INDEXED_X
	// operand is address; effective address is address incremented by Y with carry
	AM_INDEXED_Y
	// operand is zeropage address; effective address is address incremented by X without carry (in OPC $LL,X)
	AM_ZEROPAGE_X
	// operand is zeropage address; effective address is address incremented by Y without carry (in OPC $LL,Y)
	AM_ZEROPAGE_Y
	// operand is address; effective address is contents of word at address: C.w($HHLL) (in OPC ($LLHH))
	AM_INDIRECT
	// operand is zeropage address; effective address is word in (LL + X, LL + X + 1), inc. without carry:
	//	C.w($00LL + X)
	// also known as X-indexed, indirect
	AM_PRE_INDEXED
	// operand is zeropage address; effective address is word in (LL, LL + 1) incremented by Y with carry:
	//	C.w($00LL) + Y
	// also known as indirect, Y-indexed
	AM_POST_INDEXED
	// branch target is PC + signed offset BB (in OPC $BB)
	AM_RELATIVE
)

// the instruction by name
type OPCode string

const (
	OPC_ADC = "ADC"
	OPC_AND = "AND"
	OPC_ASL = "ASL"
	OPC_BCC = "BCC"
	OPC_BCS = "BCS"
	OPC_BEQ = "BEQ"
	OPC_BIT = "BIT"
	OPC_BMI = "BMI"
	OPC_BNE = "BNE"
	OPC_BPL = "BPL"
	OPC_BRK = "BRK"
	OPC_BVC = "BVC"
	OPC_BVS = "BVS"
	OPC_CLC = "CLC"
	OPC_CLD = "CLD"
	OPC_CLI = "CLI"
	OPC_CLV = "CLV"
	OPC_CMP = "CMP"
	OPC_CPX = "CPX"
	OPC_CPY = "CPY"
	OPC_DEC = "DEC"
	OPC_DEX = "DEX"
	OPC_DEY = "DEY"
	OPC_EOR = "EOR"
	OPC_INC = "INC"
	OPC_INX = "INX"
	OPC_INY = "INY"
	OPC_JMP = "JMP"
	OPC_JSR = "JSR"
	OPC_LDA = "LDA"
	OPC_LDX = "LDX"
	OPC_LDY = "LDY"
	OPC_LSR = "LSR"
	OPC_NOP = "NOP"
	OPC_ORA = "ORA"
	OPC_PHA = "PHA"
	OPC_PHP = "PHP"
	OPC_PLA = "PLA"
	OPC_PLP = "PLP"
	OPC_ROL = "ROL"
	OPC_ROR = "ROR"
	OPC_RTI = "RTI"
	OPC_RTS = "RTS"
	OPC_SBC = "SBC"
	OPC_SEC = "SEC"
	OPC_SED = "SED"
	OPC_SEI = "SEI"
	OPC_STA = "STA"
	OPC_STX = "STX"
	OPC_STY = "STY"
	OPC_TAX = "TAX"
	OPC_TAY = "TAY"
	OPC_TSX = "TSX"
	OPC_TXA = "TXA"
	OPC_TXS = "TXS"
	OPC_TYA = "TYA"
)

// the function that will be executed for this instruction
type executor func(*instruction, uint16)

type instruction struct {
	opc    OPCode
	cycles uint8
	size   uint8 // number of bytes to load
	fn     executor
	mode   AddressMode
}

func NewInstruction(opc OPCode, cycles, size uint8, fn executor, mode AddressMode) *instruction {
	if cycles == 0 {
		panic(fmt.Sprintf("instruction %s has 0 cycles", opc))
	}
	if size == 0 {
		panic(fmt.Sprintf("instruction %s has 0 size", opc))
	}

	return &instruction{
		opc:    opc,
		cycles: cycles,
		size:   size,
		fn:     fn,
		mode:   mode,
	}
}

func (i *instruction) execute(operand uint16) {
	i.fn(i, operand)
}

func (i *instruction) load(cpu *MOS6502) uint16 {
	switch i.mode {
	case AM_IMPLIED:
		// single byte instructrions
		return 0

	case AM_IMMEDIATE:
		// literal operand loaded into memory
		// always an 8 bit value
		return cpu.pc + 1

	case AM_ABSOLUTE:
		// full 16 bit address in LLHH format
		lo := cpu.memory.Read(cpu.pc + 1)
		hi := cpu.memory.Read(cpu.pc + 2)

		return (uint16(hi) << 8) + uint16(lo)

	case AM_ZEROPAGE:
		// 1 byte address in the zeropage (high byte is 0x00)
		return uint16(cpu.memory.Read(cpu.pc + 1))

	case AM_ZEROPAGE_X:
		// first byte comes from pc
		address := cpu.memory.Read(cpu.pc + 1)
		// add contents of x register
		address += cpu.x
		// address is 8 bits so will wrap around in the zeropage
		return uint16(address)

	case AM_ZEROPAGE_Y:
		// first byte comes from pc
		address := cpu.memory.Read(cpu.pc + 1)
		// add contents of y register
		address += cpu.y
		// address is 8 bits so will wrap around in the zeropage
		return uint16(address)

	case AM_INDEXED_X:
		// read 16 bit address in LLHH format
		lo := cpu.memory.Read(cpu.pc + 1)
		hi := cpu.memory.Read(cpu.pc + 2)

		address := (uint16(hi) << 8) + uint16(lo)
		offsetAddress := address + uint16(cpu.x)

		// track page boundary crossing
		if crossedPageBoundary(address, offsetAddress) {
			cpu.additionalCycles++
		}

		return offsetAddress

	case AM_INDEXED_Y:
		// read 16 bit address in LLHH format
		lo := cpu.memory.Read(cpu.pc + 1)
		hi := cpu.memory.Read(cpu.pc + 2)

		address := (uint16(hi) << 8) + uint16(lo)
		offsetAddress := address + uint16(cpu.y)

		// track page boundary crossing
		if crossedPageBoundary(address, offsetAddress) {
			cpu.additionalCycles++
		}

		return offsetAddress

	case AM_PRE_INDEXED:
		// first byte comes from pc
		address := cpu.memory.Read(cpu.pc + 1)

		// add contents of x register
		address += cpu.x

		// get the lookup from this address
		lookup := cpu.memory.ReadWord(uint16(address))

		// resolve the lookup
		return lookup

	case AM_POST_INDEXED:
		// first byte comes from pc
		address := cpu.memory.Read(cpu.pc + 1)

		// get the lookup from zeropage
		lookup := cpu.memory.ReadWord(uint16(address))

		// add contents of y register
		offsetAddress := lookup + uint16(cpu.y)

		// track page boundary crossing
		if crossedPageBoundary(lookup, offsetAddress) {
			cpu.additionalCycles++
		}

		// resolve the lookup
		return offsetAddress

	case AM_INDIRECT:
		// get the indirect address
		lo := cpu.memory.Read(cpu.pc + 1)
		hi := cpu.memory.Read(cpu.pc + 2)

		address := (uint16(hi) << 8) + uint16(lo)

		// read the address from the indirect address
		return cpu.memory.ReadWord(address)

	case AM_RELATIVE:
		address := uint16(cpu.memory.Read(cpu.pc + 1))
		return address

	default:
		panic("invalid load address mode")
	}
}

// Helper function to check if a page boundary was crossed
func crossedPageBoundary(oldAddress, newAddress uint16) bool {
	return oldAddress&0xFF00 != newAddress&0xFF00
}

func (cpu *MOS6502) setupInstructions() {
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

	// PHP
	cpu.instructions[0x08] = NewInstruction(OPC_PHP, 3, 1, cpu.php, AM_IMPLIED)

	// PLA
	cpu.instructions[0x68] = NewInstruction(OPC_PLA, 4, 1, cpu.pla, AM_IMPLIED)

	// PLP
	cpu.instructions[0x28] = NewInstruction(OPC_PLP, 4, 1, cpu.plp, AM_IMPLIED)

	// ROL
	cpu.instructions[0x2a] = NewInstruction(OPC_ROL, 2, 1, cpu.rol, AM_IMPLIED)
	cpu.instructions[0x26] = NewInstruction(OPC_ROL, 5, 2, cpu.rol, AM_ZEROPAGE)
	cpu.instructions[0x36] = NewInstruction(OPC_ROL, 6, 2, cpu.rol, AM_ZEROPAGE_X)
	cpu.instructions[0x2e] = NewInstruction(OPC_ROL, 6, 3, cpu.rol, AM_ABSOLUTE)
	cpu.instructions[0x3e] = NewInstruction(OPC_ROL, 7, 3, cpu.rol, AM_INDEXED_X)

	// ROR
	cpu.instructions[0x6a] = NewInstruction(OPC_ROR, 2, 1, cpu.ror, AM_IMPLIED)
	cpu.instructions[0x66] = NewInstruction(OPC_ROR, 5, 2, cpu.ror, AM_ZEROPAGE)
	cpu.instructions[0x76] = NewInstruction(OPC_ROR, 6, 2, cpu.ror, AM_ZEROPAGE_X)
	cpu.instructions[0x6e] = NewInstruction(OPC_ROR, 6, 3, cpu.ror, AM_ABSOLUTE)
	cpu.instructions[0x7e] = NewInstruction(OPC_ROR, 7, 3, cpu.ror, AM_INDEXED_X)

	// RTI
	cpu.instructions[0x40] = NewInstruction(OPC_RTI, 6, 1, cpu.rti, AM_IMPLIED)

	// RTS
	cpu.instructions[0x60] = NewInstruction(OPC_RTS, 6, 1, cpu.rts, AM_IMPLIED)

	// SBC
	cpu.instructions[0xe9] = NewInstruction(OPC_SBC, 2, 2, cpu.sbc, AM_IMMEDIATE)
	cpu.instructions[0xe5] = NewInstruction(OPC_SBC, 3, 2, cpu.sbc, AM_ZEROPAGE)
	cpu.instructions[0xf5] = NewInstruction(OPC_SBC, 4, 2, cpu.sbc, AM_ZEROPAGE_X)
	cpu.instructions[0xed] = NewInstruction(OPC_SBC, 4, 3, cpu.sbc, AM_ABSOLUTE)
	cpu.instructions[0xfd] = NewInstruction(OPC_SBC, 4, 3, cpu.sbc, AM_INDEXED_X)
	cpu.instructions[0xf9] = NewInstruction(OPC_SBC, 4, 3, cpu.sbc, AM_INDEXED_Y)
	cpu.instructions[0xe1] = NewInstruction(OPC_SBC, 6, 2, cpu.sbc, AM_PRE_INDEXED)
	cpu.instructions[0xf1] = NewInstruction(OPC_SBC, 5, 2, cpu.sbc, AM_POST_INDEXED)

	// SEC
	cpu.instructions[0x38] = NewInstruction(OPC_SEC, 2, 1, cpu.sec, AM_IMPLIED)

	// SED
	cpu.instructions[0xf8] = NewInstruction(OPC_SED, 2, 1, cpu.sed, AM_IMPLIED)

	// SEI
	cpu.instructions[0x78] = NewInstruction(OPC_SEI, 2, 1, cpu.sei, AM_IMPLIED)

	// STA
	cpu.instructions[0x85] = NewInstruction(OPC_STA, 3, 2, cpu.sta, AM_ZEROPAGE)
	cpu.instructions[0x95] = NewInstruction(OPC_STA, 4, 2, cpu.sta, AM_ZEROPAGE_X)
	cpu.instructions[0x8d] = NewInstruction(OPC_STA, 4, 3, cpu.sta, AM_ABSOLUTE)
	cpu.instructions[0x9d] = NewInstruction(OPC_STA, 5, 3, cpu.sta, AM_INDEXED_X)
	cpu.instructions[0x99] = NewInstruction(OPC_STA, 5, 3, cpu.sta, AM_INDEXED_Y)
	cpu.instructions[0x81] = NewInstruction(OPC_STA, 6, 2, cpu.sta, AM_PRE_INDEXED)
	cpu.instructions[0x91] = NewInstruction(OPC_STA, 6, 2, cpu.sta, AM_POST_INDEXED)

	// STX
	cpu.instructions[0x86] = NewInstruction(OPC_STX, 3, 2, cpu.stx, AM_ZEROPAGE)
	cpu.instructions[0x96] = NewInstruction(OPC_STX, 4, 2, cpu.stx, AM_ZEROPAGE_Y)
	cpu.instructions[0x8e] = NewInstruction(OPC_STX, 4, 3, cpu.stx, AM_ABSOLUTE)

	// STY
	cpu.instructions[0x84] = NewInstruction(OPC_STY, 3, 2, cpu.sty, AM_ZEROPAGE)
	cpu.instructions[0x94] = NewInstruction(OPC_STY, 4, 2, cpu.sty, AM_ZEROPAGE_X)
	cpu.instructions[0x8c] = NewInstruction(OPC_STY, 4, 3, cpu.sty, AM_ABSOLUTE)

	// TAX
	cpu.instructions[0xaa] = NewInstruction(OPC_TAX, 2, 1, cpu.tax, AM_IMPLIED)

	// TAY
	cpu.instructions[0xa8] = NewInstruction(OPC_TAY, 2, 1, cpu.tay, AM_IMPLIED)

	// TSX
	cpu.instructions[0xba] = NewInstruction(OPC_TSX, 2, 1, cpu.tsx, AM_IMPLIED)

	// TXA
	cpu.instructions[0x8a] = NewInstruction(OPC_TXA, 2, 1, cpu.txa, AM_IMPLIED)

	// TXS
	cpu.instructions[0x9a] = NewInstruction(OPC_TXS, 2, 1, cpu.txs, AM_IMPLIED)

	// TYA
	cpu.instructions[0x98] = NewInstruction(OPC_TYA, 2, 1, cpu.tya, AM_IMPLIED)
}
