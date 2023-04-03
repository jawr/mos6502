package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/jawr/mos6502/cpu"
	mos6502 "github.com/jawr/mos6502/cpu"
	term "github.com/nsf/termbox-go"
)

func main() {
	rom := flag.String("rom", "", "Path to ROM file")
	start := flag.Uint("start", uint(mos6502.RESVectorLow), "Start address")
	stop := flag.Uint("stop", 0, "Stop address")
	debug := flag.Bool("debug", false, "Output each step")
	trapDetector := flag.Bool("trapDetector", false, "Detect traps and stop")

	flag.Parse()

	memory, err := loadROM(*rom)
	if err != nil {
		log.Printf("error loading ROM: %s", err)
		os.Exit(1)
	}

	// load memory into cpu
	cpu := mos6502.NewMOS6502()
	cpu.Reset(memory)
	cpu.SetPC(uint16(*start))

	if *stop != 0 {
		cpu.StopOnPC = uint16(*stop)
	}
	cpu.Debug = *debug
	cpu.TrapDetector = *trapDetector

	// setup interrupt
	q := make(chan os.Signal, 1)
	signal.Notify(q, os.Interrupt)

	log.Printf("Starting CPU...")

	// used for stepping through cpu
	step := false

	// run cpu
MainLoop:
	for {
		if step {
			ev := term.PollEvent()
			if ev.Type != term.EventKey {
				log.Printf("event: %v", ev)
				os.Exit(1)
			}

			switch ev.Key {
			case term.KeyEnter:
				break
			case term.KeyCtrlC:
				term.Close()
				break MainLoop
			}
		}

		select {
		case <-q:
			log.Printf("CTRL-C pressed...")
			// if first ctrl c and debug drop in to step mode
			if !step && *debug {
				log.Printf("Entering step mode...")

				// setup term
				err = term.Init()
				if err != nil {
					log.Printf("error initializing termbox: %s", err)
					os.Exit(1)
				}

				step = true

				continue MainLoop
			}
			break MainLoop
		default:

			// if we are in step mode we should wait for enter key to
			// be pressed before continuing

			cpu.Cycle()

			if cpu.Halt() != mos6502.Continue {
				break MainLoop
			}

		}
	}

	log.Printf("CPU stopped...")
	log.Printf("--------------")
	log.Printf("Total Cycles: %d", cpu.TotalCycles)
	log.Printf("--------------")

	code := 0
	switch cpu.Halt() {
	case mos6502.Continue:
		log.Printf("CPU manually stopped")
	case mos6502.HaltSuccess:
		log.Printf("CPU hit stop PC successfully")
	case mos6502.HaltTrap:
		log.Printf("CPU halted on trap")
	case mos6502.HaltUnknownInstruction:
		log.Printf("CPU halted on unknown instruction")
	}

	if cpu.Halt() != mos6502.HaltSuccess {
		code = 1
	}
	os.Exit(code)

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
