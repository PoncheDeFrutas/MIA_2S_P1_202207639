package structures

import (
	"os"
)

func CreateBinaryFile(sizeInBytes int, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			panic(err)
		}
	}(file)

	err = writeToFile(file, sizeInBytes)
	if err != nil {
		return err
	}

	return nil
}

func writeToFile(file *os.File, sizeInBytes int) error {
	buffer := make([]byte, 1024*1024)

	for sizeInBytes > 0 {
		writeSize := len(buffer)

		if sizeInBytes < len(buffer) {
			writeSize = sizeInBytes
		}

		if _, err := file.Write(buffer[:writeSize]); err != nil {
			return err
		}

		sizeInBytes -= writeSize
	}

	return nil
}

func DeleteBinaryFile(path string) error {
	err := os.Remove(path)
	if err != nil {
		return err
	}

	return nil
}
