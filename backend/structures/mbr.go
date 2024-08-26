package structures

import (
	common "backend/utils"
	"encoding/binary"
	"fmt"
	"math/rand"
	"strings"
)

type MBR struct {
	MbrSize          int32
	MbrCreationDate  [4]byte
	MbrDiskSignature int32
	MbrDiskFit       byte
	MbrPartition     [4]Partition
	// Total size of the MBR is 153 bytes
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

	for i := range m.MbrPartition {
		m.MbrPartition[i].DefaultValue()
	}

	return nil
}

func (m *MBR) WriteMBR(path string) error {
	if err := common.WriteToFile(path, int64(0), int64(binary.Size(m)), m); err != nil {
		return err
	}
	return nil
}

func (m *MBR) ReadMBR(path string) error {
	if err := common.ReadFromFile(path, int64(0), m); err != nil {
		return err
	}
	return nil
}

func (m *MBR) FindFreePartition() int {
	for i, partition := range m.MbrPartition {
		if partition.PartStart == -1 {
			return i
		}
	}
	return -1
}

func (m *MBR) FreeNamePartition(name string) bool {
	for _, partition := range m.MbrPartition {
		if strings.TrimRight(string(partition.PartName[:]), "\x00") == name {
			return false
		}
	}
	return true
}

func (m *MBR) GetPartitionByName(name string) (*Partition, int) {
	for i, partition := range m.MbrPartition {
		if strings.TrimRight(string(partition.PartName[:]), "\x00") == name {
			return &m.MbrPartition[i], i
		}
	}
	return nil, -1
}

func (m *MBR) GetPartitionByID(id string) (*Partition, error) {
	for i, partition := range m.MbrPartition {
		if strings.TrimRight(string(partition.PartName[:]), "\x00") == id {
			return &m.MbrPartition[i], nil
		}
	}
	return nil, fmt.Errorf("partition not found")
}

func (m *MBR) ExtendPartitionExist() bool {
	for _, partition := range m.MbrPartition {
		if partition.PartType == 'E' {
			return true
		}
	}
	return false
}

func (m *MBR) GetExtendedPartition() *Partition {
	for i, partition := range m.MbrPartition {
		if partition.PartType == 'E' {
			return &m.MbrPartition[i]
		}
	}
	return nil
}

func (m *MBR) Print() {
	fmt.Println("/*********************** MBR ***********************/")
	fmt.Printf("Size: %d\n", m.MbrSize)
	fmt.Printf("Creation Date: %s\n", strings.TrimSpace(common.ReadDate(m.MbrCreationDate)))
	fmt.Printf("Disk Signature: %d\n", m.MbrDiskSignature)
	fmt.Printf("Disk Fit: %c\n", m.MbrDiskFit)
	fmt.Println("/******************** Partitions ********************/")
	for i, partition := range m.MbrPartition {
		fmt.Println("--------------------------------------------------")
		fmt.Printf("Partition %d\n", i)
		partition.Print()
	}
	fmt.Println("--------------------------------------------------")
}
