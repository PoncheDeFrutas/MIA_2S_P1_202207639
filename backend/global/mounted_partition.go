package global

import (
	"backend/structures"
	"errors"
)

const Carnet string = "39"

var (
	MountedPartitions map[string]string = make(map[string]string)
)

func GetMountedPartition(id string) (*structures.Partition, string, error) {
	path := MountedPartitions[id]
	if path == "" {
		return nil, "", errors.New("partition not mounted")
	}

	mbr := &structures.MBR{}

	if err := mbr.ReadMBR(path); err != nil {
		return nil, "", err
	}

	partition, err := mbr.GetPartitionByID(id)
	if err != nil {
		return nil, "", err
	}

	return partition, path, nil
}
