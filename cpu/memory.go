package cpu

/*
65k of memory, 256 pages with
256 bytes per page

addresses are 16 bits in length with the
first byte referencing the page (up to 255)
and the second byte referncing the offset
on the page (up to 255)

1
2 6 3 1
8 4 2 6 8 4 2 1
---------------
0 0 0 0 0 0 0 0

page 0 aka Zero Page is a special
page for quick access as accessing
it only requires 1 byte address rather than
two

the last six bytes of the last page (page 255)
have special addresses but are still considered ROM:

- interrupt handlers for (IRC and NMI)
- reset handler (starting point for the processor)

overall the lower half of memory is RAM and the
upper half is ROM
*/
type Memory [0x100 * 0x100]uint8

func (m *Memory) Read(address uint16) uint8 {
	// reads a 1 byte address
	return m[address]
}

func (m *Memory) ReadWord(address uint16) uint16 {
	// takes a 2 byte address and returns a 2 byte address
	return uint16(m[address]) + (uint16(m[address+1]) << 8)
}
