package cpu

import "fmt"

func (cpu *MOS6502) adc(address uint16) {
	// Add Memory to Accumulator with Carry
	// A + M + C -> A, C
	var c uint8 = 0
	if cpu.p.isSet(P_Carry) {
		c = 1
	}

	a := cpu.a
	m := cpu.memory.Read(address)

	// sum in uint16 to catch overflow
	sum := uint16(a) + uint16(m) + uint16(c)

	fmt.Printf("adc %04x v=%02x a=%02x c=%02x sum=%04x\n", address, m, a, c, sum)

	cpu.a = uint8(sum & 0xff)
	cpu.testAndSetCarry(sum)
	cpu.testAndSetNegative(cpu.a)
	cpu.testAndSetZero(cpu.a)
	cpu.testAndSetOverflow(a, m, cpu.a)
}

func (cpu *MOS6502) and(address uint16) {
	// And Memory with Accumulator
	b := cpu.memory.Read(address)
	fmt.Printf("and %04x v=%02x & a=%02x\n", address, b, cpu.a)
	cpu.a = cpu.a & b
	cpu.testAndSetNegative(cpu.a)
	cpu.testAndSetZero(cpu.a)
}

func (cpu *MOS6502) clc(address uint16) {
	// Clear Carry Flag
	cpu.p.clear(P_Carry)
}

func (cpu *MOS6502) cld(address uint16) {
	// Clear Decimal Mode
	cpu.p.clear(P_Decimal)
}

func (cpu *MOS6502) cli(address uint16) {
	// Clear Interrupt Disable Bit
	cpu.p.clear(P_InterruptDisable)
}

func (cpu *MOS6502) clv(address uint16) {
	// Clear Overflow Flag
	cpu.p.clear(P_Overflow)
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
