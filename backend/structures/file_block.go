package structures

import (
	"backend/utils"
	"fmt"
)

type FileBlock struct {
	BContent [64]byte
	// Total size of the FileBlock is 64 bytes
}

func (f *FileBlock) WriteFileBlock(path string, offset int64, maxSize int64) error {
	if err := utils.WriteToFile(path, offset, maxSize, f); err != nil {
		return err
	}
	return nil
}

func (f *FileBlock) ReadFileBlock(path string, offset int64) error {
	if err := utils.ReadFromFile(path, offset, f); err != nil {
		return err
	}
	return nil
}

func (f *FileBlock) Print() {
	fmt.Printf("--*-- FileBlock --*--\n")
	fmt.Printf("BContent: %s\n", f.BContent)
}
