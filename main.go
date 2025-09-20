package main

import (
	"encoding/binary"
	"flag"
	"math/bits"
	"fmt"
	"log"
	"os"
	"time"
)

const (
	flagCY uint8 = 1 << 0
	flagP  uint8 = 1 << 2
	flagAC uint8 = 1 << 4
	flagZ  uint8 = 1 << 6
	flagS  uint8 = 1 << 7
)


// cpu state i tink
type CPU struct {
	A, B, C, D, E, H, L uint8
	SP                  uint16
	PC                  uint16
	Flags               uint8 
	Memory              []uint8
	Halted              bool
	InterruptsEnabled   bool
}

func (cpu *CPU) setFlag(flag uint8, v bool) {
	if v {
		cpu.Flags |= flag
	} else {
		cpu.Flags &^= flag
	}
}

func (cpu *CPU) getFlag(flag uint8) bool {
	return (cpu.Flags & flag) != 0
}

// setZSP sets the zero sign and parity flags based on a result
func (cpu *CPU) setZSP(res uint8) {
	cpu.setFlag(flagZ, res == 0)
	cpu.setFlag(flagS, res&0x80 != 0)
	cpu.setFlag(flagP, bits.OnesCount8(res)%2 == 0)
}

// get/set 16bit register pairs
func (cpu *CPU) getBC() uint16 { return binary.BigEndian.Uint16([]byte{cpu.B, cpu.C}) }
func (cpu *CPU) setBC(v uint16) {
	cpu.B = uint8(v >> 8)
	cpu.C = uint8(v)
}

func (cpu *CPU) getDE() uint16 { return binary.BigEndian.Uint16([]byte{cpu.D, cpu.E}) }
func (cpu *CPU) setDE(v uint16) {
	cpu.D = uint8(v >> 8)
	cpu.E = uint8(v)
}

func (cpu *CPU) getHL() uint16 { return binary.BigEndian.Uint16([]byte{cpu.H, cpu.L}) }
func (cpu *CPU) setHL(v uint16) {
	cpu.H = uint8(v >> 8)
	cpu.L = uint8(v)
}

// M is the memory location pointed to by HL not a real register.
func (cpu *CPU) readM() uint8  { return cpu.Memory[cpu.getHL()] }
func (cpu *CPU) writeM(v uint8) { cpu.Memory[cpu.getHL()] = v }

