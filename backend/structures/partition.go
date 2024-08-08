package structures

type Partition struct {
	PartStatus      byte     //Indicates if the partition is active or not
	PartType        byte     //Indicates the type of partition
	PartFit         byte     //Indicates the way the partition is going to be formatted
	PartStart       [4]byte  //Indicates the start of the partition in bytes
	PartSize        [4]byte  //Indicates the size of the partition in bytes
	PartName        [16]byte //Name of the partition
	PartCorrelative [4]byte  //Correlative of the partition
	PartId          [4]byte  //Id of the partition
	// Total size: 35 bytes
}

func (part *Partition) DefaultValue() {
	part.PartStatus = '$'
	part.PartType = 'P'
	part.PartFit = 'W'
	copy(part.PartStart[:], "$")
	copy(part.PartSize[:], "$")
	copy(part.PartName[:], "$")
	copy(part.PartCorrelative[:], "$")
	copy(part.PartId[3:], "#")
}

func (part *Partition) String() string {

	return "Status: " + string(part.PartStatus) + "\n" +
		"Type: " + string(part.PartType) + "\n" +
		"Fit: " + string(part.PartFit) + "\n" +
		"Start: " + string(part.PartStart[:]) + "\n" +
		"Size: " + string(part.PartSize[:]) + "\n" +
		"Name: " + string(part.PartName[:]) + "\n" +
		"Correlative: " + string(part.PartCorrelative[:]) + "\n" +
		"Id: " + string(part.PartId[:]) + "\n"
}
