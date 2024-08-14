package structures

import "fmt"

type EBR struct {
	PartMount byte
	PartFit   byte
	PartStart int32
	PartSize  int32
	PartNext  int32
	PartName  [16]byte
	// Total size: 30 bytes
}

func (e *EBR) DefaultValue() {
	e.PartMount = '0'
	e.PartFit = 'W'
	e.PartStart = -1
	e.PartSize = -1
	e.PartNext = -1
	copy(e.PartName[:], "EBR-LOGICA")
}

func (e *EBR) Print() {
	fmt.Println("PartMount: ", string(e.PartMount))
	fmt.Println("PartFit: ", string(e.PartFit))
	fmt.Println("PartStart: ", e.PartStart)
	fmt.Println("PartSize: ", e.PartSize)
	fmt.Println("PartNext: ", e.PartNext)
	fmt.Println("PartName: ", string(e.PartName[:]))
}
