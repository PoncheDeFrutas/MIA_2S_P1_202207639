package structures

import (
	"encoding/binary"
	"fmt"
	"math"
)

type Partition struct {
	PartStatus      byte
	PartType        byte
	PartFit         byte
	PartStart       int32
	PartSize        int32
	PartName        [16]byte
	PartCorrelative int32
	PartId          [4]byte
	// Total size of the Partition is 35 bytes
}

func (p *Partition) DefaultValue() {
	p.PartStatus = '9'
	p.PartType = 'P'
	p.PartFit = 'W'
	p.PartStart = -1
	p.PartSize = -1
	copy(p.PartName[:], "$")
	p.PartCorrelative = -1
	copy(p.PartId[:], "$$$$")
}

func (p *Partition) SetPartition(partType string, fit string, start int32, size int32, name string) {
	p.PartStatus = '0'
	p.PartType = partType[0]
	p.PartFit = fit[0]
	p.PartStart = start
	p.PartSize = size
	copy(p.PartName[:], name)
}

func (p *Partition) IsEmpty() bool {
	return p.PartStart == -1 && p.PartSize == -1
}

func (p *Partition) MountPartition(correlative int, id string) error {
	p.PartCorrelative = int32(correlative) + 1
	copy(p.PartId[:], id)
	return nil
}

func (p *Partition) UnmountPartition() {
	p.PartCorrelative = -1
	copy(p.PartId[:], "$$$$")
}

func (p *Partition) IsMounted() bool {
	return p.PartCorrelative != -1
}

func (p *Partition) CalculateN() int32 {
	numerator := int(p.PartSize) - binary.Size(SuperBlock{})
	denominator := 4 + binary.Size(Inode{}) + 3*binary.Size(FileBlock{})
	return int32(math.Floor(float64(numerator) / float64(denominator)))
}

func (p *Partition) Print() {
	fmt.Println("PartStatus: ", string(p.PartStatus))
	fmt.Println("PartType: ", string(p.PartType))
	fmt.Println("PartFit: ", string(p.PartFit))
	fmt.Println("PartStart: ", p.PartStart)
	fmt.Println("PartSize: ", p.PartSize)
	fmt.Println("PartName: ", string(p.PartName[:]))
	fmt.Println("PartCorrelative: ", p.PartCorrelative)
	fmt.Println("PartId: ", string(p.PartId[:]))
}
