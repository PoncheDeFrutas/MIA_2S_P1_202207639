package structures

import "backend/utils"

type FolderBlock struct {
	BContent [4]FolderContent
	// Total size of the FolderBlock is 64 bytes
}

type FolderContent struct {
	BName  [12]byte
	BInode int32
	// Total size of the FolderContent is 16 bytes
}

func (f *FolderBlock) DefaultValue() {
	f.BContent = [4]FolderContent{
		{BName: [12]byte{'.'}, BInode: 0},
		{BName: [12]byte{'.', '.'}, BInode: 0},
		{BName: [12]byte{'-'}, BInode: -1},
		{BName: [12]byte{'-'}, BInode: -1},
	}
}

func (f *FolderBlock) WriteFolderBlock(path string, offset int64, maxSize int64) error {
	if err := utils.WriteToFile(path, offset, maxSize, f); err != nil {
		return err
	}
	return nil
}

func (f *FolderBlock) ReadFolderBlock(path string, offset int64) error {
	if err := utils.ReadFromFile(path, offset, f); err != nil {
		return err
	}
	return nil
}

func (f *FolderBlock) Print() {
	for i, content := range f.BContent {
		if content.BInode == -1 {
			break
		}
		content.Print(i)
	}
}

func (f *FolderContent) Print(i int) {
	println("FolderContent", i)
	println("BName", string(f.BName[:]))
	println("BInode", f.BInode)
	println()
}
