package commands

import (
	"backend/common"
	"backend/manager"
	"backend/structures"
	"encoding/binary"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type FDisk struct {
	Size int
	Unit string
	Path string
	Type string
	Fit  string
	Name string
}

func ParserFDisk(tokens []string) (string, error) {
	cmd := &FDisk{}

	args := strings.Join(tokens, " ")
	re := regexp.MustCompile(`-size=\d+|-unit=[bBkKmM]|-fit=[bBfF]{2}|-path="[^"]+"|-path=\S+|-type=[pPeElL]|-name="[^"]+"|-name=\S+`)
	matches := re.FindAllString(args, -1)

	for _, match := range matches {
		key, value, err := common.ParseToken(match)
		if err != nil {
			return "", err
		}

		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = strings.Trim(value, "\"")
		}

		switch key {
		case "-size":
			size, err := strconv.Atoi(value)
			if err != nil || size < 1 {
				return "", fmt.Errorf("invalid size: %s", value)
			}
			cmd.Size = size
		case "-unit":
			value = strings.ToUpper(value)
			if value != "B" && value != "K" && value != "M" {
				return "", fmt.Errorf("invalid unit: %s", value)
			}
			cmd.Unit = value
		case "-path":
			if value == "" {
				return "", fmt.Errorf("invalid path: %s", value)
			}
			cmd.Path = value
		case "-type":
			value = strings.ToUpper(value)
			if value != "P" && value != "E" && value != "L" {
				return "", fmt.Errorf("invalid type: %s", value)
			}
			cmd.Type = value
		case "-fit":
			value = strings.ToUpper(value)
			if value != "BF" && value != "FF" && value != "WF" {
				return "", fmt.Errorf("invalid fit: %s", value)
			}
			cmd.Fit = value
		case "-name":
			if value == "" {
				return "", fmt.Errorf("invalid name: %s", value)
			}
			cmd.Name = value
		default:
			return "", fmt.Errorf("unknown option: %s", key)
		}
	}

	if cmd.Size == 0 {
		return "", fmt.Errorf("missing size")
	}

	if cmd.Unit == "" {
		cmd.Unit = "K"
	}

	if cmd.Path == "" {
		return "", fmt.Errorf("missing path")
	}

	if cmd.Type == "" {
		cmd.Type = "P"
	}

	if cmd.Fit == "" {
		cmd.Fit = "WF"
	}

	if cmd.Name == "" {
		return "", fmt.Errorf("missing name")
	}

	if err := cmd.commandFDisk(); err != nil {
		return "", err
	}

	return "", nil
}

func (cmd *FDisk) commandFDisk() error {
	sizeInBytes, err := common.ConvertToBytes(cmd.Size, cmd.Unit)
	if err != nil {
		return err
	}

	mbr := &structures.MBR{}
	if err := mbr.ReadMBR(cmd.Path); err != nil {
		return err
	}

	if cmd.Type != "P" {
		if err := cmd.validatePartitionType(mbr); err != nil {
			return err
		}
	}

	return cmd.allocatePartition(mbr, sizeInBytes)
}

func (cmd *FDisk) validatePartitionType(mbr *structures.MBR) error {
	extendExist := mbr.ExtendPartitionExist()

	if (extendExist && cmd.Type == "E") || (!extendExist && cmd.Type == "L") {
		return fmt.Errorf("invalid partition type: %s", cmd.Type)
	}

	return nil
}

func (cmd *FDisk) allocatePartition(mbr *structures.MBR, size int) error {
	indexPart, indexByte, err := cmd.findAvailableSpace(mbr, size)
	if err != nil {
		return err
	}

	return cmd.createPartitionEntry(mbr, indexPart, indexByte, size)
}

func (cmd *FDisk) findAvailableSpace(mbr *structures.MBR, size int) (int, int32, error) {
	objects := make([]interface{}, len(mbr.MbrPartitions))
	for i, partition := range mbr.MbrPartitions {
		objects[i] = partition
	}

	indexByte := manager.FirstFit(objects, int32(size), int32(153), mbr.MbrSize)
	indexPart := mbr.FindFreePartition()

	if indexByte == -1 || indexPart == -1 {
		return -1, -1, fmt.Errorf("no space available")
	}

	return indexPart, indexByte, nil
}

func (cmd *FDisk) createPartitionEntry(mbr *structures.MBR, indexPart int, indexByte int32, size int) error {
	partition := &mbr.MbrPartitions[indexPart]

	partition.PartType = cmd.Type[0]
	partition.PartFit = cmd.Fit[0]
	partition.PartStart = indexByte
	partition.PartSize = int32(size)

	if !mbr.FreeNamePartition(cmd.Name) {
		return fmt.Errorf("partition name already exists")
	}

	copy(partition.PartName[:], cmd.Name)

	if partition.PartType != 'L' {
		if err := common.WriteToFile(cmd.Path, 0, int64(binary.Size(mbr)), mbr); err != nil {
			return err
		}
	}

	if err := cmd.handlePartitionType(partition, mbr); err != nil {
		return err
	}

	if err := common.ReadFromFile(cmd.Path, 0, mbr); err != nil {
		return err
	}
	return nil
}

func (cmd *FDisk) handlePartitionType(partition *structures.Partition, mbr *structures.MBR) error {
	switch cmd.Type {
	case "P":
		return cmd.createPrimaryPartition(partition)
	case "E":
		return cmd.createExtendedPartition(partition)
	case "L":
		return cmd.createLogicalPartition(partition, mbr)
	default:
		return fmt.Errorf("unknown partition type")
	}
}

func (cmd *FDisk) createPrimaryPartition(partition *structures.Partition) error {
	// Lógica para crear partición primaria
	return nil
}

func (cmd *FDisk) createExtendedPartition(partition *structures.Partition) error {
	ebr := &structures.EBR{}
	ebr.DefaultValue()
	if err := common.WriteToFile(cmd.Path, int64(partition.PartStart), int64(partition.PartStart+partition.PartSize), ebr); err != nil {
		return err
	}
	return nil
}

func (cmd *FDisk) createLogicalPartition(partition *structures.Partition, mbr *structures.MBR) error {
	extPartition := &structures.Partition{}

	for i := range mbr.MbrPartitions {
		if mbr.MbrPartitions[i].PartType == 'E' {
			extPartition = &mbr.MbrPartitions[i]
			partition.PartStart = extPartition.PartStart
			break
		}
	}

	ebr := &structures.EBR{}

	if err := common.ReadFromFile(cmd.Path, int64(extPartition.PartStart), ebr); err != nil {
		return err
	}

	for ebr.PartNext != -1 {
		partition.PartStart = ebr.PartNext
		if err := common.ReadFromFile(cmd.Path, int64(ebr.PartNext), ebr); err != nil {
			return err
		}
	}

	ebr.PartFit = partition.PartFit
	ebr.PartStart = partition.PartStart + 30
	ebr.PartSize = partition.PartSize
	ebr.PartNext = ebr.PartSize + ebr.PartStart
	copy(ebr.PartName[:], partition.PartName[:])

	if err := common.WriteToFile(cmd.Path, int64(ebr.PartStart-30), int64(extPartition.PartStart+extPartition.PartSize), ebr); err != nil {
		return err
	}

	ebrDefault := &structures.EBR{}
	ebrDefault.DefaultValue()

	if err := common.WriteToFile(cmd.Path, int64(ebr.PartNext), int64(extPartition.PartStart+extPartition.PartSize), ebrDefault); err != nil {
		return err
	}

	if err := common.ReadFromFile(cmd.Path, int64(ebr.PartStart-30), ebr); err != nil {
		return err
	}
	return nil
}
