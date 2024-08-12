package structures

import (
	"backend/common"
	"encoding/binary"
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

	if err := common.WriteToFile(path, int64(0), int64(binary.Size(m)), m); err != nil {
		return err
	}

	return nil
}

func (m *MBR) FindFreePartition() int {
	for i := range m.MbrPartitions {
		if m.MbrPartitions[i].PartStart == -1 {
			return i
		}
	}
	return -1
}

func (m *MBR) FreeNamePartition(name string) bool {
	for i := range m.MbrPartitions {
		if strings.TrimRight(string(m.MbrPartitions[i].PartName[:]), "\x00") == name {
			return false
		}
	}
	return true
}

func (m *MBR) ExtendPartitionExist() bool {
	for i := range m.MbrPartitions {
		if m.MbrPartitions[i].PartType == 'E' {
			return true
		}
	}
	return false
}

func (m *MBR) ReadMBR(path string) error {
	if err := common.ReadFromFile(path, int64(0), m); err != nil {
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
