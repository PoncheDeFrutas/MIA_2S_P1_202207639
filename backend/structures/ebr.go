package structures

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
