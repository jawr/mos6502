package cpu

const trapDetectorBufferSize = 16

type trapDetector struct {
	buffer [trapDetectorBufferSize]uint16
	index  int
}

func (ld *trapDetector) push(value uint16) {
	ld.buffer[ld.index] = value
	ld.index = (ld.index + 1) % trapDetectorBufferSize
}

func (ld *trapDetector) hastrap() bool {
	for i := 0; i < trapDetectorBufferSize/2; i++ {
		if ld.buffer[i] != ld.buffer[i+trapDetectorBufferSize/2] {
			return false
		}
	}
	return true
}
