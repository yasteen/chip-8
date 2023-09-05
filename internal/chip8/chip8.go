package chip8

import (
	"fmt"
	"os"
	"time"
)

const c_STACK_LEN = 16
const c_MEM_LEN = 4 * 1024
const c_NUM_REG = 16
const c_FONT_INDEX = uint16(0x50)

type stackArray [c_STACK_LEN]uint16
type stack struct {
	data     stackArray
	stackPtr uint8
}

type memoryArray [c_MEM_LEN]uint8
type registerArray [c_NUM_REG]uint8

type Config struct {
	ShiftUseVY         bool
	JumpWeird          bool
	AddToIndexOverflow bool
	MemChangeIndex     bool
}

type chip8 struct {
	Memory     memoryArray
	Pc         uint16
	Index      uint16
	Stack      stack
	DelayTimer uint8
	SoundTimer uint8
	V          registerArray
	Rl         rlWrapper
	Config     Config
}

func DefaultConfig() Config {
	return Config{
		ShiftUseVY:         true,  // true : original; use ?
		JumpWeird:          false, // false: original; use false
		AddToIndexOverflow: false, // false: original; use true?
		MemChangeIndex:     false, // true : original; use false
	}
}

func NewChip8(r rlWrapper) chip8 {
	c := chip8{Rl: r}
	c.reset()
	return c
}

func (c *chip8) reset() {
	c.Memory = memoryArray{}
	c.Pc = 0
	c.Index = 0
	c.Stack = stack{data: stackArray{}, stackPtr: c_STACK_LEN - 1}
	c.DelayTimer = 0
	c.SoundTimer = 0
	c.V = registerArray{}
	c.Rl = NewRlWrapper(c.Rl.scale)
}

func (c *chip8) Run(filename string, ips uint32) {
	defer c.Rl.end()
	program, err := readRom(filename)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	copy(c.Memory[c_FONT_INDEX:c_FONT_INDEX+0x50], v_FONT[:])
	copy(c.Memory[0x200:], program)
	c.Pc = 0x200

	for true {
		instruction := c.fetch()
		if !c.execute(instruction) {
			panic(fmt.Sprintf("Command failed; 0x%04x at 0x%04x", instruction, c.Pc))
		}

		if c.SoundTimer != 0 {
			c.playSound()
		} else {
			c.stopSound()
		}

		// c.updateWindow()
		time.Sleep(time.Second / time.Duration(ips))
	}
}

func (c *chip8) DecrementTimers() {
	for true {
		if c.DelayTimer != 0 {
			c.DelayTimer--
		}
		if c.SoundTimer != 0 {
			c.SoundTimer--
		}
		time.Sleep(time.Second / 60)
	}
}

func (c *chip8) fetch() uint16 {
	i := uint16(c.Memory[c.Pc])<<8 + uint16(c.Memory[c.Pc+1])
	return i
}

func (c *chip8) execute(i uint16) bool {
	if !isValidInstruction(i) {
		return false
	}
	opcode := i >> 12
	// fmt.Printf("PC: 0x%04x; I: 0x%04x; op: 0x%01x\n", c.Pc, i, opcode)
	c.Pc += 2

	switch opcode {
	case 0:
		if i == 0x00E0 {
			c.i_clearScreen()
		} else {
			c.i_ret()
		}
	case 1:
		c.i_jump(i)
	case 2:
		c.i_subroutine(i)
	case 3:
		c.i_skipIfEqualNN(i)
	case 4:
		c.i_skipIfNotEqualNN(i)
	case 5:
		c.i_skipIfEqualReg(i)
	case 9:
		c.i_skipIfNotEqualReg(i)
	case 6:
		c.i_setRegister(i)
	case 7:
		c.i_addToRegister(i)
	case 8:
		c.i_aluOp(i)
	case 0xA:
		c.i_setIndexRegister(i)
	case 0xB:
		c.i_jumpWithOffset(i)
	case 0xC:
		c.i_randAndNum(i)
	case 0xD:
		c.i_display(i)
	case 0xE:
		c.i_skipIfKeyDown(i)
	case 0xF:
		x := uint8((i >> 8) & 0xF)
		cmd := uint8(i & 0xFF)
		if cmd == 0x07 || cmd == 0x15 || cmd == 0x18 {
			c.i_timer(x, cmd)
		} else if cmd == 0x1E {
			c.i_addToIndex(x)
		} else if cmd == 0x0A {
			c.i_getKey(x)
		} else if cmd == 0x29 {
			c.i_mvIndexToFontChar(x)
		} else if cmd == 0x33 {
			c.i_binaryCodedDecimalConversion(x)
		} else if cmd == 0x55 {
			c.i_storeMem(x)
		} else if cmd == 0x65 {
			c.i_loadMem(x)
		}
	default:
		panic("ERROR: SHOULD NOT REACH")
	}
	return true
}

func isValidInstruction(i uint16) bool {
	opcode := i >> 12

	// AND mask
	iF00F := i & 0xF00F
	iF0FF := i & 0xF0FF

	if (opcode >= 1 && opcode <= 4) || opcode == 6 || opcode == 7 || (opcode >= 0xA && opcode <= 0xD) {
		return true
	}

	return i == 0x00E0 || i == 0x00EE || // opcode 0
		iF00F == 0x5000 || // opcode 5
		(opcode == 8 && (i&0xF <= 0x7 || i&0xF == 0xE)) || // opcode 8
		iF00F == 0x9000 || // opcode 9
		iF0FF == 0xE09E || iF0FF == 0xE0A1 || // opcode E
		iF0FF == 0xF007 || iF0FF == 0xF015 || iF0FF == 0xF018 || iF0FF == 0xF01E || iF0FF == 0xF00A || iF0FF == 0xF029 || iF0FF == 0xF033 || iF0FF == 0xF055 || iF0FF == 0xF065 // opcode F
}

func (s *stack) push(addr uint16) {
	s.data[s.stackPtr] = addr
	s.stackPtr--
}

func (s *stack) pop() uint16 {
	s.stackPtr++
	return s.data[s.stackPtr]
}
