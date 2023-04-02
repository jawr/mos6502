# mos6502

mostly based off the instruction set documented here: https://www.masswerk.at/6502/6502_instruction_set.html#CLI

# functional tests

functional tests taken from [6502_65C02_functional_tests](https://github.com/amb5l/6502_65C02_functional_tests) which is a ca65 port of [this repo](https://github.com/Klaus2m5/6502_65C02_functional_tests).

used to fully test functionality of the emulator. so far not passing :D

```
go run cmd/mos6502/main.go -rom testdata/6502_functional_test.bin -start 0x0400
```
