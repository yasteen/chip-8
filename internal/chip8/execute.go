package chip8

import (
	"fmt"
	"math/rand"
	"os"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// 00E0
func (c *chip8) i_clearScreen() {
	c.Rl.display = DisplayArray{}
}

// 1NNN
func (c *chip8) i_jump(i uint16) {
	c.Pc = i & 0x0FFF
}

// 2NNN
func (c *chip8) i_subroutine(i uint16) {
	c.Stack.push(c.Pc)
	c.Pc = i & 0x0FFF
}

// 00EE
func (c *chip8) i_ret() {
	c.Pc = c.Stack.pop()
}

// 3XNN
func (c *chip8) i_skipIfEqualNN(i uint16) {
	x := c.V[(i>>8)&0xF]
	nn := uint8(i & 0xFF)
	if x == nn {
		c.Pc += 2
	}
}

// 4XNN
func (c *chip8) i_skipIfNotEqualNN(i uint16) {
	x := c.V[(i>>8)&0xF]
	nn := uint8(i & 0xFF)
	if x != nn {
		c.Pc += 2
	}
}

// 5XY0
func (c *chip8) i_skipIfEqualReg(i uint16) {
	x := c.V[(i>>8)&0xF]
	y := c.V[(i>>4)&0xF]
	if x == y {
		c.Pc += 2
	}
}

// 9XY0
func (c *chip8) i_skipIfNotEqualReg(i uint16) {
	x := c.V[(i>>8)&0xF]
	y := c.V[(i>>4)&0xF]
	if x != y {
		c.Pc += 2
	}
}

// 6XNN
func (c *chip8) i_setRegister(i uint16) {
	x := (i >> 8) & 0xF
	c.V[x] = uint8(i & 0xFF)
}

// 7XNN
func (c *chip8) i_addToRegister(i uint16) {
	x := (i >> 8) & 0xF
	c.V[x] += uint8(i & 0xFF)
}

// 8XYI
func (c *chip8) i_aluOp(i uint16) {
	xReg := (i >> 8) & 0xF
	yReg := (i >> 4) & 0xF
	x := c.V[xReg]
	y := c.V[yReg]
	operation := i & 0xF
	switch operation {
	case 0:
		c.V[xReg] = y
	case 1:
		c.V[xReg] |= y
	case 2:
		c.V[xReg] &= y
	case 3:
		c.V[xReg] ^= y
	case 4:
		if uint16(x)+uint16(y) > 0xFF {
			c.V[0xF] = 1
		} else {
			c.V[0xF] = 0
		}
		c.V[xReg] += y
	case 5:
		if x > y {
			c.V[0xF] = 1
		} else {
			c.V[0xF] = 0
		}
		c.V[xReg] = x - y
	case 7:
		if y > x {
			c.V[0xF] = 1
		} else {
			c.V[0xF] = 0
		}
		c.V[xReg] = y - x
	case 6:
		if c.Config.ShiftUseVY {
			c.V[xReg] = y
		}
		c.V[0xF] = c.V[xReg] & 0x1
		c.V[xReg] = c.V[xReg] >> 1
	case 0xE:
		if c.Config.ShiftUseVY {
			c.V[xReg] = y
		}
		c.V[0xF] = (c.V[xReg] >> 7) & 0x1
		c.V[xReg] = c.V[xReg] << 1
	}
}

// ANNN
func (c *chip8) i_setIndexRegister(i uint16) {
	c.Index = i & 0x0FFF
}

// BNNN
func (c *chip8) i_jumpWithOffset(i uint16) {
	nnn := i & 0x0FFF
	if c.Config.JumpWeird {
		nnn += uint16(c.V[nnn>>8])
	} else {
		nnn += uint16(c.V[0])
	}
	c.Pc = nnn
}

// CXNN
func (c *chip8) i_randAndNum(i uint16) {
	x := (i >> 8) & 0xF
	nn := uint8(i & 0xFF)
	c.V[x] = uint8(rand.Uint32()) & nn
}

// DXYN
func (c *chip8) i_display(i uint16) {
	xReg := uint8((i >> 8) & 0xF)
	yReg := uint8(i >> 4 & 0xF)
	x := c.V[xReg] % c_SCREEN_WIDTH_PIXEL
	y := c.V[yReg] % c_SCREEN_HEIGHT_PIXEL
	n := uint8(i & 0xF)
	overflow := false
	for j := uint8(0); j < n; j++ {
		if y+j >= c_SCREEN_HEIGHT_PIXEL {
			break
		}
		spriteData := c.Memory[c.Index+uint16(j)]
		rowOverflow := c.Rl.drawRow(spriteData, x, y+j)
		overflow = rowOverflow || overflow
	}
	if overflow {
		c.V[0xF] = 1
	} else {
		c.V[0xF] = 0
	}
}

// EX9E, EXA1
func (c *chip8) i_skipIfKeyDown(i uint16) {
	x := c.V[(i>>8)&0xF]
	if x > 0xF {
		fmt.Printf("Invalid keycode 0x%02x should be less than 0xF", x)
		os.Exit(1)
	}

	isKeyPressed := rl.IsKeyDown(v_IND_TO_KEY[x])
	trueIfDown := i&0xFF == 0x9E

	if isKeyPressed == trueIfDown {
		c.Pc += 2
	}
}

// FX07, FX15, FX18
func (c *chip8) i_timer(x uint8, cmd uint8) {
	if cmd == 0x07 {
		c.V[x] = c.DelayTimer
	} else if cmd == 0x15 {
		c.DelayTimer = c.V[x]
	} else {
		c.SoundTimer = c.V[x]
	}
}

// FX1E
func (c *chip8) i_addToIndex(x uint8) {
	c.Index += uint16(c.V[x])
	if c.Config.AddToIndexOverflow {
		if c.Index > 0xFFF {
			c.V[0xF] = 1
		} else {
			c.V[0xF] = 0
		}
	}
}

// FX0A
func (c *chip8) i_getKey(x uint8) {
	latestKey := rl.GetKeyPressed()
	if code, exist := v_KEY_TO_IND[int(latestKey)]; exist && rl.IsKeyDown(latestKey) {
		c.V[x] = code
		return
	}

	c.Pc -= 2
}

// FX29
func (c *chip8) i_mvIndexToFontChar(x uint8) {
	valX := c.V[x]
	if valX > 0xF {
		fmt.Printf("Font data does not exist for 0x%02x; should be less than 0xF", x)
		os.Exit(1)
	}

	c.Index = c_FONT_INDEX + uint16(5*valX)
}

// FX33
func (c *chip8) i_binaryCodedDecimalConversion(x uint8) {
	xVal := c.V[x]
	c.Memory[c.Index+0] = (xVal / 100) % 10
	c.Memory[c.Index+1] = (xVal / 10) % 10
	c.Memory[c.Index+2] = xVal % 10
}

// FX55
func (c *chip8) i_storeMem(x uint8) {
	for i := uint8(0); i <= x; i++ {
		c.Memory[c.Index+uint16(i)] = c.V[i]
	}
	if c.Config.MemChangeIndex {
		c.Index += uint16(x + 1)
	}
}

// FX65
func (c *chip8) i_loadMem(x uint8) {
	for i := uint8(0); i <= x; i++ {
		c.V[i] = c.Memory[c.Index+uint16(i)]
	}
	if c.Config.MemChangeIndex {
		c.Index += uint16(x + 1)
	}
}
