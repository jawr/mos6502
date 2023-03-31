package cpu

import (
	"testing"
)

// helper function to test a flag is set to an expected value
func expectFlag(t *testing.T, cpu *MOS6502, f flag, expect bool) {
	t.Helper()

	if expect != cpu.p.isSet(f) {
		t.Errorf("expected p=%08b expected: %t got: %t", f, expect, cpu.p.isSet(f))
	}
}
