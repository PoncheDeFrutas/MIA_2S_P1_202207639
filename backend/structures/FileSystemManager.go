package structures

import (
	"fmt"
	"strings"
	"time"
)

func (sb *SuperBlock) GetFile(path string, index int32, filePath []string) string {
	inode := &Inode{}
	if err := inode.ReadInode(path, int64(sb.SInodeStart+index*sb.SInodeSize)); err != nil {
		return ""
	}

	if inode.IType == '0' {
		for i := int32(0); i < 12 && inode.IBlock[i] != -1; i++ {
			inodeIndex := sb.GetIndexInode(path, inode.IBlock[i], filePath[0])

			if inodeIndex != -1 {
				response := sb.GetFile(path, inodeIndex, filePath[1:])
				if response != "" {
					return response
				}
			}
		}
	} else if inode.IType == '1' {
		return sb.getFileContent(path, inode)
	}

	return ""
}

func (sb *SuperBlock) getFileContent(path string, inode *Inode) string {
	var content strings.Builder
	for i := int32(0); i < 12; i++ {
		if inode.IBlock[i] == -1 {
			continue
		}
		content.WriteString(sb.GetContentBlock(path, inode.IBlock[i]))
	}
	return content.String()
}

func (sb *SuperBlock) GetIndexInode(path string, index int32, file string) int32 {
	block := &FolderBlock{}
	if err := block.ReadFolderBlock(path, int64(sb.SBlockStart+index*sb.SBlockSize)); err != nil {
		return -1
	}

	for i := 2; i < len(block.BContent); i++ {
		name := strings.TrimRight(string(block.BContent[i].BName[:]), "\x00")
		if name == file {
			return block.BContent[i].BInode
		}
	}
	return -1
}

func (sb *SuperBlock) GetContentBlock(path string, index int32) string {
	block := &FileBlock{}
	if err := block.ReadFileBlock(path, int64(sb.SBlockStart+index*sb.SBlockSize)); err != nil {
		return ""
	}

	return strings.TrimRight(string(block.BContent[:]), "\x00")
}

func (sb *SuperBlock) WriteFile(path string, index int32, filePath []string, content string) (int, error) {
	inode := &Inode{}
	if err := inode.ReadInode(path, int64(sb.SInodeStart+index*sb.SInodeSize)); err != nil {
		return 0, err
	}

	if inode.IType == '0' {
		for i := int32(0); i < 12; i++ {
			if inode.IBlock[i] == -1 {
				continue
			}
			inodeIndex := sb.GetIndexInode(path, inode.IBlock[i], filePath[0])

			if inodeIndex != -1 {
				writtenBytes, err := sb.WriteFile(path, inodeIndex, filePath[1:], content)
				if err != nil {
					return 0, err
				}

				if writtenBytes > 0 {
					return writtenBytes, err
				}
			}
		}
	} else if inode.IType == '1' {
		return sb.writeFileContent(path, inode, content, index)
	}

	return 0, nil
}

func (sb *SuperBlock) writeFileContent(path string, inode *Inode, content string, index int32) (int, error) {
	contentBytes := []byte(content)
	contentLength := len(contentBytes)
	remainingContent := contentBytes
	totalWritten := 0

	for i := int32(0); i < 12 && len(remainingContent) > 0; i++ {
		blockIndex := inode.IBlock[i]

		var fileBlock FileBlock

		bytesToWrite := mini(64, contentLength)
		copy(fileBlock.BContent[:], remainingContent[:bytesToWrite])

		if blockIndex == -1 {
			if err := fileBlock.WriteFileBlock(path, int64(sb.SFirstBlo), int64(sb.SFirstBlo+sb.SBlockSize)); err != nil {
				return 0, err
			}

			inode.IBlock[i] = sb.SBlocksCount

			if err := sb.UpdateBitmapBlock(path); err != nil {
				return 0, err
			}

		} else {
			if err := fileBlock.WriteFileBlock(path, int64(sb.SBlockStart+blockIndex*sb.SBlockSize), int64(sb.SBlockStart+(blockIndex+1)*sb.SBlockSize)); err != nil {
				return 0, err
			}
		}
		inode.IMTime = float32(time.Now().Unix())
		if err := inode.WriteInode(path, int64(sb.SInodeStart+index*sb.SInodeSize), int64(sb.SInodeStart+(index+1)*sb.SInodeSize)); err != nil {
			return 0, err
		}

		remainingContent = remainingContent[bytesToWrite:]
		contentLength = len(remainingContent)
		totalWritten += bytesToWrite
	}

	if len(remainingContent) > 0 {
		// TODO add a new pointer block
	}

	return totalWritten, nil
}

