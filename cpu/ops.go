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
	cpu.branch(address)
}

func (cpu *MOS6502) bcs(ins *instruction, address uint16) {
	// Branch on Carry Set
	if !cpu.p.isSet(P_Carry) {
		return
	}
	cpu.branch(address)
}

func (cpu *MOS6502) beq(ins *instruction, address uint16) {
	// Branch on Result Zero
	if !cpu.p.isSet(P_Zero) {
		return
	}
	cpu.branch(address)
}

func (cpu *MOS6502) bit(ins *instruction, address uint16) {
	// Test Bits in Memory with Accumulator
	// bits 7 and 6 of operand are transfered to bit 7 and 6 of SR (N,V);
	// the zero-flag is set to the result of operand AND accumulator.

	value := cpu.memory.Read(address)

	cpu.testAndSetZero(cpu.a & value)

	// check if 8th bit is set
	cpu.p.set(P_Negative, value&(1<<7) != 0)
	// check if 7th bit is set
	cpu.p.set(P_Overflow, value&(1<<6) != 0)
}

func (cpu *MOS6502) branch(offset uint16) {
	begin := cpu.pc

	if offset < 0x80 {
		cpu.pc += offset
	} else {
		cpu.pc -= 0x100 - offset
	}

	cpu.additionalCycles++
	if (begin & 0xff00) != (cpu.pc & 0xff00) {
		cpu.additionalCycles++
	}
}

func (cpu *MOS6502) bmi(ins *instruction, address uint16) {
	// Branch on Result Minus
	if !cpu.p.isSet(P_Negative) {
		return
	}
	cpu.branch(address)
}

func (cpu *MOS6502) bne(ins *instruction, address uint16) {
	// Branch on Result not Zero
	if cpu.p.isSet(P_Zero) {
		return
	}
	cpu.branch(address)
}

func (cpu *MOS6502) bpl(ins *instruction, address uint16) {
	// Branch on Result Plus
	if cpu.p.isSet(P_Negative) {
		return
	}
	cpu.branch(address)
}

func (cpu *MOS6502) brk(ins *instruction, address uint16) {
	// increment the pc so that BRK takes up the space of
	// a 2 byte instruction and can replace it
	cpu.pc++

	// Force Break
	// push return address to stack
	cpu.push(uint8(cpu.pc >> 8))
	cpu.push(uint8(cpu.pc & 0xff))

	// push status register to stack with break flag set
	cpu.p.set(P_Break, true)
	cpu.push(uint8(cpu.p))

	// set intterupt disable
	cpu.p.set(P_InterruptDisable, true)

	// push interrupt vector to pc
	hi := uint16(cpu.memory.Read(IRQVectorHigh)) << 8
	lo := uint16(cpu.memory.Read(IRQVectorLow))

	cpu.pc = uint16(lo | hi)
}

func (cpu *MOS6502) bvc(ins *instruction, address uint16) {
	// Branch on Overflow Clear
	if cpu.p.isSet(P_Overflow) {
		return
	}
	cpu.branch(address)
}

func (cpu *MOS6502) bvs(ins *instruction, address uint16) {
	// Branch on Overflow Set
	if !cpu.p.isSet(P_Overflow) {
		return
	}
	cpu.branch(address)
}

func (cpu *MOS6502) clc(ins *instruction, address uint16) {
	// Clear Carry Flag
	cpu.p.set(P_Carry, false)
}

func (cpu *MOS6502) cld(ins *instruction, address uint16) {
	// Clear Decimal Mode
	cpu.p.set(P_Decimal, false)
}

func (cpu *MOS6502) cli(ins *instruction, address uint16) {
	// Clear Interrupt Disable Bit
	cpu.p.set(P_InterruptDisable, false)
}

func (cpu *MOS6502) clv(ins *instruction, address uint16) {
	// Clear Overflow Flag
	cpu.p.set(P_Overflow, false)
}

func (cpu *MOS6502) cmp(ins *instruction, address uint16) {
	// Compare Memory with Accumulator
	b := cpu.memory.Read(address)

	// check if the memory is less than the accumulator
	sub := cpu.a - b

	cpu.p.set(P_Carry, cpu.a >= b)

	cpu.testAndSetNegative(sub)
	cpu.testAndSetZero(sub)
}

