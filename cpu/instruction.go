package cpu

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
