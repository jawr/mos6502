package cpu

func (cpu *MOS6502) adc(ins *instruction, address uint16) {
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

	cpu.a = uint8(sum & 0xff)
	cpu.testAndSetCarry(sum)
	cpu.testAndSetNegative(cpu.a)
	cpu.testAndSetZero(cpu.a)
	cpu.testAndSetOverflow(a, m, cpu.a)
}

func (cpu *MOS6502) and(ins *instruction, address uint16) {
	// And Memory with Accumulator
	b := cpu.memory.Read(address)
	cpu.a = cpu.a & b
	cpu.testAndSetNegative(cpu.a)
	cpu.testAndSetZero(cpu.a)
}

func (cpu *MOS6502) asl(ins *instruction, address uint16) {
	// Shift Left One Bit (Memory or Accumulator)
	accumulator := ins.mode == AM_IMPLIED

	// if we are immediate get from the accumulator
	value := cpu.a
	if !accumulator {
		value = cpu.memory.Read(address)
	}

	// shift right
	shifted := uint16(value) << 1

	if accumulator {
		cpu.a = uint8(shifted)
	} else {
		cpu.memory[address] = uint8(shifted)
	}

	cpu.testAndSetNegative(uint8(shifted))
	cpu.testAndSetZero(uint8(shifted))
	cpu.testAndSetCarry(shifted)
}

func (cpu *MOS6502) bcc(ins *instruction, address uint16) {
	// Branch on Carry Clear
	if cpu.p.isSet(P_Carry) {
		return
	}
	cpu.pc = address
}

func (cpu *MOS6502) bcs(ins *instruction, address uint16) {
	// Branch on Carry Set
	if !cpu.p.isSet(P_Carry) {
		return
	}
	cpu.pc = address
}

func (cpu *MOS6502) beq(ins *instruction, address uint16) {
	// Branch on Result Zero
	if !cpu.p.isSet(P_Zero) {
		return
	}
	cpu.pc = address
}

func (cpu *MOS6502) bit(ins *instruction, address uint16) {
	// Test Bits in Memory with Accumulator
	// bits 7 and 6 of operand are transfered to bit 7 and 6 of SR (N,V);
	// the zero-flag is set to the result of operand AND accumulator.

	value := cpu.memory.Read(address)

	cpu.testAndSetZero(cpu.a & value)

	if value&(1<<7) != 0 {
		cpu.p.set(P_Negative)
	}

	if value&(1<<6) != 0 {
		cpu.p.set(P_Overflow)
	}
}

func (cpu *MOS6502) bmi(ins *instruction, address uint16) {
	// Branch on Result Minus
	if !cpu.p.isSet(P_Negative) {
		return
	}
	cpu.pc = address
}

func (cpu *MOS6502) bne(ins *instruction, address uint16) {
	// Branch on Result not Zero
	if cpu.p.isSet(P_Zero) {
		return
	}
	cpu.pc = address
}

func (cpu *MOS6502) bpl(ins *instruction, address uint16) {
	// Branch on Result Plus
	if cpu.p.isSet(P_Negative) {
		return
	}
	cpu.pc = address
}

func (cpu *MOS6502) brk(ins *instruction, address uint16) {
	// Force Break
	// push return address to stack
	cpu.push(uint8(cpu.pc >> 8))
	cpu.push(uint8(cpu.pc & 0xff))

	// push status register to stack with break flag set
	cpu.p.set(P_Break)
	cpu.push(uint8(cpu.p))

	// set intterupt disable
	cpu.p.set(P_InterruptDisable)

	// push interrupt vector to pc
	hi := uint16(cpu.memory.Read(0xfffa)) << 8
	lo := uint16(cpu.memory.Read(0xfffb))

	cpu.pc = uint16(lo | hi)
}

func (cpu *MOS6502) clc(ins *instruction, address uint16) {
	// Clear Carry Flag
	cpu.p.clear(P_Carry)
}

func (cpu *MOS6502) cld(ins *instruction, address uint16) {
	// Clear Decimal Mode
	cpu.p.clear(P_Decimal)
}

func (cpu *MOS6502) cli(ins *instruction, address uint16) {
	// Clear Interrupt Disable Bit
	cpu.p.clear(P_InterruptDisable)
}

func (cpu *MOS6502) clv(ins *instruction, address uint16) {
	// Clear Overflow Flag
	cpu.p.clear(P_Overflow)
}

func (cpu *MOS6502) inx(ins *instruction, address uint16) {
	// Increment Index X by One
	cpu.x++
	cpu.testAndSetNegative(cpu.x)
	cpu.testAndSetZero(cpu.x)
}

func (cpu *MOS6502) iny(ins *instruction, address uint16) {
	// Increment Index Y by One
	cpu.y++
	cpu.testAndSetNegative(cpu.y)
	cpu.testAndSetZero(cpu.y)
}

func (cpu *MOS6502) inc(ins *instruction, address uint16) {
	// Increment Memory by One
	cpu.memory[address]++
	value := cpu.memory.Read(address)
	cpu.testAndSetNegative(value)
	cpu.testAndSetZero(value)
}

func (cpu *MOS6502) jmp(ins *instruction, address uint16) {
	// Jump to New Location
	cpu.pc = address
}

func (cpu *MOS6502) jsr(ins *instruction, address uint16) {
	// Jump to New Location Saving Return Address
	lo := uint8(address)
	hi := uint8(address >> 8)

	// push the lo then the hi bytes on to the stack
	cpu.push(lo)
	cpu.push(hi)
}

func (cpu *MOS6502) nop(ins *instruction, address uint16) {
	// No Operation
}

func (cpu *MOS6502) lda(ins *instruction, address uint16) {
	// Load Accumulator with Memory
	value := cpu.memory.ReadWord(address)
	cpu.a = uint8(value)
	cpu.testAndSetNegative(cpu.a)
	cpu.testAndSetZero(cpu.a)
}

func (cpu *MOS6502) ldx(ins *instruction, address uint16) {
	// Load Index X with Memory
	value := cpu.memory.ReadWord(address)
	cpu.x = uint8(value)
	cpu.testAndSetNegative(cpu.x)
	cpu.testAndSetZero(cpu.x)
}

func (cpu *MOS6502) ldy(ins *instruction, address uint16) {
	// Load Index X with Memory
	value := cpu.memory.ReadWord(address)
	cpu.y = uint8(value)
	cpu.testAndSetNegative(cpu.y)
	cpu.testAndSetZero(cpu.y)
}

func (cpu *MOS6502) lsr(ins *instruction, address uint16) {
	// Shift One Bit Right (Memory or Accumulator)

	accumulator := ins.mode == AM_IMPLIED

	// if we are immediate get from the accumulator
	value := cpu.a
	if !accumulator {
		value = cpu.memory.Read(address)
	}

	// shift right
	shifted := uint16(value) >> 1

	if accumulator {
		cpu.a = uint8(shifted)
	} else {
		cpu.memory[address] = uint8(shifted)
	}

	cpu.testAndSetZero(uint8(shifted))
	cpu.testAndSetCarry(shifted)
}

func (cpu *MOS6502) ora(ins *instruction, address uint16) {
	// Or Memory with Accumulator
	value := cpu.memory.Read(address)
	cpu.a = cpu.a | value

	cpu.testAndSetNegative(cpu.a)
	cpu.testAndSetZero(cpu.a)
}

func (cpu *MOS6502) sta(ins *instruction, address uint16) {
	// Store Accumulator in Memory
	cpu.memory[address] = cpu.a
}
