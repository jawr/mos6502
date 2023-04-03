package cpu

import (
	"strings"
)

// flags
const (
	// C
	P_Carry flag = 1 << iota
	// Z
	P_Zero
	// I
	P_InterruptDisable
	// D
	P_Decimal
	// B
	P_Break
	// -
	P_Reserved
	// V
	P_Overflow
	// N
	P_Negative
)

type flag uint8
type flags uint8

func (a *flags) isSet(b flag) bool {
	return uint8(*a)&uint8(b) != 0x0
}

func (a *flags) set(b flag, v bool) {
	if v {
		*a = flags(uint8(*a) | uint8(b))
	} else {
		*a = flags(uint8(*a) &^ uint8(b))
	}
}

func stringer(b *strings.Builder, f string, set bool) {
	if !set {
		b.WriteString("-")
	} else {
		b.WriteString(f)
	}
}

func (a *flags) String() string {
	b := &strings.Builder{}
	stringer(b, "N", a.isSet(P_Negative))
	stringer(b, "V", a.isSet(P_Overflow))
	stringer(b, "-", a.isSet(P_Reserved))
	stringer(b, "B", a.isSet(P_Break))
	stringer(b, "D", a.isSet(P_Decimal))
	stringer(b, "I", a.isSet(P_InterruptDisable))
	stringer(b, "Z", a.isSet(P_Zero))
	stringer(b, "C", a.isSet(P_Carry))
	return b.String()
}

func (cpu *MOS6502) testAndSetNegative(b uint8) {
	cpu.p.set(P_Negative, b&0x80 == 0x80)
}

func (cpu *MOS6502) testAndSetZero(b uint8) {
	cpu.p.set(P_Zero, b == 0x0)
}

func (cpu *MOS6502) testAndSetCarry(b uint16) {
	// if b is bigger than an 8bit set the carry
	cpu.p.set(P_Carry, b > 0xff)
}

func (cpu *MOS6502) testAndSetOverflow(a, b, sum uint8) {
	// Calculate the overflow by checking if the sign bit of the operands and
	// the result differ (which indicates a signed overflow has occurred).
	cpu.p.set(P_Overflow, (a^sum)&(b^sum)&0x80 == 0x80)
}
