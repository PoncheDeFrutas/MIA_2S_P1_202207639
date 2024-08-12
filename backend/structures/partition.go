package structures

import "fmt"

type Partition struct {
	PartStatus      byte     //Indicates if the partition is active or not
	PartType        byte     //Indicates the type of partition
	PartFit         byte     //Indicates the way the partition is going to be formatted
	PartStart       int32    //Indicates the start of the partition in bytes
	PartSize        int32    //Indicates the size of the partition in bytes
	PartName        [16]byte //Name of the partition
	PartCorrelative int32    //Correlative of the partition
	PartId          int32    //Id of the partition
	// Total size: 35 bytes
}

func (p *Partition) DefaultValue() {
	p.PartStatus = '0'
	p.PartType = 'P'
	p.PartFit = 'W'
	p.PartStart = -1
	p.PartSize = -1
	copy(p.PartName[:], "$")
	p.PartCorrelative = -1
	p.PartId = -1
}

func (p *Partition) Print() {
	fmt.Println("PartStatus: ", string(p.PartStatus))
	fmt.Println("PartType: ", string(p.PartType))
	fmt.Println("PartFit: ", string(p.PartFit))
	fmt.Println("PartStart: ", p.PartStart)
	fmt.Println("PartSize: ", p.PartSize)
	fmt.Println("PartName: ", string(p.PartName[:]))
	fmt.Println("PartCorrelative: ", p.PartCorrelative)
	fmt.Println("PartId: ", p.PartId)
}