func (cpu *MOS6502) cpx(ins *instruction, address uint16) {
	// Compare Memory with Accumulator
	b := cpu.memory.Read(address)

	// check if the memory is less than the accumulator
	sub := cpu.x - b

	cpu.p.set(P_Carry, cpu.x >= b)

	cpu.testAndSetNegative(sub)
	cpu.testAndSetZero(sub)
}

func (cpu *MOS6502) cpy(ins *instruction, address uint16) {
	// Compare Memory with Accumulator
	b := cpu.memory.Read(address)

	// check if the memory is less than the accumulator
	sub := cpu.y - b

	cpu.p.set(P_Carry, cpu.y >= b)

	cpu.testAndSetNegative(sub)
	cpu.testAndSetZero(sub)
}

func (cpu *MOS6502) dec(ins *instruction, address uint16) {
	// Decrement Memory by One
	b := cpu.memory.Read(address)
	b = b - 1
	cpu.memory[address] = b

	cpu.testAndSetNegative(b)
	cpu.testAndSetZero(b)
}

func (cpu *MOS6502) dex(ins *instruction, address uint16) {
	// Decrement Index X by One
	// wrapping is handled by go uint
	cpu.x--
	cpu.testAndSetNegative(cpu.x)
	cpu.testAndSetZero(cpu.x)
}

func (cpu *MOS6502) dey(ins *instruction, address uint16) {
	// Decrement Index Y by One
	// wrapping is handled by go uint
	cpu.y--
	cpu.testAndSetNegative(cpu.y)
	cpu.testAndSetZero(cpu.y)
}

func (cpu *MOS6502) eor(ins *instruction, address uint16) {
	// Exclusive-OR Memory with Accumulator
	value := cpu.memory.Read(address)
	cpu.a = cpu.a ^ value
	cpu.testAndSetNegative(cpu.a)
	cpu.testAndSetZero(cpu.a)
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
	pc := cpu.pc - 1

	// push the lo then the hi bytes on to the stack
	hi := uint8(pc >> 8)
	lo := uint8(pc)

	cpu.push(hi)
	cpu.push(lo)

	cpu.pc = address
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
	cpu.p.set(P_Negative, false)
}

func (cpu *MOS6502) nop(ins *instruction, address uint16) {
	// No Operation
}

func (cpu *MOS6502) ora(ins *instruction, address uint16) {
	// Or Memory with Accumulator
	value := cpu.memory.Read(address)
	cpu.a = cpu.a | value

	cpu.testAndSetNegative(cpu.a)
	cpu.testAndSetZero(cpu.a)
}

func (cpu *MOS6502) pha(ins *instruction, address uint16) {
	// Push Accumulator on Stack
	cpu.push(cpu.a)
}

func (cpu *MOS6502) php(ins *instruction, address uint16) {
	// Push Processor Status on Stack
	// The status register will be pushed with the break
	// flag and bit 5 set to 1.
	p := uint8(cpu.p) | uint8(P_Break) | uint8(P_Reserved)
	cpu.push(p)
}

func (cpu *MOS6502) pla(ins *instruction, address uint16) {
	// Pull Accumulator from Stack
	cpu.a = cpu.pop()
	cpu.testAndSetNegative(cpu.a)
	cpu.testAndSetZero(cpu.a)
}

func (cpu *MOS6502) plp(ins *instruction, address uint16) {
	// Pull Processor Status from Stack
	p := cpu.pop()
	cpu.p = flags(p)
	cpu.p.set(P_Reserved, true)
}

func (cpu *MOS6502) rol(ins *instruction, address uint16) {
	// Rotate One Bit Left (Memory or Accumulator)
	accumulator := ins.mode == AM_IMPLIED

	// if we are immediate get from the accumulator
	value := cpu.a
	if !accumulator {
		value = cpu.memory.Read(address)
	}

	var c uint8 = 0
	if cpu.p.isSet(P_Carry) {
		c = 1
	}

	// roll left
	rolled := (uint16(value) << 1) | uint16(c)

	if accumulator {
		cpu.a = uint8(rolled)
	} else {
		cpu.memory[address] = uint8(rolled)
	}

	cpu.p.set(P_Carry, value&0x80 == 0x80)
	cpu.testAndSetNegative(uint8(rolled))
	cpu.testAndSetZero(uint8(rolled))
}

