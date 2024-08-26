package structures

import (
	"backend/utils"
	"fmt"
)

type EBR struct {
	PartMount byte
	PartFit   byte
	PartStart int32
	PartSize  int32
	PartNext  int32
	PartName  [16]byte
	// Total size of the EBR is 30 bytes
}

func (e *EBR) DefaultValue() {
	e.PartMount = '9'
	e.PartFit = 'W'
	e.PartStart = -1
	e.PartSize = -1
	e.PartNext = -1
	copy(e.PartName[:], "EBR-LOGIC")
}

func (e *EBR) SetEBR(fit string, start int32, size int32, next int32, name string) {
	e.PartMount = '0'
	e.PartFit = fit[0]
	e.PartStart = start
	e.PartSize = size
	e.PartNext = next
	copy(e.PartName[:], name)
}

func (e *EBR) WriteEBR(path string, offset int64, maxSize int64) error {
	if err := utils.WriteToFile(path, offset, maxSize, e); err != nil {
		return err
	}
	return nil
}

func (e *EBR) ReadEBR(path string, offset int64) error {
	if err := utils.ReadFromFile(path, offset, e); err != nil {
		return err
	}
	return nil
}

func (e *EBR) Print() {
	fmt.Println("PartMount: ", string(e.PartMount))
	fmt.Println("PartFit: ", string(e.PartFit))
	fmt.Println("PartStart: ", e.PartStart)
	fmt.Println("PartSize: ", e.PartSize)
	fmt.Println("PartNext: ", e.PartNext)
	fmt.Println("PartName: ", string(e.PartName[:]))
}
