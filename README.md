# chip-8

A CHIP-8 emulator

Run this program with a specified CHIP-8 file
```go run ./cmd/chip-8 [FILENAME]```

There are unfortunately multiple differences in implementations over the
years. They can be configured by editing `emulator.Config.[QUIRK_NAME]`
in `cmd/chip-8/main.go`.