// the PSW is the A and flags registers together
func (cpu *CPU) getPSW() uint16 { return binary.BigEndian.Uint16([]byte{cpu.A, cpu.Flags}) }
func (cpu *CPU) setPSW(v uint16) {
	cpu.A = uint8(v >> 8)
	// bit 1 is always 1 bit 3 and 5 are always 0 some programs rely on this
	cpu.Flags = (uint8(v) & 0xD5) | 0x02
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

// push a word onto the stack
func (cpu *CPU) push(val uint16) {
	hi := uint8(val >> 8)
	lo := uint8(val)
	cpu.Memory[cpu.SP-1] = hi
	cpu.Memory[cpu.SP-2] = lo
	cpu.SP -= 2
}

// pop a word from the stack
func (cpu *CPU) pop() uint16 {
	lo := cpu.Memory[cpu.SP]
	hi := cpu.Memory[cpu.SP+1]
	cpu.SP += 2
	return binary.LittleEndian.Uint16([]byte{lo, hi})
}


// execute one instruction
func (cpu *CPU) step() {
	if cpu.Halted {
		return
	}

	opcode := cpu.fetch()
	mnemonic := op[opcode]
	fmt.Printf("PC=%04X OP=%02X %-10s A=%02X B=%02X C=%02X D=%02X E=%02X H=%02X L=%02X SP=%04X CY=%t Z=%t S=%t P=%t AC=%t\n",
	cpu.PC-1, opcode, mnemonic, cpu.A, cpu.B, cpu.C, cpu.D, cpu.E, cpu.H, cpu.L, cpu.SP,
	cpu.getFlag(flagCY), cpu.getFlag(flagZ), cpu.getFlag(flagS), cpu.getFlag(flagP), cpu.getFlag(flagAC))

switch opcode {
	// nop and hlt
	case 0x00: // NOP
		// its literally no operation it wont do shit
	case 0x76: // HLT
		cpu.Halted = true

	// 8bit mvis
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
	case 0x36: // MVI M,D8
		cpu.writeM(cpu.fetch())
	case 0x3E: // MVI A,D8
		cpu.A = cpu.fetch()

	// 8bit movs
	case 0x40: // MOV B,B
		cpu.B = cpu.B // lol
	case 0x41: // MOV B,C
		cpu.B = cpu.C
	case 0x42: // MOV B,D
		cpu.B = cpu.D
	case 0x43: // MOV B,E
		cpu.B = cpu.E
	case 0x44: // MOV B,H
		cpu.B = cpu.H
	case 0x45: // MOV B,L
		cpu.B = cpu.L
	case 0x46: // MOV B,M
		cpu.B = cpu.readM()
	case 0x47: // MOV B,A
		cpu.B = cpu.A
	case 0x48: // MOV C,B
		cpu.C = cpu.B
	case 0x49: // MOV C,C
		cpu.C = cpu.C
	case 0x4A: // MOV C,D
		cpu.C = cpu.D
	case 0x4B: // MOV C,E
		cpu.C = cpu.E
	case 0x4C: // MOV C,H
		cpu.C = cpu.H
	case 0x4D: // MOV C,L
		cpu.C = cpu.L
	case 0x4E: // MOV C,M
		cpu.C = cpu.readM()
	case 0x4F: // MOV C,A
		cpu.C = cpu.A
	case 0x50: // MOV D,B
		cpu.D = cpu.B
	case 0x51: // MOV D,C
		cpu.D = cpu.C
	case 0x52: // MOV D,D
		cpu.D = cpu.D
	case 0x53: // MOV D,E
		cpu.D = cpu.E
	case 0x54: // MOV D,H
		cpu.D = cpu.H
	case 0x55: // MOV D,L
		cpu.D = cpu.L
	case 0x56: // MOV D,M
		cpu.D = cpu.readM()
	case 0x57: // MOV D,A
		cpu.D = cpu.A
	case 0x58: // MOV E,B
		cpu.E = cpu.B
	case 0x59: // MOV E,C
		cpu.E = cpu.C
	case 0x5A: // MOV E,D
		cpu.E = cpu.D
	case 0x5B: // MOV E,E
		cpu.E = cpu.E
	case 0x5C: // MOV E,H
		cpu.E = cpu.H
	case 0x5D: // MOV E,L
		cpu.E = cpu.L
	case 0x5E: // MOV E,M
		cpu.E = cpu.readM()
	case 0x5F: // MOV E,A
		cpu.E = cpu.A
	case 0x60: // MOV H,B
		cpu.H = cpu.B
	case 0x61: // MOV H,C
		cpu.H = cpu.C
	case 0x62: // MOV H,D
		cpu.H = cpu.D
	case 0x63: // MOV H,E
		cpu.H = cpu.E
	case 0x64: // MOV H,H
		cpu.H = cpu.H
	case 0x65: // MOV H,L
		cpu.H = cpu.L
	case 0x66: // MOV H,M
		cpu.H = cpu.readM()
	case 0x67: // MOV H,A
		cpu.H = cpu.A
	case 0x68: // MOV L,B
		cpu.L = cpu.B
	case 0x69: // MOV L,C
		cpu.L = cpu.C
	case 0x6A: // MOV L,D
		cpu.L = cpu.D
	case 0x6B: // MOV L,E
		cpu.L = cpu.E
	case 0x6C: // MOV L,H
		cpu.L = cpu.H
	case 0x6D: // MOV L,L
		cpu.L = cpu.L
	case 0x6E: // MOV L,M
		cpu.L = cpu.readM()
	case 0x6F: // MOV L,A
		cpu.L = cpu.A
	case 0x70: // MOV M,B
		cpu.writeM(cpu.B)
	case 0x71: // MOV M,C
		cpu.writeM(cpu.C)
	case 0x72: // MOV M,D
		cpu.writeM(cpu.D)
	case 0x73: // MOV M,E
		cpu.writeM(cpu.E)
	case 0x74: // MOV M,H
		cpu.writeM(cpu.H)
	case 0x75: // MOV M,L
		cpu.writeM(cpu.L)
	case 0x77: // MOV M,A
		cpu.writeM(cpu.A)
	case 0x78: // MOV A,B
		cpu.A = cpu.B
	case 0x79: // MOV A,C
		cpu.A = cpu.C
	case 0x7A: // MOV A,D
		cpu.A = cpu.D
	case 0x7B: // MOV A,E
		cpu.A = cpu.E
	case 0x7C: // MOV A,H
		cpu.A = cpu.H
	case 0x7D: // MOV A,L
		cpu.A = cpu.L
	case 0x7E: // MOV A,M
		cpu.A = cpu.readM()
	case 0x7F: // MOV A,A
		cpu.A = cpu.A

	// 16 bit lxis
	case 0x01: // LXI B,D16
		cpu.setBC(cpu.fetchWord())
	case 0x11: // LXI D,D16
		cpu.setDE(cpu.fetchWord())
	case 0x21: // LXI H,D16
		cpu.setHL(cpu.fetchWord())
	case 0x31: // LXI SP,D16
		cpu.SP = cpu.fetchWord()

	// dma
	case 0x22: // SHLD adr
		addr := cpu.fetchWord()
		cpu.Memory[addr] = cpu.L
		cpu.Memory[addr+1] = cpu.H
	case 0x2A: // LHLD adr
		addr := cpu.fetchWord()
		cpu.L = cpu.Memory[addr]
		cpu.H = cpu.Memory[addr+1]
	case 0x32: // STA adr
		addr := cpu.fetchWord()
		cpu.Memory[addr] = cpu.A
	case 0x3A: // LDA adr
		addr := cpu.fetchWord()
		cpu.A = cpu.Memory[addr]

	// indirect memory aces
	case 0x02: // STAX B
		cpu.Memory[cpu.getBC()] = cpu.A
	case 0x0A: // LDAX B
		cpu.A = cpu.Memory[cpu.getBC()]
	case 0x12: // STAX D
		cpu.Memory[cpu.getDE()] = cpu.A
	case 0x1A: // LDAX D
		cpu.A = cpu.Memory[cpu.getDE()]

	// 8bit decremtnt and tings
	case 0x04: // INR B
		cpu.B++
		cpu.setZSP(cpu.B)
		cpu.setFlag(flagAC, (cpu.B&0x0F) == 0x00)
	case 0x05: // DCR B
		cpu.B--
		cpu.setZSP(cpu.B)
		cpu.setFlag(flagAC, (cpu.B&0x0F) != 0x0F)
	case 0x0C: // INR C
		cpu.C++
		cpu.setZSP(cpu.C)
		cpu.setFlag(flagAC, (cpu.C&0x0F) == 0x00)
	case 0x0D: // DCR C
		cpu.C--
		cpu.setZSP(cpu.C)
		cpu.setFlag(flagAC, (cpu.C&0x0F) != 0x0F)
	case 0x14: // INR D
		cpu.D++
		cpu.setZSP(cpu.D)
		cpu.setFlag(flagAC, (cpu.D&0x0F) == 0x00)
	case 0x15: // DCR D
		cpu.D--
		cpu.setZSP(cpu.D)
		cpu.setFlag(flagAC, (cpu.D&0x0F) != 0x0F)
	case 0x1C: // INR E
		cpu.E++
		cpu.setZSP(cpu.E)
		cpu.setFlag(flagAC, (cpu.E&0x0F) == 0x00)
	case 0x1D: // DCR E
		cpu.E--
		cpu.setZSP(cpu.E)
		cpu.setFlag(flagAC, (cpu.E&0x0F) != 0x0F)
	case 0x24: // INR H
		cpu.H++
		cpu.setZSP(cpu.H)
		cpu.setFlag(flagAC, (cpu.H&0x0F) == 0x00)
	case 0x25: // DCR H
		cpu.H--
		cpu.setZSP(cpu.H)
		cpu.setFlag(flagAC, (cpu.H&0x0F) != 0x0F)
	case 0x2C: // INR L
		cpu.L++
		cpu.setZSP(cpu.L)
		cpu.setFlag(flagAC, (cpu.L&0x0F) == 0x00)
	case 0x2D: // DCR L
		cpu.L--
		cpu.setZSP(cpu.L)
		cpu.setFlag(flagAC, (cpu.L&0x0F) != 0x0F)
	case 0x34: // INR M
		res := cpu.readM() + 1
		cpu.writeM(res)
		cpu.setZSP(res)
		cpu.setFlag(flagAC, (res&0x0F) == 0x00)
	case 0x35: // DCR M
		res := cpu.readM() - 1
		cpu.writeM(res)
		cpu.setZSP(res)
		cpu.setFlag(flagAC, (res&0x0F) != 0x0F)
	case 0x3C: // INR A
		cpu.A++
		cpu.setZSP(cpu.A)
		cpu.setFlag(flagAC, (cpu.A&0x0F) == 0x00)
	case 0x3D: // DCR A
		cpu.A--
		cpu.setZSP(cpu.A)
		cpu.setFlag(flagAC, (cpu.A&0x0F) != 0x0F)

	// same ting as above but 16 bit
	case 0x03: // INX B
		cpu.setBC(cpu.getBC() + 1)
	case 0x0B: // DCX B
		cpu.setBC(cpu.getBC() - 1)
	case 0x13: // INX D
		cpu.setDE(cpu.getDE() + 1)
	case 0x1B: // DCX D
		cpu.setDE(cpu.getDE() - 1)
	case 0x23: // INX H
		cpu.setHL(cpu.getHL() + 1)
	case 0x2B: // DCX H
		cpu.setHL(cpu.getHL() - 1)
	case 0x33: // INX SP
		cpu.SP++
	case 0x3B: // DCX SP
		cpu.SP--

	// 8bit addition
	case 0x80, 0x81, 0x82, 0x83, 0x84, 0x85, 0x86, 0x87: // ADD r
		var val uint8
		switch opcode & 0x07 {
		case 0: val = cpu.B
		case 1: val = cpu.C
		case 2: val = cpu.D
		case 3: val = cpu.E
		case 4: val = cpu.H
		case 5: val = cpu.L
		case 6: val = cpu.readM()
		case 7: val = cpu.A
		}
		res16 := uint16(cpu.A) + uint16(val)
		res8 := uint8(res16)
		cpu.setFlag(flagCY, res16 > 0xFF)
		cpu.setFlag(flagAC, (cpu.A&0x0F)+(val&0x0F) > 0x0F)
		cpu.setZSP(res8)
		cpu.A = res8
	case 0xC6: // ADI D8
		val := cpu.fetch()
		res16 := uint16(cpu.A) + uint16(val)
		res8 := uint8(res16)
		cpu.setFlag(flagCY, res16 > 0xFF)
		cpu.setFlag(flagAC, (cpu.A&0x0F)+(val&0x0F) > 0x0F)
		cpu.setZSP(res8)
		cpu.A = res8

	// 8bit addition with carry
	case 0x88, 0x89, 0x8A, 0x8B, 0x8C, 0x8D, 0x8E, 0x8F: // ADC r
		var val uint8
		switch opcode & 0x07 {
		case 0: val = cpu.B
		case 1: val = cpu.C
		case 2: val = cpu.D
		case 3: val = cpu.E
		case 4: val = cpu.H
		case 5: val = cpu.L
		case 6: val = cpu.readM()
		case 7: val = cpu.A
		}
		carry := uint8(0)
		if cpu.getFlag(flagCY) {
			carry = 1
		}
		res16 := uint16(cpu.A) + uint16(val) + uint16(carry)
		res8 := uint8(res16)
		cpu.setFlag(flagCY, res16 > 0xFF)
		cpu.setFlag(flagAC, (cpu.A&0x0F)+(val&0x0F)+carry > 0x0F)
		cpu.setZSP(res8)
		cpu.A = res8
	case 0xCE: // ACI D8
		val := cpu.fetch()
		carry := uint8(0)
		if cpu.getFlag(flagCY) {
			carry = 1
		}
		res16 := uint16(cpu.A) + uint16(val) + uint16(carry)
		res8 := uint8(res16)
		cpu.setFlag(flagCY, res16 > 0xFF)
		cpu.setFlag(flagAC, (cpu.A&0x0F)+(val&0x0F)+carry > 0x0F)
		cpu.setZSP(res8)
		cpu.A = res8

	// 8bit subtract
	case 0x90, 0x91, 0x92, 0x93, 0x94, 0x95, 0x96, 0x97: // SUB r
		var val uint8
		switch opcode & 0x07 {
		case 0: val = cpu.B
		case 1: val = cpu.C
		case 2: val = cpu.D
		case 3: val = cpu.E
		case 4: val = cpu.H
		case 5: val = cpu.L
		case 6: val = cpu.readM()
		case 7: val = cpu.A
		}
		res16 := uint16(cpu.A) - uint16(val)
		res8 := uint8(res16)
		cpu.setFlag(flagCY, res16 > 0xFF) // Or val > cpu.A
		cpu.setFlag(flagAC, (cpu.A&0x0F) < (val&0x0F))
		cpu.setZSP(res8)
		cpu.A = res8
	case 0xD6: // SUI D8
		val := cpu.fetch()
		res16 := uint16(cpu.A) - uint16(val)
		res8 := uint8(res16)
		cpu.setFlag(flagCY, res16 > 0xFF)
		cpu.setFlag(flagAC, (cpu.A&0x0F) < (val&0x0F))
		cpu.setZSP(res8)
		cpu.A = res8

	// as above but with borrow
	case 0x98, 0x99, 0x9A, 0x9B, 0x9C, 0x9D, 0x9E, 0x9F: // SBB r
		var val uint8
		switch opcode & 0x07 {
		case 0: val = cpu.B
		case 1: val = cpu.C
		case 2: val = cpu.D
		case 3: val = cpu.E
		case 4: val = cpu.H
		case 5: val = cpu.L
		case 6: val = cpu.readM()
		case 7: val = cpu.A
		}
		borrow := uint8(0)
		if cpu.getFlag(flagCY) {
			borrow = 1
		}
		res16 := uint16(cpu.A) - uint16(val) - uint16(borrow)
		res8 := uint8(res16)
		cpu.setFlag(flagCY, res16 > 0xFF)
		cpu.setFlag(flagAC, (cpu.A&0x0F) < (val&0x0F)+borrow)
		cpu.setZSP(res8)
		cpu.A = res8
	case 0xDE: // SBI D8
		val := cpu.fetch()
		borrow := uint8(0)
		if cpu.getFlag(flagCY) {
			borrow = 1
		}
		res16 := uint16(cpu.A) - uint16(val) - uint16(borrow)
		res8 := uint8(res16)
		cpu.setFlag(flagCY, res16 > 0xFF)
		cpu.setFlag(flagAC, (cpu.A&0x0F) < (val&0x0F)+borrow)
		cpu.setZSP(res8)
		cpu.A = res8

	// 16 bit addition i dont want to do comments anymore
	case 0x09: // DAD B
		hl := uint32(cpu.getHL())
		bc := uint32(cpu.getBC())
		res := hl + bc
		cpu.setFlag(flagCY, res > 0xFFFF)
		cpu.setHL(uint16(res))
	case 0x19: // DAD D
		hl := uint32(cpu.getHL())
		de := uint32(cpu.getDE())
		res := hl + de
		cpu.setFlag(flagCY, res > 0xFFFF)
		cpu.setHL(uint16(res))
	case 0x29: // DAD H
		hl := uint32(cpu.getHL())
		res := hl + hl
		cpu.setFlag(flagCY, res > 0xFFFF)
		cpu.setHL(uint16(res))
	case 0x39: // DAD SP
		hl := uint32(cpu.getHL())
		sp := uint32(cpu.SP)
		res := hl + sp
		cpu.setFlag(flagCY, res > 0xFFFF)
		cpu.setHL(uint16(res))

	// logical AND
	case 0xA0, 0xA1, 0xA2, 0xA3, 0xA4, 0xA5, 0xA6, 0xA7: // ANA r
		var val uint8
		switch opcode & 0x07 {
		case 0: val = cpu.B
		case 1: val = cpu.C
		case 2: val = cpu.D
		case 3: val = cpu.E
		case 4: val = cpu.H
		case 5: val = cpu.L
		case 6: val = cpu.readM()
		case 7: val = cpu.A
		}
		cpu.A &= val
		cpu.setFlag(flagCY, false)
		cpu.setFlag(flagAC, true) // this is weird
		cpu.setZSP(cpu.A)
	case 0xE6: // ANI D8
		cpu.A &= cpu.fetch()
		cpu.setFlag(flagCY, false)
		cpu.setFlag(flagAC, true)
		cpu.setZSP(cpu.A)

	// logical XOR
	case 0xA8, 0xA9, 0xAA, 0xAB, 0xAC, 0xAD, 0xAE, 0xAF: // XRA r
		var val uint8
		switch opcode & 0x07 {
		case 0: val = cpu.B
		case 1: val = cpu.C
		case 2: val = cpu.D
		case 3: val = cpu.E
		case 4: val = cpu.H
		case 5: val = cpu.L
		case 6: val = cpu.readM()
		case 7: val = cpu.A
		}
		cpu.A ^= val
		cpu.setFlag(flagCY, false)
		cpu.setFlag(flagAC, false)
		cpu.setZSP(cpu.A)
	case 0xEE: // XRI D8
		cpu.A ^= cpu.fetch()
		cpu.setFlag(flagCY, false)
		cpu.setFlag(flagAC, false)
		cpu.setZSP(cpu.A)

	// logical OR
	case 0xB0, 0xB1, 0xB2, 0xB3, 0xB4, 0xB5, 0xB6, 0xB7: // ORA r
		var val uint8
		switch opcode & 0x07 {
		case 0: val = cpu.B
		case 1: val = cpu.C
		case 2: val = cpu.D
		case 3: val = cpu.E
		case 4: val = cpu.H
		case 5: val = cpu.L
		case 6: val = cpu.readM()
		case 7: val = cpu.A
		}
		cpu.A |= val
		cpu.setFlag(flagCY, false)
		cpu.setFlag(flagAC, false)
		cpu.setZSP(cpu.A)
	case 0xF6: // ORI D8
		cpu.A |= cpu.fetch()
		cpu.setFlag(flagCY, false)
		cpu.setFlag(flagAC, false)
		cpu.setZSP(cpu.A)

	// compare
	case 0xB8, 0xB9, 0xBA, 0xBB, 0xBC, 0xBD, 0xBE, 0xBF: // CMP r
		var val uint8
		switch opcode & 0x07 {
		case 0: val = cpu.B
		case 1: val = cpu.C
		case 2: val = cpu.D
		case 3: val = cpu.E
		case 4: val = cpu.H
		case 5: val = cpu.L
		case 6: val = cpu.readM()
		case 7: val = cpu.A
		}
		res16 := uint16(cpu.A) - uint16(val)
		res8 := uint8(res16)
		cpu.setFlag(flagCY, res16 > 0xFF) // Or val > cpu.A
		cpu.setFlag(flagAC, (cpu.A&0x0F) < (val&0x0F))
		cpu.setZSP(res8)
	case 0xFE: // CPI D8
		val := cpu.fetch()
		res16 := uint16(cpu.A) - uint16(val)
		res8 := uint8(res16)
		cpu.setFlag(flagCY, res16 > 0xFF)
		cpu.setFlag(flagAC, (cpu.A&0x0F) < (val&0x0F))
		cpu.setZSP(res8)

	// rotate
	case 0x07: // RLC - Rotate Left
		carry := (cpu.A & 0x80) >> 7
		cpu.A = (cpu.A << 1) | carry
		cpu.setFlag(flagCY, carry == 1)
	case 0x0F: // RRC - Rotate Right
		carry := cpu.A & 0x01
		cpu.A = (cpu.A >> 1) | (carry << 7)
		cpu.setFlag(flagCY, carry == 1)
	case 0x17: // RAL - Rotate Left through Carry
		carry := uint8(0)
		if cpu.getFlag(flagCY) {
			carry = 1
		}
		newCarry := (cpu.A & 0x80) >> 7
		cpu.A = (cpu.A << 1) | carry
		cpu.setFlag(flagCY, newCarry == 1)
	case 0x1F: // RAR - Rotate Right through Carry
		carry := uint8(0)
		if cpu.getFlag(flagCY) {
			carry = 1
		}
		newCarry := cpu.A & 0x01
		cpu.A = (cpu.A >> 1) | (carry << 7)
		cpu.setFlag(flagCY, newCarry == 1)

	// jump from bridge
	case 0xC3: // JMP adr
		cpu.PC = cpu.fetchWord()
	case 0xC2: // JNZ adr
		addr := cpu.fetchWord()
		if !cpu.getFlag(flagZ) {
			cpu.PC = addr
		}
	case 0xCA: // JZ adr
		addr := cpu.fetchWord()
		if cpu.getFlag(flagZ) {
			cpu.PC = addr
		}
	case 0xD2: // JNC adr
		addr := cpu.fetchWord()
		if !cpu.getFlag(flagCY) {
			cpu.PC = addr
		}
	case 0xDA: // JC adr
		addr := cpu.fetchWord()
		if cpu.getFlag(flagCY) {
			cpu.PC = addr
		}
	case 0xE2: // JPO adr
		addr := cpu.fetchWord()
		if !cpu.getFlag(flagP) {
			cpu.PC = addr
		}
	case 0xEA: // JPE adr
		addr := cpu.fetchWord()
		if cpu.getFlag(flagP) {
			cpu.PC = addr
		}
	case 0xF2: // JP adr
		addr := cpu.fetchWord()
		if !cpu.getFlag(flagS) {
			cpu.PC = addr
		}
	case 0xFA: // JM adr
		addr := cpu.fetchWord()
		if cpu.getFlag(flagS) {
			cpu.PC = addr
		}
	case 0xE9: // PCHL
		cpu.PC = cpu.getHL()

	// call
	case 0xCD: // CALL adr
		addr := cpu.fetchWord()
		cpu.push(cpu.PC)
		cpu.PC = addr
	case 0xC4: // CNZ adr
		addr := cpu.fetchWord()
		if !cpu.getFlag(flagZ) {
			cpu.push(cpu.PC)
			cpu.PC = addr
		}
	case 0xCC: // CZ adr
		addr := cpu.fetchWord()
		if cpu.getFlag(flagZ) {
			cpu.push(cpu.PC)
			cpu.PC = addr
		}
	case 0xD4: // CNC adr
		addr := cpu.fetchWord()
		if !cpu.getFlag(flagCY) {
			cpu.push(cpu.PC)
			cpu.PC = addr
		}
	case 0xDC: // CC adr
		addr := cpu.fetchWord()
		if cpu.getFlag(flagCY) {
			cpu.push(cpu.PC)
			cpu.PC = addr
		}
	case 0xE4: // CPO adr
		addr := cpu.fetchWord()
		if !cpu.getFlag(flagP) {
			cpu.push(cpu.PC)
			cpu.PC = addr
		}
	case 0xEC: // CPE adr
		addr := cpu.fetchWord()
		if cpu.getFlag(flagP) {
			cpu.push(cpu.PC)
			cpu.PC = addr
		}
	case 0xF4: // CP adr
		addr := cpu.fetchWord()
		if !cpu.getFlag(flagS) {
			cpu.push(cpu.PC)
			cpu.PC = addr
		}
	case 0xFC: // CM adr
		addr := cpu.fetchWord()
		if cpu.getFlag(flagS) {
			cpu.push(cpu.PC)
			cpu.PC = addr
		}

	// reuturn
	case 0xC9: // RET
		cpu.PC = cpu.pop()
	case 0xC0: // RNZ
		if !cpu.getFlag(flagZ) {
			cpu.PC = cpu.pop()
		}
	case 0xC8: // RZ
		if cpu.getFlag(flagZ) {
			cpu.PC = cpu.pop()
		}
	case 0xD0: // RNC
		if !cpu.getFlag(flagCY) {
			cpu.PC = cpu.pop()
		}
	case 0xD8: // RC
		if cpu.getFlag(flagCY) {
			cpu.PC = cpu.pop()
		}
	case 0xE0: // RPO
		if !cpu.getFlag(flagP) {
			cpu.PC = cpu.pop()
		}
	case 0xE8: // RPE
		if cpu.getFlag(flagP) {
			cpu.PC = cpu.pop()
		}
	case 0xF0: // RP
		if !cpu.getFlag(flagS) {
			cpu.PC = cpu.pop()
		}
	case 0xF8: // RM
		if cpu.getFlag(flagS) {
			cpu.PC = cpu.pop()
		}

	// retards, i mean restarts
	case 0xC7, 0xCF, 0xD7, 0xDF, 0xE7, 0xEF, 0xF7, 0xFF: // RST n
		cpu.push(cpu.PC)
		cpu.PC = uint16(opcode & 0x38) // n * 8

	// stack
	case 0xC1: // POP B
		cpu.setBC(cpu.pop())
	case 0xD1: // POP D
		cpu.setDE(cpu.pop())
	case 0xE1: // POP H
		cpu.setHL(cpu.pop())
	case 0xF1: // POP PSW
		cpu.setPSW(cpu.pop())
	case 0xC5: // PUSH B
		cpu.push(cpu.getBC())
	case 0xD5: // PUSH D
		cpu.push(cpu.getDE())
	case 0xE5: // PUSH H
		cpu.push(cpu.getHL())
	case 0xF5: // PUSH PSW
		cpu.push(cpu.getPSW())

	// i/o
	case 0xD3: // OUT D8
		// port := cpu.fetch()
		// fmt.Printf("OUT to port %02X\n", port)
		_ = cpu.fetch() // consume the byte but do nothing
	case 0xDB: // IN D8
		// port := cpu.fetch()
		// fmt.Printf("IN from port %02X\n", port)
		_ = cpu.fetch() // consume the byte but do nothing
		cpu.A = 0       // return 0 for now

	// special instructions
	case 0x27: // DAA - Decimal Adjust Accumulator FUCK THIS SHIT I HATE DAA
		lowNibble := cpu.A & 0x0F
		ac := cpu.getFlag(flagAC)
		cy := cpu.getFlag(flagCY)
		correction := uint8(0)

		if ac || lowNibble > 9 {
			correction += 0x06
		}

		highNibble := (cpu.A >> 4) + (correction >> 4) // include carry from low nibble correction
		if cy || highNibble > 9 {
			correction += 0x60
		}

		res16 := uint16(cpu.A) + uint16(correction)
		cpu.A += correction
		cpu.setFlag(flagCY, cy || (res16 > 0xFF))
		cpu.setZSP(cpu.A)
		cpu.setFlag(flagAC, ((cpu.A&0x0F) < (correction&0x0F)))

	case 0x2F: // CMA - complement accumulator
		cpu.A = ^cpu.A
	case 0x37: // STC - Set Carry
		cpu.setFlag(flagCY, true)
	case 0x3F: // CMC - complement carry
		cpu.setFlag(flagCY, !cpu.getFlag(flagCY))
	case 0xEB: // XCHG - exchange DE and HL
		cpu.D, cpu.H = cpu.H, cpu.D
		cpu.E, cpu.L = cpu.L, cpu.E
	case 0xE3: // XTHL - exchange top of stack with HL
		l := cpu.Memory[cpu.SP]
		h := cpu.Memory[cpu.SP+1]
		cpu.Memory[cpu.SP], cpu.L = cpu.L, l
		cpu.Memory[cpu.SP+1], cpu.H = cpu.H, h
	case 0xF9: // SPHL - Move HL to SP
		cpu.SP = cpu.getHL()
	case 0xF3: // DI - Disable Interrupts
		cpu.InterruptsEnabled = false
	case 0xFB: // EI - Enable Interrupts
		cpu.InterruptsEnabled = true

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
		Memory:            make([]uint8, *memSize),
		PC:                0,
		Flags:             0x02, // bit 1 is always set to 1 on the 8080
		InterruptsEnabled: true,
	}

	// SP is set by progam but lets just put it at the top
	cpu.SP = uint16(*memSize)

	if *romPath != "" {
		cpu.LoadROM(*romPath)
	} else {
		fmt.Println("No ROM specified...")
		
	}
	
	// not cycle accurate
	// write it urself for accuracy
	cycleDuration := time.Duration(1e9/(*mhz*1e6)) * time.Nanosecond
	ticker := time.NewTicker(cycleDuration)
	defer ticker.Stop()
	
	// Main emulation loop
	for !cpu.Halted {
		// real implementation would check for interrupts here
		cpu.step()
		<-ticker.C
	}

	fmt.Println("CPU halted.")
}
