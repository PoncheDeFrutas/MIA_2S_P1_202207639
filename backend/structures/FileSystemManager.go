package structures

import (
	"fmt"
	"strings"
	"time"
)

func (sb *SuperBlock) GetFile(path string, index int32, filePath []string) string {
	inode := &Inode{}
	inodePath := int64(sb.SInodeStart + index*sb.SInodeSize)

	if err := inode.ReadInode(path, inodePath); err != nil {
		return ""
	}

	if inode.IType == '0' {
		inodeIndex := sb.findInodeInBlock(path, filePath[0], inode)
		if inodeIndex != -1 {
			return sb.GetFile(path, inodeIndex, filePath[1:])
		}
	} else if inode.IType == '1' {
		return sb.getFileContent(path, inode)
	}

	return "Error path not found"
}

func (sb *SuperBlock) getFileContent(path string, inode *Inode) string {
	var content strings.Builder

	for _, block := range inode.IBlock[:12] {
		if block == -1 {
			continue
		}
		content.WriteString(sb.getContentBlock(path, block))
	}

	if inode.IBlock[12] != -1 {
		content.WriteString(sb.getIndirectBlockContent(path, inode.IBlock[12], 1))
	}

	if inode.IBlock[13] != -1 {
		content.WriteString(sb.getIndirectBlockContent(path, inode.IBlock[13], 2))
	}

	if inode.IBlock[14] != -1 {
		content.WriteString(sb.getIndirectBlockContent(path, inode.IBlock[14], 3))
	}

	return content.String()
}

func (sb *SuperBlock) getContentBlock(path string, index int32) string {
	block := &FileBlock{}
	blockPath := int64(sb.SBlockStart + index*sb.SBlockSize)

	if err := block.ReadFileBlock(path, blockPath); err != nil {
		return ""
	}

	return strings.TrimRight(string(block.BContent[:]), "\x00")
}