func (cpu *MOS6502) ror(ins *instruction, address uint16) {
	// Rotate One Bit Right (Memory or Accumulator)
	accumulator := ins.mode == AM_IMPLIED

	// if we are immediate get from the accumulator
	value := cpu.a
	if !accumulator {
		value = cpu.memory.Read(address)
	}

	var c uint8 = 0
	if cpu.p.isSet(P_Carry) {
		c = 1
	}

	// roll right
	rolled := uint16(value)>>1 | uint16(c)<<7

	if accumulator {
		cpu.a = uint8(rolled)
	} else {
		cpu.memory[address] = uint8(rolled)
	}

	cpu.p.set(P_Carry, value&0x01 == 0x01)
	cpu.testAndSetNegative(uint8(rolled))
	cpu.testAndSetZero(uint8(rolled))
}

func (cpu *MOS6502) rti(ins *instruction, address uint16) {
	// Return from Interrupt
	// pop the status register
	cpu.p = flags(cpu.pop())
	cpu.p.set(P_Reserved, true)
	cpu.p.set(P_Break, false)

	// pop the program counter
	lo := cpu.pop()
	hi := cpu.pop()

	cpu.pc = uint16(hi)<<8 | uint16(lo)
}

func (cpu *MOS6502) rts(ins *instruction, address uint16) {
	// Return from Subroutine
	// pop the program counter
	lo := cpu.pop()
	hi := cpu.pop()

	cpu.pc = (uint16(lo) | (uint16(hi) << 8))

	cpu.pc++ // Increment the program counter by 1
}

func (cpu *MOS6502) sbc(ins *instruction, address uint16) {
	// Subtract Memory from Accumulator with Borrow

	var c uint8 = 0
	if cpu.p.isSet(P_Carry) {
		c = 1
	}

	a := cpu.a
	m := cpu.memory.Read(address)

	// sum in uint16 to catch overflow
	sum := uint16(a) - uint16(m) + uint16(c)

	cpu.a = uint8(sum & 0xff)
	cpu.p.set(P_Carry, sum < 0x100)
	cpu.testAndSetNegative(cpu.a)
	cpu.testAndSetZero(cpu.a)

	// set overrflow
	cpu.p.set(P_Overflow, (a^m)&0x80 == 0x80 && (a^cpu.a)&0x80 == 0x80)
}

func (cpu *MOS6502) sec(ins *instruction, address uint16) {
	// Set Carry Flag
	cpu.p.set(P_Carry, true)
}

func (cpu *MOS6502) sed(ins *instruction, address uint16) {
	// Set Decimal Flag
	cpu.p.set(P_Decimal, true)
}

func (cpu *MOS6502) sei(ins *instruction, address uint16) {
	// Set Interrupt Disable Status
	cpu.p.set(P_InterruptDisable, true)
}

func (cpu *MOS6502) sta(ins *instruction, address uint16) {
	// Store Accumulator in Memory
	cpu.memory[address] = cpu.a
}

func (cpu *MOS6502) stx(ins *instruction, address uint16) {
	// Store Index X in Memory
	cpu.memory[address] = cpu.x
}

func (cpu *MOS6502) sty(ins *instruction, address uint16) {
	// Store Index Y in Memory
	cpu.memory[address] = cpu.y
}

func (cpu *MOS6502) tax(ins *instruction, address uint16) {
	// Transfer Accumulator to Index X
	cpu.x = cpu.a
	cpu.testAndSetNegative(cpu.x)
	cpu.testAndSetZero(cpu.x)
}

func (cpu *MOS6502) tay(ins *instruction, address uint16) {
	// Transfer Accumulator to Index Y
	cpu.y = cpu.a
	cpu.testAndSetNegative(cpu.y)
	cpu.testAndSetZero(cpu.y)
}

func (cpu *MOS6502) tsx(ins *instruction, address uint16) {
	// Transfer Stack Pointer to Index X
	cpu.x = uint8(cpu.sp)
	cpu.testAndSetNegative(cpu.x)
	cpu.testAndSetZero(cpu.x)
}

func (cpu *MOS6502) txa(ins *instruction, address uint16) {
	// Transfer Index X to Accumulator
	cpu.a = cpu.x
	cpu.testAndSetNegative(cpu.a)
	cpu.testAndSetZero(cpu.a)
}

func (cpu *MOS6502) txs(ins *instruction, address uint16) {
	// Transfer Index X to Stack Register
	cpu.sp = uint16(0x100) + uint16(cpu.x)
}

func (cpu *MOS6502) tya(ins *instruction, address uint16) {
	// Transfer Index Y to Accumulator
	cpu.a = cpu.y
	cpu.testAndSetNegative(cpu.a)
	cpu.testAndSetZero(cpu.a)
}
