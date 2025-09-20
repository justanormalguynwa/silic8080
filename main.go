package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"os"
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
//
//	as hv why are we deporting stockman
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
	case 0x16: // MVI D,D8
		cpu.D = cpu.fetch()
	case 0x1E: // MVI E,D8
		cpu.E = cpu.fetch()
	case 0x26: // MVI H,D8
		cpu.H = cpu.fetch()
	case 0x2E: // MVI L,D8
		cpu.L = cpu.fetch()
	case 0x32: // STA adr
		addr := cpu.fetchWord()
		cpu.Memory[addr] = cpu.A
	case 0x3A: // LDA adr
		addr := cpu.fetchWord()
		cpu.A = cpu.Memory[addr]
	case 0xC3: // JMP adr
		addr := cpu.fetchWord()
		cpu.PC = addr
	case 0x1: // LXI B
		// PLACEHOLDER
	case 0x2: // STAX B
		// PLACEHOLDER
	case 0x3: // INX B
		// PLACEHOLDER
	case 0x4: // INR B
		// PLACEHOLDER
	case 0x5: // DCR B
		// PLACEHOLDER
	case 0x7: // RLC
		// PLACEHOLDER
	case 0x9: // DAD B
		// PLACEHOLDER
	case 0xa: // LDAX B
		// PLACEHOLDER
	case 0xb: // DCX B
		// PLACEHOLDER
	case 0xc: // INR C
		// PLACEHOLDER
	case 0xd: // DCR C
		// PLACEHOLDER
	case 0xf: // RRC
		// PLACEHOLDER
	case 0x11: // LXI D
		// PLACEHOLDER
	case 0x12: // STAX D
		// PLACEHOLDER
	case 0x13: // INX D
		// PLACEHOLDER
	case 0x14: // INR D
		// PLACEHOLDER
	case 0x15: // DCR D
		// PLACEHOLDER
	case 0x17: // RAL
		// PLACEHOLDER
	case 0x19: // DAD D
		// PLACEHOLDER
	case 0x1a: // LDAX D
		// PLACEHOLDER
	case 0x1b: // DCX D
		// PLACEHOLDER
	case 0x1c: // INR E
		// PLACEHOLDER
	case 0x1d: // DCR E
		// PLACEHOLDER
	case 0x1f: // RAR
		// PLACEHOLDER
	case 0x21: // LXI H
		// PLACEHOLDER
	case 0x22: // SHLD
		// PLACEHOLDER
	case 0x23: // INX H
		// PLACEHOLDER
	case 0x24: // INR H
		// PLACEHOLDER
	case 0x25: // DCR H
		// PLACEHOLDER
	case 0x27: // DAA
		// PLACEHOLDER
	case 0x29: // DAD H
		// PLACEHOLDER
	case 0x2a: // LHLD
		// PLACEHOLDER
	case 0x2b: // DCX H
		// PLACEHOLDER
	case 0x2c: // INR L
		// PLACEHOLDER
	case 0x2d: // DCR L
		// PLACEHOLDER
	case 0x2f: // CMA
		// PLACEHOLDER
	case 0x31: // LXI SP
		// PLACEHOLDER
	case 0x33: // INX SP
		// PLACEHOLDER
	case 0x34: // INR M
		// PLACEHOLDER
	case 0x35: // DCR M
		// PLACEHOLDER
	case 0x36: // MVI M
		// PLACEHOLDER
	case 0x37: // STC
		// PLACEHOLDER
	case 0x39: // DAD SP
		// PLACEHOLDER
	case 0x3b: // DCX SP
		// PLACEHOLDER
	case 0x3c: // INR A
		// PLACEHOLDER
	case 0x3d: // DCR A
		// PLACEHOLDER
	case 0x3f: // CMC
		// PLACEHOLDER
	case 0x40: // MOV B,B
		// PLACEHOLDER
	case 0x41: // MOV B,C
		// PLACEHOLDER
	case 0x42: // MOV B,D
		// PLACEHOLDER
	case 0x43: // MOV B,E
		// PLACEHOLDER
	case 0x44: // MOV B,H
		// PLACEHOLDER
	case 0x45: // MOV B,L
		// PLACEHOLDER
	case 0x46: // MOV B,M
		// PLACEHOLDER
	case 0x47: // MOV B,A
		// PLACEHOLDER
	case 0x48: // MOV C,B
		// PLACEHOLDER
	case 0x49: // MOV C,C
		// PLACEHOLDER
	case 0x4a: // MOV C,D
		// PLACEHOLDER
	case 0x4b: // MOV C,E
		// PLACEHOLDER
	case 0x4c: // MOV C,H
		// PLACEHOLDER
	case 0x4d: // MOV C,L
		// PLACEHOLDER
	case 0x4e: // MOV C,M
		// PLACEHOLDER
	case 0x4f: // MOV C,A
		// PLACEHOLDER
	case 0x50: // MOV D,B
		// PLACEHOLDER
	case 0x51: // MOV D,C
		// PLACEHOLDER
	case 0x52: // MOV D,D
		// PLACEHOLDER
	case 0x53: // MOV D,E
		// PLACEHOLDER
	case 0x54: // MOV D,H
		// PLACEHOLDER
	case 0x55: // MOV D,L
		// PLACEHOLDER
	case 0x56: // MOV D,M
		// PLACEHOLDER
	case 0x57: // MOV D,A
		// PLACEHOLDER
	case 0x58: // MOV E,B
		// PLACEHOLDER
	case 0x59: // MOV E,C
		// PLACEHOLDER
	case 0x5a: // MOV E,D
		// PLACEHOLDER
	case 0x5b: // MOV E,E
		// PLACEHOLDER
	case 0x5c: // MOV E,H
		// PLACEHOLDER
	case 0x5d: // MOV E,L
		// PLACEHOLDER
	case 0x5e: // MOV E,M
		// PLACEHOLDER
	case 0x5f: // MOV E,A
		// PLACEHOLDER
	case 0x60: // MOV H,B
		// PLACEHOLDER
	case 0x61: // MOV H,C
		// PLACEHOLDER
	case 0x62: // MOV H,D
		// PLACEHOLDER
	case 0x63: // MOV H,E
		// PLACEHOLDER
	case 0x64: // MOV H,H
		// PLACEHOLDER
	case 0x65: // MOV H,L
		// PLACEHOLDER
	case 0x66: // MOV H,M
		// PLACEHOLDER
	case 0x67: // MOV H,A
		// PLACEHOLDER
	case 0x68: // MOV L,B
		// PLACEHOLDER
	case 0x69: // MOV L,C
		// PLACEHOLDER
	case 0x6a: // MOV L,D
		// PLACEHOLDER
	case 0x6b: // MOV L,E
		// PLACEHOLDER
	case 0x6c: // MOV L,H
		// PLACEHOLDER
	case 0x6d: // MOV L,L
		// PLACEHOLDER
	case 0x6e: // MOV L,M
		// PLACEHOLDER
	case 0x6f: // MOV L,A
		// PLACEHOLDER
	case 0x70: // MOV M,B
		// PLACEHOLDER
	case 0x71: // MOV M,C
		// PLACEHOLDER
	case 0x72: // MOV M,D
		// PLACEHOLDER
	case 0x73: // MOV M,E
		// PLACEHOLDER
	case 0x74: // MOV M,H
		// PLACEHOLDER
	case 0x75: // MOV M,L
		// PLACEHOLDER
	case 0x77: // MOV M,A
		// PLACEHOLDER
	case 0x78: // MOV A,B
		// PLACEHOLDER
	case 0x79: // MOV A,C
		// PLACEHOLDER
	case 0x7a: // MOV A,D
		// PLACEHOLDER
	case 0x7b: // MOV A,E
		// PLACEHOLDER
	case 0x7c: // MOV A,H
		// PLACEHOLDER
	case 0x7d: // MOV A,L
		// PLACEHOLDER
	case 0x7e: // MOV A,M
		// PLACEHOLDER
	case 0x7f: // MOV A,A
		// PLACEHOLDER
	case 0x80: // ADD B
		// PLACEHOLDER
	case 0x81: // ADD C
		// PLACEHOLDER
	case 0x82: // ADD D
		// PLACEHOLDER
	case 0x83: // ADD E
		// PLACEHOLDER
	case 0x84: // ADD H
		// PLACEHOLDER
	case 0x85: // ADD L
		// PLACEHOLDER
	case 0x86: // ADD M
		// PLACEHOLDER
	case 0x87: // ADD A
		// PLACEHOLDER
	case 0x88: // ADC B
		// PLACEHOLDER
	case 0x89: // ADC C
		// PLACEHOLDER
	case 0x8a: // ADC D
		// PLACEHOLDER
	case 0x8b: // ADC E
		// PLACEHOLDER
	case 0x8c: // ADC H
		// PLACEHOLDER
	case 0x8d: // ADC L
		// PLACEHOLDER
	case 0x8e: // ADC M
		// PLACEHOLDER
	case 0x8f: // ADC A
		// PLACEHOLDER
	case 0x90: // SUB B
		// PLACEHOLDER
	case 0x91: // SUB C
		// PLACEHOLDER
	case 0x92: // SUB D
		// PLACEHOLDER
	case 0x93: // SUB E
		// PLACEHOLDER
	case 0x94: // SUB H
		// PLACEHOLDER
	case 0x95: // SUB L
		// PLACEHOLDER
	case 0x96: // SUB M
		// PLACEHOLDER
	case 0x97: // SUB A
		// PLACEHOLDER
	case 0x98: // SBB B
		// PLACEHOLDER
	case 0x99: // SBB C
		// PLACEHOLDER
	case 0x9a: // SBB D
		// PLACEHOLDER
	case 0x9b: // SBB E
		// PLACEHOLDER
	case 0x9c: // SBB H
		// PLACEHOLDER
	case 0x9d: // SBB L
		// PLACEHOLDER
	case 0x9e: // SBB M
		// PLACEHOLDER
	case 0x9f: // SBB A
		// PLACEHOLDER
	case 0xa0: // ANA B
		// PLACEHOLDER
	case 0xa1: // ANA C
		// PLACEHOLDER
	case 0xa2: // ANA D
		// PLACEHOLDER
	case 0xa3: // ANA E
		// PLACEHOLDER
	case 0xa4: // ANA H
		// PLACEHOLDER
	case 0xa5: // ANA L
		// PLACEHOLDER
	case 0xa6: // ANA M
		// PLACEHOLDER
	case 0xa7: // ANA A
		// PLACEHOLDER
	case 0xa8: // XRA B
		// PLACEHOLDER
	case 0xa9: // XRA C
		// PLACEHOLDER
	case 0xaa: // XRA D
		// PLACEHOLDER
	case 0xab: // XRA E
		// PLACEHOLDER
	case 0xac: // XRA H
		// PLACEHOLDER
	case 0xad: // XRA L
		// PLACEHOLDER
	case 0xae: // XRA M
		// PLACEHOLDER
	case 0xaf: // XRA A
		// PLACEHOLDER
	case 0xb0: // ORA B
		// PLACEHOLDER
	case 0xb1: // ORA C
		// PLACEHOLDER
	case 0xb2: // ORA D
		// PLACEHOLDER
	case 0xb3: // ORA E
		// PLACEHOLDER
	case 0xb4: // ORA H
		// PLACEHOLDER
	case 0xb5: // ORA L
		// PLACEHOLDER
	case 0xb6: // ORA M
		// PLACEHOLDER
	case 0xb7: // ORA A
		// PLACEHOLDER
	case 0xb8: // CMP B
		// PLACEHOLDER
	case 0xb9: // CMP C
		// PLACEHOLDER
	case 0xba: // CMP D
		// PLACEHOLDER
	case 0xbb: // CMP E
		// PLACEHOLDER
	case 0xbc: // CMP H
		// PLACEHOLDER
	case 0xbd: // CMP L
		// PLACEHOLDER
	case 0xbe: // CMP M
		// PLACEHOLDER
	case 0xbf: // CMP A
		// PLACEHOLDER
	case 0xc0: // RNZ
		// PLACEHOLDER
	case 0xc1: // POP B
		// PLACEHOLDER
	case 0xc2: // JNZ
		// PLACEHOLDER
	case 0xc4: // CNZ
		// PLACEHOLDER
	case 0xc5: // PUSH B
		// PLACEHOLDER
	case 0xc6: // ADI D8
		// PLACEHOLDER
	case 0xc7: // RST 0
		// PLACEHOLDER
	case 0xc8: // RZ
		// PLACEHOLDER
	case 0xc9: // RET
		// PLACEHOLDER
	case 0xca: // JZ
		// PLACEHOLDER
	case 0xcc: // CZ
		// PLACEHOLDER
	case 0xcd: // CALL
		// PLACEHOLDER
	case 0xce: // ACI D8
		// PLACEHOLDER
	case 0xcf: // RST 1
		// PLACEHOLDER
	case 0xd0: // RNC
		// PLACEHOLDER
	case 0xd1: // POP D
		// PLACEHOLDER
	case 0xd2: // JNC
		// PLACEHOLDER
	case 0xd3: // OUT D8
		// PLACEHOLDER
	case 0xd4: // CNC
		// PLACEHOLDER
	case 0xd5: // PUSH D
		// PLACEHOLDER
	case 0xd6: // SUI D8
		// PLACEHOLDER
	case 0xd7: // RST 2
		// PLACEHOLDER
	case 0xd8: // RC
		// PLACEHOLDER
	case 0xda: // JC
		// PLACEHOLDER
	case 0xdb: // IN D8
		// PLACEHOLDER
	case 0xdc: // CC
		// PLACEHOLDER
	case 0xde: // SBI D8
		// PLACEHOLDER
	case 0xdf: // RST 3
		// PLACEHOLDER
	case 0xe0: // RPO
		// PLACEHOLDER
	case 0xe1: // POP H
		// PLACEHOLDER
	case 0xe2: // JPO
		// PLACEHOLDER
	case 0xe3: // XTHL
		// PLACEHOLDER
	case 0xe4: // CPO
		// PLACEHOLDER
	case 0xe5: // PUSH H
		// PLACEHOLDER
	case 0xe6: // ANI D8
		// PLACEHOLDER
	case 0xe7: // RST 4
		// PLACEHOLDER
	case 0xe8: // RPE
		// PLACEHOLDER
	case 0xe9: // PCHL
		// PLACEHOLDER
	case 0xea: // JPE
		// PLACEHOLDER
	case 0xeb: // XCHG
		// PLACEHOLDER
	case 0xec: // CPE
		// PLACEHOLDER
	case 0xee: // XRI D8
		// PLACEHOLDER
	case 0xef: // RST 5
		// PLACEHOLDER
	case 0xf0: // RP
		// PLACEHOLDER
	case 0xf1: // POP PSW
		// PLACEHOLDER
	case 0xf2: // JP
		// PLACEHOLDER
	case 0xf3: // DI
		// PLACEHOLDER
	case 0xf4: // CP
		// PLACEHOLDER
	case 0xf5: // PUSH PSW
		// PLACEHOLDER
	case 0xf6: // ORI D8
		// PLACEHOLDER
	case 0xf7: // RST 6
		// PLACEHOLDER
	case 0xf8: // RM
		// PLACEHOLDER
	case 0xf9: // SPHL
		// PLACEHOLDER
	case 0xfa: // JM
		// PLACEHOLDER
	case 0xfb: // EI
		// PLACEHOLDER
	case 0xfc: // CM
		// PLACEHOLDER
	case 0xfe: // CPI D8
		// PLACEHOLDER
	case 0xff: // RST 7
		// PLACEHOLDER
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