func (sb *SuperBlock) getIndirectBlockContent(path string, blockIndex int32, level int) string {
	if level == 0 {
		return sb.getContentBlock(path, blockIndex)
	}

	block := &PointerBlock{}
	blockPath := int64(sb.SBlockStart + blockIndex*sb.SBlockSize)

	if err := block.ReadPointerBlock(path, blockPath); err != nil {
		return ""
	}

	var content strings.Builder
	for _, subBlockIndex := range block.PPointers {
		if subBlockIndex == -1 {
			continue
		}
		content.WriteString(sb.getIndirectBlockContent(path, subBlockIndex, level-1))
	}

	return content.String()
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
			inodeIndex := sb.GetIndexInode(path, filePath[0], inode.IBlock[i])

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

func (sb *SuperBlock) CreatePath(path string, index int32, filePath []string, root, isFile bool) error {
	inode := &Inode{}
	inodePath := int64(sb.SInodeStart + index*sb.SInodeSize)
	if err := inode.ReadInode(path, inodePath); err != nil {
		return err
	}

	if inode.IType != '0' {
		return fmt.Errorf("invalid inode type: %c", inode.IType)
	}

	if len(filePath) == 1 {
		return sb.addPathToFolder(path, filePath[0], inode, isFile, index)
	}

	indexInode := sb.findInodeInBlock(path, filePath[0], inode)

	if indexInode != -1 {
		return sb.CreatePath(path, indexInode, filePath[1:], root, isFile)
	}

	if root {
		if err := sb.addPathToFolder(path, filePath[0], inode, false, index); err != nil {
			return err
		}
		indexInode = sb.findInodeInBlock(path, filePath[0], inode)
		return sb.CreatePath(path, indexInode, filePath[1:], root, isFile)
	}

	return fmt.Errorf("path not found")
}

func (sb *SuperBlock) addPathToFolder(path, name string, inode *Inode, isFile bool, indexInode int32) error {
	for _, blockIndex := range inode.IBlock[:12] {
		if blockIndex == -1 {
			continue
		}
		condition, err := sb.addContentToFolderBlock(path, name, blockIndex)
		if err != nil {
			return err
		}

		if !condition {
			continue
		}

		if _, err := sb.CreateInode(path, isFile); err != nil {
			return err
		}
		return nil
	}

	for i, blockIndex := range inode.IBlock[12:] {
		if blockIndex == -1 {
			continue
		}
		condition, err := sb.addContentToPointerBlock(path, name, blockIndex, int32(i+1))
		if err != nil {
			return err
		}

		if !condition {
			continue
		}

		if _, err := sb.CreateInode(path, isFile); err != nil {
			return err
		}
		return nil
	}

	return sb.CreateNewBlock(path, name, isFile, indexInode, inode)
}

func (sb *SuperBlock) addContentToFolderBlock(path, name string, blockIndex int32) (bool, error) {
	block := &FolderBlock{}
	if err := block.ReadFolderBlock(path, int64(sb.SBlockStart+blockIndex*sb.SBlockSize)); err != nil {
		return false, err
	}

	if block.BContent[3].BInode != -1 {
		return false, nil
	}

	copy(block.BContent[3].BName[:], name)
	block.BContent[3].BInode = sb.SInodesCount

	if err := block.WriteFolderBlock(path, int64(sb.SBlockStart+blockIndex*sb.SBlockSize), int64(sb.SBlockStart+(blockIndex+1)*sb.SBlockSize)); err != nil {
		return false, err
	}

	return true, nil
}

func (sb *SuperBlock) addContentToPointerBlock(path, name string, blockIndex, level int32) (bool, error) {
	pointerBlock := &PointerBlock{}
	if err := pointerBlock.ReadPointerBlock(path, int64(sb.SBlockStart+blockIndex*sb.SBlockSize)); err != nil {
		return false, err
	}

	for _, pointer := range pointerBlock.PPointers {
		if pointer == -1 {
			continue
		}

		if level == 1 {
			condition, err := sb.addContentToFolderBlock(path, name, pointer)
			if err != nil {
				return false, err
			}
			if condition {
				return true, nil
			}
		} else if level == 2 {
			condition, err := sb.addContentToPointerBlock(path, name, pointer, level-1)
			if err != nil {
				return false, err
			}
			if condition {
				return true, nil
			}
		} else if level == 3 {
			condition, err := sb.addContentToPointerBlock(path, name, pointer, level-1)
			if err != nil {
				return false, err
			}
			if condition {
				return true, nil
			}
		}
	}

	freePointerIndex := pointerBlock.FindFreePointer()

	if freePointerIndex == -1 {
		return false, nil
	}

	if level == 1 {
		pointerBlock.PPointers[freePointerIndex] = sb.SBlocksCount
		if err := pointerBlock.WritePointerBlock(path, int64(sb.SBlockStart+blockIndex*sb.SBlockSize),
			int64(sb.SBlockStart+(blockIndex+1)*sb.SBlockSize)); err != nil {
			return false, err
		}

		if err := sb.CreateDirectoryBlock(path, name, sb.SInodesCount); err != nil {
			return false, err
		}
		return true, nil
	} else if level == 2 {
		pointerBlock.PPointers[freePointerIndex] = sb.SBlocksCount
		if err := pointerBlock.WritePointerBlock(path, int64(sb.SBlockStart+blockIndex*sb.SBlockSize),
			int64(sb.SBlockStart+(blockIndex+1)*sb.SBlockSize)); err != nil {
			return false, err
		}

		if err := sb.createPointerBlock(path, level-1); err != nil {
			return false, err
		}

		condition, err := sb.addContentToPointerBlock(path, name, sb.SBlocksCount, level-1)
		if err != nil {
			return false, err
		}

		if condition {
			return true, nil
		}

	} else if level == 3 {
		// TODO IN FOLDER 1078
		// hasta cuando unicamente ya no haya espacio disponible en 304 crear nuevo bloque doble
	}

	return false, nil
}

func (sb *SuperBlock) CreateInode(path string, isFile bool) (*Inode, error) {
	newInode := &Inode{}
	newInode.DefaultValue(-1)
	if isFile {
		newInode.IType = '1' // Archivo
	} else {
		newInode.IType = '0' // Directorio
	}
	newInode.IPerm = [3]byte{'6', '6', '4'}

	if err := newInode.WriteInode(path, int64(sb.SFirstIno), int64(sb.SFirstIno+sb.SInodeSize)); err != nil {
		return nil, err
	}

	if err := sb.UpdateBitmapInode(path); err != nil {
		return nil, err
	}
	return newInode, nil
}

func (sb *SuperBlock) CreateDirectoryBlock(path, name string, indexInode int32) error {
	block := &FolderBlock{}
	block.DefaultValue()
	block.BContent[0].BInode = sb.SInodesCount // TODO FIX THIS
	block.BContent[1].BInode = indexInode      // TODO FIX THIS

	block.BContent[2].BInode = sb.SInodesCount
	copy(block.BContent[2].BName[:], name)

	if err := block.WriteFolderBlock(path, int64(sb.SFirstBlo), int64(sb.SFirstBlo+sb.SBlockSize)); err != nil {
		return err
	}
	if err := sb.UpdateBitmapBlock(path); err != nil {
		return err
	}
	//sb.BlocksCount - 1
	return nil
}

func (sb *SuperBlock) CreateNewBlock(path, name string, isFile bool, indexInode int32, inode *Inode) error {
	freeBlockIndex := sb.findFreeBlockIndex(inode)

	if int(freeBlockIndex) >= len(inode.IBlock) {
		return fmt.Errorf("no more free blocks available for inode")
	}

	if freeBlockIndex >= 0 && freeBlockIndex <= 11 {
		inode.IBlock[freeBlockIndex] = sb.SBlocksCount
		inode.IMTime = float32(time.Now().Unix())
	}

	inodeStart := int64(sb.SInodeStart + indexInode*sb.SInodeSize)
	inodeEnd := int64(sb.SInodeStart + (indexInode+1)*sb.SInodeSize)

	if freeBlockIndex > 11 {
		if err := sb.CreateNewPointerBlock(path, freeBlockIndex, inode); err != nil {
			return fmt.Errorf("failed to create pointer block: %w", err)
		}
	}

	if err := inode.WriteInode(path, inodeStart, inodeEnd); err != nil {
		return fmt.Errorf("failed to write inode: %w", err)
	}

	if err := sb.CreateDirectoryBlock(path, name, indexInode); err != nil {
		return fmt.Errorf("failed to create directory block: %w", err)
	}

	if _, err := sb.CreateInode(path, isFile); err != nil {
		return fmt.Errorf("failed to create inode: %w", err)
	}

	return nil
}

func (sb *SuperBlock) CreateNewPointerBlock(path string, freeBlockIndex int32, inode *Inode) error {
	if freeBlockIndex == 12 {
		inode.IBlock[freeBlockIndex] = sb.SBlocksCount
		if err := sb.createPointerBlock(path, freeBlockIndex-11); err != nil {
			return fmt.Errorf("failed to create pointer block: %w", err)
		}
	} else if freeBlockIndex == 13 {
		if sb.findFreePointerIndex(path, inode.IBlock[freeBlockIndex-1]) != -1 {
			return nil
		}

		inode.IBlock[freeBlockIndex] = sb.SBlocksCount
		if err := sb.createPointerBlock(path, freeBlockIndex-11); err != nil {
			return fmt.Errorf("failed to create pointer block: %w", err)
		}
	} else if freeBlockIndex == 14 {
		if sb.findFreePointerInPointer(path, inode.IBlock[freeBlockIndex-1]) != -1 { //cambiar a cualquiera dentro
			return nil
		}

		inode.IBlock[freeBlockIndex] = sb.SBlocksCount
		if err := sb.createPointerBlock(path, freeBlockIndex-11); err != nil {
			return fmt.Errorf("failed to create pointer block: %w", err)
		}
	}

	return nil
}

func (sb *SuperBlock) createPointerBlock(path string, level int32) error {
	pointerBlock := &PointerBlock{}
	pointerBlock.DefaultValue()
	pointerBlock.PPointers[0] = sb.SBlocksCount + 1

	if err := pointerBlock.WritePointerBlock(path, int64(sb.SFirstBlo), int64(sb.SFirstBlo+sb.SBlockSize)); err != nil {
		return err
	}

	if err := sb.UpdateBitmapBlock(path); err != nil {
		return err
	}

	if level > 1 {
		return sb.createPointerBlock(path, level-1)
	}

	return nil
}

func (sb *SuperBlock) findInodeInBlock(path, part string, inode *Inode) int32 {
	for _, block := range inode.IBlock[:12] {
		if block == -1 {
			continue
		}

		blockIndex := sb.GetIndexInode(path, part, block)
		if blockIndex != -1 {
			return blockIndex
		}
	}

	for i, block := range inode.IBlock[12:] {
		if block == -1 {
			continue
		}

		blockIndex := sb.findInodeInPointerBlock(path, part, block, int32(i+1))
		if blockIndex != -1 {
			return blockIndex
		}
	}

	return -1
}

func (sb *SuperBlock) findInodeInPointerBlock(path, part string, blockIndex, level int32) int32 {
	if level < 0 {
		return -1
	}

	pointerBlock := &PointerBlock{}
	if err := pointerBlock.ReadPointerBlock(path, int64(sb.SBlockStart+blockIndex*sb.SBlockSize)); err != nil {
		return -1
	}

	for _, pointer := range pointerBlock.PPointers {
		if pointer == -1 {
			continue
		}
		if level == 0 {
			inodeIndex := sb.GetIndexInode(path, part, pointer)
			if inodeIndex != -1 {
				return inodeIndex
			}
			continue
		}
		inodeIndex := sb.findInodeInPointerBlock(path, part, pointer, level-1)
		if inodeIndex != -1 {
			return inodeIndex
		}
	}
	return -1
}

func (sb *SuperBlock) findFreeBlockIndex(inode *Inode) int32 {
	for i, block := range inode.IBlock {
		if block == -1 {
			return int32(i)
		}
	}
	return -1
}

func (sb *SuperBlock) findFreePointerIndex(path string, indexBlock int32) int32 {
	pointerBlock := &PointerBlock{}
	if err := pointerBlock.ReadPointerBlock(path, int64(sb.SBlockStart+indexBlock*sb.SBlockSize)); err != nil {
		return -1
	}
	return pointerBlock.FindFreePointer()
}

func (sb *SuperBlock) findFreePointerInPointer(path string, indexBlock int32) int32 {
	pointerBlock := &PointerBlock{}
	if err := pointerBlock.ReadPointerBlock(path, int64(sb.SBlockStart+indexBlock*sb.SBlockSize)); err != nil {
		return -1
	}

	if pointerBlock.FindFreePointer() != -1 {
		return pointerBlock.FindFreePointer()
	}

	for _, pointer := range pointerBlock.PPointers {
		if pointer == -1 {
			continue
		}
		if sb.findFreePointerIndex(path, pointer) != -1 {
			return pointer
		}
	}
	return -1
}

func (sb *SuperBlock) GetIndexInode(path, file string, index int32) int32 {
	block := &FolderBlock{}
	blockPath := int64(sb.SBlockStart + index*sb.SBlockSize)

	if err := block.ReadFolderBlock(path, blockPath); err != nil {
		return -1
	}

	for _, entry := range block.BContent[2:] {
		name := strings.TrimRight(string(entry.BName[:]), "\x00")
		if name == file {
			return entry.BInode
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
