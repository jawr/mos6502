# mos6502

mostly based off the instruction set documented here: https://www.masswerk.at/6502/6502_instruction_set.html as well as using GPT-4 to help write tests/debug issues and github copilot to write out the instruction table (which also introduced a bad opcode!).

# clock

originally there was a system in place to run each instruction cycle and page boundary cross on a clock tick. however, the system was removed to speed up testing.

# functional tests

functional tests taken from [6502_65C02_functional_tests](https://github.com/amb5l/6502_65C02_functional_tests) which is a ca65 port of [this repo](https://github.com/Klaus2m5/6502_65C02_functional_tests).

```
go run cmd/mos6502/main.go -trapDetector -stop 0x00336D -start 0x0400 -rom testdata/6502_functional_test.bin
```

available options:

```
  -debug
        Output each step
  -rom string
        Path to ROM file
  -start uint
        Start address (default 65532)
  -stop uint
        Stop address
  -trapDetector
        Detect traps and stop
```

output on M1 Pro/32GB:

```
2023/04/03 15:22:31 Loaded ROM: testdata/6502_functional_test.bin (65536)
2023/04/03 15:22:31 Starting CPU...
2023/04/03 15:22:32 CPU stopped...
2023/04/03 15:22:32 --------------
2023/04/03 15:22:32 Total Cycles: 83799852
2023/04/03 15:22:32 --------------
2023/04/03 15:22:32 CPU hit stop PC successfully

________________________________________________________
Executed in  568.25 millis    fish           external
   usr time  490.94 millis    0.17 millis  490.76 millis
   sys time  152.01 millis    1.50 millis  150.52 millis
```
