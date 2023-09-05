package chip8

import (
	"errors"
	"fmt"
	"os"
)

func readRom(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return []byte{}, errors.New(fmt.Sprintf("io: Unable to read file %s.", path))
	}

	return data, nil
}
