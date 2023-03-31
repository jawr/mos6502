package cpu

// flags
const (
	P_Carry flag = 1 << iota
	P_Zero
	P_InterruptDisable
	P_Decimal
	P_Break
	P_Reserved
	P_Overflow
	P_Negative
)

type flag uint8
type flags uint8

func (a flags) isSet(b flag) bool {
	return uint8(a)&uint8(b) != 0x0
}

func (a *flags) set(b flag) {
	*a = flags(uint8(*a) | uint8(b))
}

func (a *flags) clear(b flag) {
	*a = flags(uint8(*a) &^ uint8(b))
}

func (cpu *MOS6502) testAndSetNegative(b uint8) {
	if b&0x80 == 0x80 {
		cpu.p.set(P_Negative)
	} else {
		cpu.p.clear(P_Negative)
	}
}

func (cpu *MOS6502) testAndSetZero(b uint8) {
	if b == 0x0 {
		cpu.p.set(P_Zero)
	} else {
		cpu.p.clear(P_Zero)
	}
}

func (cpu *MOS6502) testAndSetCarry(b uint16) {
	// if b is bigger than an 8bit set the carry
	if b > 0xff {
		cpu.p.set(P_Carry)
	} else {
		cpu.p.clear(P_Carry)
	}
}

func (cpu *MOS6502) testAndSetOverflow(a, b, sum uint8) {
	// Calculate the overflow by checking if the sign bit of the operands and
	// the result differ (which indicates a signed overflow has occurred).
	overflow := (a^sum)&(b^sum)&0x80 == 0x80

	if overflow {
		cpu.p.set(P_Overflow)
	} else {
		cpu.p.clear(P_Overflow)
	}
}
