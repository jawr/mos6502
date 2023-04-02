package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/jawr/mos6502/cpu"
)

func main() {
	rom := flag.String("rom", "", "Path to ROM file")
	start := flag.Uint("start", uint(cpu.RESVectorLow), "Start address")
	flag.Parse()

	memory, err := loadROM(*rom)
	if err != nil {
		log.Printf("error loading ROM: %s", err)
		os.Exit(1)
	}

	// load memory into cpu
	cpu := cpu.NewMOS6502()
	cpu.Reset(memory)
	cpu.SetPC(uint16(*start))
	cpu.Debug = true
	cpu.TrapDetector = true

	// setup interrupt
	q := make(chan os.Signal, 1)
	signal.Notify(q, os.Interrupt)

	// setup clock
	clock := time.NewTicker(33 * time.Nanosecond)
	defer clock.Stop()

	log.Printf("Starting CPU...")

	// run cpu
	for {
		select {
		case <-q:
			os.Exit(0)
		case <-clock.C:
			cpu.Cycle()
			if cpu.Stop() {
				log.Printf("CPU stopped...")
				os.Exit(1)
			}

		}
	}
}

func loadROM(path string) (*cpu.Memory, error) {
	// open rom
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	stats, err := file.Stat()
	if err != nil {
		return nil, err
	}

	memory := &cpu.Memory{}

	if stats.Size() > int64(len(memory)) {
		return nil, fmt.Errorf("ROM too large. Wanted %d got %d", len(memory), stats.Size())
	}

	buff := make([]byte, stats.Size())
	reader := bufio.NewReader(file)

	_, err = reader.Read(buff)
	if err != nil {
		return nil, err
	}

	for i, ch := range buff {
		memory[i] = ch
	}

	log.Printf("Loaded ROM: %s (%d)", path, stats.Size())

	return memory, nil
}
