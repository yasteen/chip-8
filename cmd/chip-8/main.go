package main

import (
	"chip-8/internal/chip8"
	"fmt"
	"os"
)

const scale = 10
const fps = 60
const ips = 600

func main() {
	if len(os.Args) == 1 {
		fmt.Fprintf(os.Stderr, "Usage: chip8 [FILENAME]")
		return
	}
	filename := os.Args[1]

	emulator := chip8.NewChip8(chip8.NewRlWrapper(scale))

	emulator.Rl.Setup(fps)
	go emulator.Run(filename, ips)
	go emulator.DecrementTimers()
	emulator.StartWindowDraw()
}