func (sb *SuperBlock) CreateInode(path string, index int32, dirPath []string, createParents, isFile bool) error {
	inode := &Inode{}
	if err := inode.ReadInode(path, int64(sb.SInodeStart+index*sb.SInodeSize)); err != nil {
		return err
	}

	if inode.IType != '0' {
		return fmt.Errorf("invalid inode type: %c", inode.IType)
	}

	if len(dirPath) == 1 {
		return sb.addInodeToParent(path, inode, dirPath[0], index, isFile)
	}

	for _, blockIndex := range inode.IBlock {
		if blockIndex == -1 {
			continue
		}
		inodeIndex := sb.GetIndexInode(path, blockIndex, dirPath[0])
		if inodeIndex != -1 {
			return sb.CreateInode(path, inodeIndex, dirPath[1:], createParents, isFile)
		}
	}

	if createParents {
		if err := sb.addInodeToParent(path, inode, dirPath[0], index, false); err != nil {
			return err
		}
		newIndexNode := sb.GetIndexInode(path, inode.IBlock[0], dirPath[0])
		return sb.CreateInode(path, newIndexNode, dirPath[1:], createParents, isFile)
	}

	return fmt.Errorf("directory not found: %s", dirPath[0])
}

func (sb *SuperBlock) addInodeToParent(path string, parentInode *Inode, name string, parentIndex int32, isFile bool) error {
	for _, blockIndex := range parentInode.IBlock {
		if blockIndex == -1 {
			continue
		}

		block := &FolderBlock{}
		if err := block.ReadFolderBlock(path, int64(sb.SBlockStart+blockIndex*sb.SBlockSize)); err != nil {
			return err
		}

		for j := 2; j < len(block.BContent); j++ {
			if block.BContent[j].BInode != -1 {
				continue
			}

			copy(block.BContent[j].BName[:], name)
			block.BContent[j].BInode = sb.SInodesCount

			newInode := &Inode{}
			newInode.DefaultValue(-1)
			newInode.IType = '1' // Archivo
			if !isFile {
				newInode.IType = '0' // Directorio
			}
			newInode.IPerm = [3]byte{'6', '6', '4'}

			if err := newInode.WriteInode(path, int64(sb.SFirstIno), int64(sb.SFirstIno+sb.SInodeSize)); err != nil {
				return err
			}

			if err := sb.UpdateBitmapInode(path); err != nil {
				return err
			}

			if err := block.WriteFolderBlock(path, int64(sb.SBlockStart+blockIndex*sb.SBlockSize),
				int64(sb.SBlockStart+(blockIndex+1)*sb.SBlockSize)); err != nil {
				return err
			}

			return nil
		}
	}

	return sb.createNewInode(path, parentInode, name, parentIndex, sb.findFreeBlockIndex(parentInode), isFile)
}

func (sb *SuperBlock) createNewInode(path string, parentInode *Inode, name string, parentIndex, blockIndex int32, isFile bool) error {
	newInode := &Inode{}
	newInode.DefaultValue(-1)
	newInode.IType = '1' // Archivo
	if !isFile {
		newInode.IType = '0' // Directorio
	}
	newInode.IPerm = [3]byte{'6', '6', '4'}

	block := &FolderBlock{}
	block.DefaultValue()
	block.BContent[0].BInode = sb.SInodesCount
	block.BContent[1].BInode = parentIndex

	parentInode.IBlock[blockIndex] = sb.SBlocksCount
	parentInode.ISize += sb.SInodeSize

	for i := int32(2); i < int32(len(block.BContent)); i++ {
		if block.BContent[i].BInode != -1 {
			continue
		}

		copy(block.BContent[i].BName[:], name)
		block.BContent[i].BInode = sb.SInodesCount

		if err := block.WriteFolderBlock(path, int64(sb.SFirstBlo), int64(sb.SFirstBlo+sb.SBlockSize)); err != nil {
			return err
		}
		break
	}

	if err := parentInode.WriteInode(path, int64(sb.SInodeStart+parentIndex*sb.SInodeSize),
		int64(sb.SInodeStart+(parentIndex+1)*sb.SInodeSize)); err != nil {
		return err
	}

	if err := newInode.WriteInode(path, int64(sb.SFirstIno),
		int64(sb.SFirstIno+sb.SInodeSize)); err != nil {
		return err
	}

	if err := sb.UpdateBitmapInode(path); err != nil {
		return err
	}
	if err := sb.UpdateBitmapBlock(path); err != nil {
		return err
	}

	return nil
}

func (sb *SuperBlock) findFreeBlockIndex(inode *Inode) int32 {
	for i, block := range inode.IBlock {
		if block == -1 {
			return int32(i)
		}
	}
	return -1
}

func mini(a, b int) int {
	if a < b {
		return a
	}
	return b
}
