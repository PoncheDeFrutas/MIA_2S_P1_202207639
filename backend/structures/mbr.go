package structures

import (
	"backend/common"
	"encoding/binary"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
)

type MBR struct {
	MbrSize          [4]byte      //Indicates the size of the Disk
	MbrCreationDate  [4]byte      //Indicates the date and time when the disk was created
	MbrDiskSignature [4]byte      //Disk signature
	DskFit           byte         //Indicates the way the disk is going to be formatted
	MbrPartition     [4]Partition //Array of partitions
	// Total size: 153 bytes
}

func (mbr *MBR) CreateMBR(size int, fit string, path string) error {
	creationDate, err := common.GetCreationDate(path)
	if err != nil {
		return err
	}

	binary.LittleEndian.PutUint32(mbr.MbrSize[:], uint32(size))
	copy(mbr.MbrCreationDate[:], creationDate[:])
	binary.LittleEndian.PutUint32(mbr.MbrDiskSignature[:], uint32(rand.Int31()))
	mbr.DskFit = fit[0]

	for i := range mbr.MbrPartition {
		mbr.MbrPartition[i].DefaultValue()
		copy(mbr.MbrPartition[i].PartCorrelative[:], strconv.Itoa(i))
	}

	return nil
}

func (mbr *MBR) String() string {
	size := binary.LittleEndian.Uint32(mbr.MbrSize[:])
	creationDate := strings.TrimSpace(common.ReadDate(mbr.MbrCreationDate))
	diskSignature := binary.LittleEndian.Uint32(mbr.MbrDiskSignature[:])

	return "Size: " + fmt.Sprint(size) + "\n" +
		"Creation Date: " + creationDate + "\n" +
		"Disk Signature: " + fmt.Sprint(diskSignature) + "\n" +
		"Fit: " + string(mbr.DskFit) + "\n" +
		"Partitions: \n" + mbr.PartitionsToString()
}

func (mbr *MBR) PartitionsToString() string {
	var partitions string
	for i, partition := range mbr.MbrPartition {
		partitions += fmt.Sprintf("Partition %d\n%s", i+1, partition.String())
	}
	return partitions
}
