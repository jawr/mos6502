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
	OPC_NOP = "NOP"
	OPC_ORA = "ORA"
	OPC_PHA = "PHA"
	OPC_STA = "STA"
	OPC_LSR = "LSR"
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

func (i *instruction) parseOperand(cpu *MOS6502) uint16 {
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
		// address is 8 bits so will wrap around in the zeropage
		return uint16(address)

	case AM_ZEROPAGE_Y:
		// first byte comes from pc
		address := cpu.memory.Read(cpu.pc)
		// add contents of y register
		address += cpu.y
		// address is 8 bits so will wrap around in the zeropage
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

	case AM_RELATIVE:
		address := uint16(cpu.memory.Read(cpu.pc))
		address += cpu.pc
		return address

	default:
		panic("unsupported address mode")
	}
}
