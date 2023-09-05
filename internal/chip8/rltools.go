package chip8

import (
	"fmt"
	"image/color"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const c_SCREEN_WIDTH_PIXEL = 64
const c_SCREEN_HEIGHT_PIXEL = 32

type DisplayArray [c_SCREEN_HEIGHT_PIXEL][c_SCREEN_WIDTH_PIXEL]bool

var v_FONT [16 * 5]uint8 = [16 * 5]uint8{
	0xF0, 0x90, 0x90, 0x90, 0xF0, // 0
	0x20, 0x60, 0x20, 0x20, 0x70, // 1
	0xF0, 0x10, 0xF0, 0x80, 0xF0, // 2
	0xF0, 0x10, 0xF0, 0x10, 0xF0, // 3
	0x90, 0x90, 0xF0, 0x10, 0x10, // 4
	0xF0, 0x80, 0xF0, 0x10, 0xF0, // 5
	0xF0, 0x80, 0xF0, 0x90, 0xF0, // 6
	0xF0, 0x10, 0x20, 0x40, 0x40, // 7
	0xF0, 0x90, 0xF0, 0x90, 0xF0, // 8
	0xF0, 0x90, 0xF0, 0x10, 0xF0, // 9
	0xF0, 0x90, 0xF0, 0x90, 0x90, // A
	0xE0, 0x90, 0xE0, 0x90, 0xE0, // B
	0xF0, 0x80, 0x80, 0x80, 0xF0, // C
	0xE0, 0x90, 0x90, 0x90, 0xE0, // D
	0xF0, 0x80, 0xF0, 0x80, 0xF0, // E
	0xF0, 0x80, 0xF0, 0x80, 0x80, // F
}

var v_IND_TO_KEY = [16]int32{
	rl.KeyX, rl.KeyOne, rl.KeyTwo, rl.KeyThree,
	rl.KeyQ, rl.KeyW, rl.KeyE, rl.KeyA,
	rl.KeyS, rl.KeyD, rl.KeyZ, rl.KeyC,
	rl.KeyFour, rl.KeyR, rl.KeyF, rl.KeyV,
}

var v_KEY_TO_IND = map[int]uint8{
	rl.KeyOne: 1, rl.KeyTwo: 2, rl.KeyThree: 3, rl.KeyFour: 0xC,
	rl.KeyQ: 4, rl.KeyW: 5, rl.KeyE: 6, rl.KeyR: 0xD,
	rl.KeyA: 7, rl.KeyS: 0x8, rl.KeyD: 0x9, rl.KeyF: 0xE,
	rl.KeyZ: 0xA, rl.KeyX: 0, rl.KeyC: 0xB, rl.KeyV: 0xF,
}

type rlWrapper struct {
	scale   uint8
	width   int32
	height  int32
	display DisplayArray
	beep    interface{}
}

func NewRlWrapper(scale uint8) rlWrapper {
	wrapper := rlWrapper{
		scale:   scale,
		width:   int32(scale) * c_SCREEN_WIDTH_PIXEL,
		height:  int32(scale) * c_SCREEN_HEIGHT_PIXEL,
		display: DisplayArray{},
		beep:    rl.Sound{FrameCount: 0},
	}

	return wrapper
}

// drawRow takes in sprite data and coordinates
// 0 <= x < 64
// 0 <= y < 32
// Returns true if overflow (both display and sprite data is on at some pixel)
func (r *rlWrapper) drawRow(spriteRow uint8, x uint8, y uint8) bool {
	flag := false
	for i := uint8(0); i < 8; i++ {
		if x+i >= c_SCREEN_WIDTH_PIXEL {
			break
		}
		spriteData := (spriteRow >> (7 - i) & 0x1) == 0x1
		displayData := r.display[y][x+i]
		flag = flag || (spriteData && displayData)
		r.display[y][x+i] = spriteData != displayData
	}
	return flag
}

func (c *chip8) updateWindow() {
	rl.BeginDrawing()
	rl.ClearBackground(rl.Black)
	c.drawDebugInfo()
	defer rl.EndDrawing()

	black := color.RGBA{0x0, 0x0, 0x0, 0xFF}
	white := color.RGBA{0xFF, 0xFF, 0xFF, 0xFF}
	for y := int32(0); y < c_SCREEN_HEIGHT_PIXEL; y++ {
		for x := int32(0); x < c_SCREEN_WIDTH_PIXEL; x++ {
			col := black
			if c.Rl.display[y][x] {
				col = white
			}
			rl.DrawRectangle(
				int32(c.Rl.scale)*x,
				int32(c.Rl.scale)*y,
				int32(c.Rl.scale),
				int32(c.Rl.scale),
				col,
			)
		}
	}

}

func (r *rlWrapper) Setup(fps uint32) {
	rl.SetTraceLogLevel(rl.LogNone)
	rl.InitWindow(r.width, r.height, "CHIP-8 Emulator")
	rl.SetTargetFPS(int32(fps))

	rl.InitAudioDevice()
	r.beep = rl.LoadSound("assets/a440sin.ogg")
}

func (r *rlWrapper) end() {
	if r.beep != nil {
		rl.UnloadSound(r.beep.(rl.Sound))
	}

	rl.CloseAudioDevice()
	rl.CloseWindow()
}

func (c *chip8) StartWindowDraw() {
	defer c.Rl.end()
	for !rl.WindowShouldClose() {
		c.updateWindow()
	}
}

func (c *chip8) drawDebugInfo() {
	rl.DrawText(fmt.Sprintf("FPS: %d\n", rl.GetFPS()), 5, 5, 10, rl.Red)
	rl.DrawText(fmt.Sprintf("Timer: %d\n", c.DelayTimer), 5, 20, 10, rl.Red)
}

func (c *chip8) playSound() {
	if !rl.IsSoundPlaying(c.Rl.beep.(rl.Sound)) {
		rl.PlaySound(c.Rl.beep.(rl.Sound))
	}
}

func (c *chip8) stopSound() {
	if rl.IsSoundPlaying(c.Rl.beep.(rl.Sound)) {
		rl.StopSound(c.Rl.beep.(rl.Sound))
	}
}
