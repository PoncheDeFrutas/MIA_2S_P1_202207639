package structures

import (
	"backend/common"
	"math/rand"
	"strings"
)

type MBR struct {
	MbrSize          int32
	MbrCreationDate  [4]byte
	MbrDiskSignature int32
	MbrDiskFit       byte
	MbrPartitions    [4]Partition
	// Total size: 153 bytes
}

func (m *MBR) CreateMBR(size int, fit string, path string) error {
	creationDate, err := common.GetCreationDate(path)
	if err != nil {
		return err
	}

	m.MbrSize = int32(size)
	copy(m.MbrCreationDate[:], creationDate[:])
	m.MbrDiskSignature = rand.Int31()
	m.MbrDiskFit = fit[0]

	for i := range m.MbrPartitions {
		m.MbrPartitions[i].DefaultValue()
		m.MbrPartitions[i].PartCorrelative = int32(i + 1)
	}

	if err := common.WriteToFile(path, int64(0), m); err != nil {
		return err
	}

	return nil
}

func (m *MBR) Print() {
	println("MBR")
	println("MbrSize: ", m.MbrSize)
	println("MbrCreationDate: ", strings.TrimSpace(common.ReadDate(m.MbrCreationDate)))
	println("MbrDiskSignature: ", m.MbrDiskSignature)
	println("MbrDiskFit: ", rune(m.MbrDiskFit))
	println("Partitions")
	for i := range m.MbrPartitions {
		println("------------------------------------------------")
		m.MbrPartitions[i].Print()
	}
}
