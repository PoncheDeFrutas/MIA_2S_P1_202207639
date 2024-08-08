package common

import (
	"encoding/binary"
	"fmt"
	"os"
)

func WriteToFile(path string, offset int64, data interface{}) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			panic(err)
		}
	}(file)

	if _, err = file.Seek(offset, 0); err != nil {
		return fmt.Errorf("failed to seek file: %v", err)
	}

	if err = binary.Write(file, binary.LittleEndian, data); err != nil {
		return fmt.Errorf("failed to write to file: %v", err)
	}

	return nil
}

func ReadFromFile(path string, offset int64, data interface{}) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			panic(err)
		}
	}(file)

	if _, err = file.Seek(offset, 0); err != nil {
		return fmt.Errorf("failed to seek file: %v", err)
	}

	if err = binary.Read(file, binary.LittleEndian, data); err != nil {
		return fmt.Errorf("failed to read from file: %v", err)
	}

	return nil
}
