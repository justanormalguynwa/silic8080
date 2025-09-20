package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"
)

// cpu state i tink
type CPU struct {
	A, B, C, D, E, H, L uint8
	SP                  uint16
	PC                  uint16
	Flags               struct {
		Z, S, P, CY, AC bool
	}
	Memory []uint8
	Halted bool
}

// load ze fucking rom into the fucking memory
func (cpu *CPU) LoadROM(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("failed to load ROM: %v", err)
	}
	copy(cpu.Memory, data)
}

// fetch next byte you nword
func (cpu *CPU) fetch() uint8 {
	op := cpu.Memory[cpu.PC]
	cpu.PC++
	return op
}

// fetch lil endian word and deport sockman
func (cpu *CPU) fetchWord() uint16 {
	lo := cpu.fetch()
	hi := cpu.fetch()
	return binary.LittleEndian.Uint16([]byte{lo, hi})
}

// execute one instruction i implemented like 5 :mewhenthe:
func (cpu *CPU) step() {
	opcode := cpu.fetch()
	mnemonic := op[opcode] // from opcode.go nword
	fmt.Printf("PC=%04X OP=%02X %-10s\n", cpu.PC-1, opcode, mnemonic)

	switch opcode {
	case 0x00: // NOP
		// nothing
	case 0x76: // HLT
		cpu.Halted = true
	case 0x3E: // MVI A,D8
		cpu.A = cpu.fetch()
	case 0x06: // MVI B,D8
		cpu.B = cpu.fetch()
	case 0x0E: // MVI C,D8
		cpu.C = cpu.fetch()
	case 0x32: // STA adr
		addr := cpu.fetchWord()
		cpu.Memory[addr] = cpu.A
	case 0x3A: // LDA adr
		addr := cpu.fetchWord()
		cpu.A = cpu.Memory[addr]
	case 0xC3: // JMP adr
		addr := cpu.fetchWord()
		cpu.PC = addr
	default:
		fmt.Printf("Unimplemented opcode %02X (%s)\n", opcode, mnemonic)
		cpu.Halted = true
	}
}

// type in opcodes manually cus fuck you
func (cpu *CPU) manualMode() {
	fmt.Println("Manual opcode entry (hex bytes). Type 'q' to quit.")
	for {
		var s string
		fmt.Print("Enter opcode: ")
		_, err := fmt.Scanln(&s)
		if err != nil {
			return
		}
		if s == "q" {
			return
		}
		var b byte
		_, err = fmt.Sscanf(s, "%02X", &b)
		if err != nil {
			fmt.Println("Invalid hex")
			continue
		}
		cpu.Memory[cpu.PC] = b
		cpu.step()
	}
}

func main() {
	romPath := flag.String("rom", "", "Path to ROM .bin file")
	memSize := flag.Int("mem", 64*1024, "Memory size in bytes (default 64KB)")
	mhz := flag.Float64("mhz", 2.0, "Clock speed in MHz")
	flag.Parse()

	cpu := &CPU{
		Memory: make([]uint8, *memSize),
		PC:     0,
		SP:     uint16(*memSize - 1),
	}

	if *romPath != "" {
		cpu.LoadROM(*romPath)
	}


	// Main emulation loop
	for !cpu.Halted {
		<-ticker.C
		cpu.step()
	}

	fmt.Println("CPU halted.")
}
