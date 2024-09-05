package commands

import (
	"backend/global"
	"backend/structures"
	"backend/utils"
	"fmt"
	"github.com/goccy/go-graphviz"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type REP struct {
	Name       string
	Path       string
	Id         string
	PathFileLs string
}

func ParserREP(tokens []string) (string, error) {
	cmd := &REP{}

	args := strings.Join(tokens, " ")
	re := regexp.MustCompile(`(?i)-name="[^"]+"|-name=\S+|-path="[^"]+"|-path=\S+|-id=\S+|-path_file_ls="[^"]+"|-path_file_ls=\S+`)
	matches := re.FindAllString(args, -1)

	for _, match := range matches {
		key, value, err := utils.ParseToken(match)
		if err != nil {
			return "", err
		}

		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = strings.Trim(value, "\"")
		}

		switch key {
		case "-name":
			if value == "" {
				return "", fmt.Errorf("invalid name: %s", value)
			}
			cmd.Name = value
		case "-path":
			if value == "" {
				return "", fmt.Errorf("invalid path: %s", value)
			}
			cmd.Path = value
		case "-id":
			if value == "" {
				return "", fmt.Errorf("invalid id: %s", value)
			}
			cmd.Id = value
		case "-path_file_ls":
			if value == "" {
				return "", fmt.Errorf("invalid path_file_ls: %s", value)
			}
			cmd.PathFileLs = value
		}
	}

	if cmd.Name == "" {
		return "", fmt.Errorf("missing name")
	}

	if cmd.Path == "" {
		return "", fmt.Errorf("missing path")
	}

	if cmd.Id == "" {
		return "", fmt.Errorf("missing id")
	}

	if cmd.PathFileLs == "" {
		cmd.PathFileLs = "disk.txt"
	}

	if err := cmd.commandREP(); err != nil {
		return "", err
	}

	return "", nil
}

func (cmd *REP) commandREP() error {
	switch cmd.Name {
	case "mbr":
		return cmd.repMBR()
	case "disk":
		//return cmd.repDisk()
	case "inode":
		return cmd.repInode()
	case "block":
		return cmd.repBlock()
	case "bm_inode":
		return cmd.repBMInode()
	case "bm_block":
		return cmd.repBMBlock()
	case "sb":
		return cmd.repSB()
	case "file":
		return cmd.repFile()
	case "ls":
		//return cmd.repLS()
	default:
		return fmt.Errorf("invalid name: %s", cmd.Name)
	}
	return nil
}

func (cmd *REP) repMBR() error {
	_, path, err := global.GetMountedPartition(cmd.Id)
	if err != nil {
		return err
	}

	mbr := &structures.MBR{}
	if err := mbr.ReadMBR(path); err != nil {
		return err
	}

	var sb strings.Builder

	// Create DOT content based on MBR object
	sb.WriteString("digraph G {\n")
	sb.WriteString("\tnode [shape=plaintext];\n")
	sb.WriteString("\tReporteMBR [label=<\n")
	sb.WriteString("\t<TABLE BORDER=\"1\" CELLBORDER=\"1\" CELLSPACING=\"0\" CELLPADDING=\"4\">\n")

	sb.WriteString(mbr.GetStringBuilder())

	// Partitions title row
	sb.WriteString(fmt.Sprintf("<TR><TD COLSPAN=\"2\" BGCOLOR=\"%s\"><B>Particiones</B></TD></TR>\n", "#AAAAAA"))

	// Partitions rows
	for i, partition := range mbr.MbrPartition {
		sb.WriteString(partition.GetStringBuilder(i))

		if partition.PartType != 'E' {
			continue
		}
		ebr := &structures.EBR{}
		if err := ebr.ReadEBR(path, int64(mbr.GetExtendedPartition().PartStart)); err != nil {
			return err
		}
		sb.WriteString(ebr.GetStringBuilder())

		for ebr.PartNext != -1 {
			if err := ebr.ReadEBR(path, int64(ebr.PartNext)); err != nil {
				return err
			}
			sb.WriteString(ebr.GetStringBuilder())
		}
	}

	sb.WriteString("    </TABLE>\n")
	sb.WriteString("    >];\n")
	sb.WriteString("}\n")

	return cmd.generateImage(sb.String())
}

func (cmd *REP) repSB() error {
	partition, path, err := global.GetMountedPartition(cmd.Id)
	if err != nil {
		return err
	}

	superBlock := &structures.SuperBlock{}
	if err := superBlock.ReadSuperBlock(path, int64(partition.PartStart)); err != nil {
		return err
	}

	var sb strings.Builder

	// Create DOT content based on SuperBlock object
	sb.WriteString("digraph G {\n")
	sb.WriteString("\tnode [shape=plaintext];\n")
	sb.WriteString("\tReporteMBR [label=<\n")
	sb.WriteString("\t<TABLE BORDER=\"1\" CELLBORDER=\"1\" CELLSPACING=\"0\" CELLPADDING=\"4\">\n")

	sb.WriteString(superBlock.GetStringBuilder())

	sb.WriteString("    </TABLE>\n")
	sb.WriteString("    >];\n")
	sb.WriteString("}\n")

	return cmd.generateImage(sb.String())
}

func (cmd *REP) repInode() error {
	partition, path, err := global.GetMountedPartition(cmd.Id)
	if err != nil {
		return err
	}

	superBlock := &structures.SuperBlock{}
	if err := superBlock.ReadSuperBlock(path, int64(partition.PartStart)); err != nil {
		return err
	}

	var sb strings.Builder

	// Create DOT content based on Inode object
	sb.WriteString("digraph G {\n")
	sb.WriteString("\tnode [shape=plaintext];\n")
	sb.WriteString("\trankdir=LR;\n")

	inode := &structures.Inode{}
	for i := int32(0); i < superBlock.SInodesCount; i++ {
		if err := inode.ReadInode(path, int64(superBlock.SInodeStart+(i*superBlock.SInodeSize))); err != nil {
			return err
		}
		sb.WriteString(inode.GetStringBuilder(fmt.Sprintf("Inodo_%d", i)))

		if i < superBlock.SInodesCount-1 {
			sb.WriteString(fmt.Sprintf("Inodo_%d -> Inodo_%d\n", i, i+1))
		}
	}

	sb.WriteString("}")
	return cmd.generateImage(sb.String())
}

func (cmd *REP) repBlock() error {
	partition, path, err := global.GetMountedPartition(cmd.Id)
	if err != nil {
		return err
	}

	superBlock := &structures.SuperBlock{}
	if err := superBlock.ReadSuperBlock(path, int64(partition.PartStart)); err != nil {
		return err
	}

	var sb strings.Builder

	// Create DOT content based on Inode object
	sb.WriteString("digraph G {\n")
	sb.WriteString("\tnode [shape=plaintext];\n")
	sb.WriteString("\trankdir=LR;\n")

	inode := &structures.Inode{}
	for i := int32(0); i < superBlock.SInodesCount; i++ {
		if err := inode.ReadInode(path, int64(superBlock.SInodeStart+(i*superBlock.SInodeSize))); err != nil {
			return err
		}

		sb.WriteString(inode.GetStringBuilder(fmt.Sprintf("Inodo_%d", i)))

		// Determine the block type and its corresponding handler
		blockType := inode.IType
		var readBlock func(path string, blockIndex int64) (string, error)

		switch blockType {
		case '0': // FolderBlock
			readBlock = func(path string, blockIndex int64) (string, error) {
				block := &structures.FolderBlock{}
				if err := block.ReadFolderBlock(path, blockIndex); err != nil {
					return "", err
				}
				return block.GetStringBuilder(fmt.Sprintf("Bloque_%d", (blockIndex-int64(superBlock.SBlockStart))/64)), nil
			}
		case '1': // FileBlock
			readBlock = func(path string, blockIndex int64) (string, error) {
				block := &structures.FileBlock{}
				if err := block.ReadFileBlock(path, blockIndex); err != nil {
					return "", err
				}
				return block.GetStringBuilder(fmt.Sprintf("Bloque_%d", (blockIndex-int64(superBlock.SBlockStart))/64)), nil
			}
		default:
			return fmt.Errorf("unknown inode type: %c", blockType)
		}

		// Process each block for the inode
		for j := 0; j < 12; j++ {
			blockIndex := inode.IBlock[j]
			if blockIndex == -1 {
				break
			}
			blockStart := int64(superBlock.SBlockStart + (blockIndex * superBlock.SBlockSize))
			blockString, err := readBlock(path, blockStart)
			if err != nil {
				return err
			}
			sb.WriteString(blockString)

			sb.WriteString(fmt.Sprintf("Inodo_%d -> Bloque_%d\n", i, blockIndex))
		}

		// Process indirect blocks
		if inode.IBlock[12] != -1 {
			indirectBlock := &structures.PointerBlock{}
			if err := indirectBlock.ReadPointerBlock(path, int64(superBlock.SBlockStart+(inode.IBlock[12]*superBlock.SBlockSize))); err != nil {
				return err
			}
			sb.WriteString(indirectBlock.GetStringBuilder(fmt.Sprintf("Bloque_%d", inode.IBlock[12])))
			sb.WriteString(fmt.Sprintf("Inodo_%d -> Bloque_%d\n", i, inode.IBlock[12]))

			for j := 0; j < 16; j++ {
				blockIndex := indirectBlock.PPointers[j]
				if blockIndex == -1 {
					break
				}
				blockStart := int64(superBlock.SBlockStart + (blockIndex * superBlock.SBlockSize))
				blockString, err := readBlock(path, blockStart)
				if err != nil {
					return err
				}
				sb.WriteString(blockString)

				sb.WriteString(fmt.Sprintf("Bloque_%d -> Bloque_%d\n", inode.IBlock[12], blockIndex))
			}
		}

		// Process double indirect blocks
		if inode.IBlock[13] != -1 {
			doubleIndirectBlock := &structures.PointerBlock{}
			if err := doubleIndirectBlock.ReadPointerBlock(path, int64(superBlock.SBlockStart+(inode.IBlock[13]*superBlock.SBlockSize))); err != nil {
				return err
			}
			sb.WriteString(doubleIndirectBlock.GetStringBuilder(fmt.Sprintf("Bloque_%d", inode.IBlock[13])))
			sb.WriteString(fmt.Sprintf("Inodo_%d -> Bloque_%d\n", i, inode.IBlock[13]))

			for j := 0; j < 16; j++ {
				indirectBlock := &structures.PointerBlock{}
				if doubleIndirectBlock.PPointers[j] == -1 {
					continue
				}

				if err := indirectBlock.ReadPointerBlock(path, int64(superBlock.SBlockStart+(doubleIndirectBlock.PPointers[j]*superBlock.SBlockSize))); err != nil {
					return err
				}
				sb.WriteString(indirectBlock.GetStringBuilder(fmt.Sprintf("Bloque_%d", doubleIndirectBlock.PPointers[j])))
				sb.WriteString(fmt.Sprintf("Bloque_%d -> Bloque_%d\n", inode.IBlock[13], doubleIndirectBlock.PPointers[j]))

				for k := 0; k < 16; k++ {
					blockIndex := indirectBlock.PPointers[k]
					if blockIndex == -1 {
						continue
					}
					blockStart := int64(superBlock.SBlockStart + (blockIndex * superBlock.SBlockSize))
					blockString, err := readBlock(path, blockStart)
					if err != nil {
						return err
					}
					sb.WriteString(blockString)

					sb.WriteString(fmt.Sprintf("Bloque_%d -> Bloque_%d\n", doubleIndirectBlock.PPointers[j], blockIndex))
				}
			}
		}

		// Process triple indirect blocks
		if inode.IBlock[14] != -1 {
			tripleIndirectBlock := &structures.PointerBlock{}
			if err := tripleIndirectBlock.ReadPointerBlock(path, int64(superBlock.SBlockStart+(inode.IBlock[14]*superBlock.SBlockSize))); err != nil {
				return err
			}
			sb.WriteString(tripleIndirectBlock.GetStringBuilder(fmt.Sprintf("Bloque_%d", inode.IBlock[14])))
			sb.WriteString(fmt.Sprintf("Inodo_%d -> Bloque_%d\n", i, inode.IBlock[14]))

			for j := 0; j < 16; j++ {
				doubleIndirectBlock := &structures.PointerBlock{}
				if tripleIndirectBlock.PPointers[j] == -1 {
					continue
				}

				if err := doubleIndirectBlock.ReadPointerBlock(path, int64(superBlock.SBlockStart+(tripleIndirectBlock.PPointers[j]*superBlock.SBlockSize))); err != nil {
					return err
				}
				sb.WriteString(doubleIndirectBlock.GetStringBuilder(fmt.Sprintf("Bloque_%d", tripleIndirectBlock.PPointers[j])))
				sb.WriteString(fmt.Sprintf("Bloque_%d -> Bloque_%d\n", inode.IBlock[14], tripleIndirectBlock.PPointers[j]))

				for k := 0; k < 16; k++ {
					indirectBlock := &structures.PointerBlock{}
					if doubleIndirectBlock.PPointers[k] == -1 {
						continue
					}

					if err := indirectBlock.ReadPointerBlock(path, int64(superBlock.SBlockStart+(doubleIndirectBlock.PPointers[k]*superBlock.SBlockSize))); err != nil {
						return err
					}

					sb.WriteString(indirectBlock.GetStringBuilder(fmt.Sprintf("Bloque_%d", doubleIndirectBlock.PPointers[k])))
					sb.WriteString(fmt.Sprintf("Bloque_%d -> Bloque_%d\n", tripleIndirectBlock.PPointers[j], doubleIndirectBlock.PPointers[k]))

					for l := 0; l < 16; l++ {
						blockIndex := indirectBlock.PPointers[l]
						if blockIndex == -1 {
							continue
						}
						blockStart := int64(superBlock.SBlockStart + (blockIndex * superBlock.SBlockSize))
						blockString, err := readBlock(path, blockStart)
						if err != nil {
							return err
						}
						sb.WriteString(blockString)

						sb.WriteString(fmt.Sprintf("Bloque_%d -> Bloque_%d\n", doubleIndirectBlock.PPointers[k], blockIndex))

					}
				}
			}
		}
	}
	sb.WriteString("}")
	return cmd.generateImage(sb.String())
}

func (cmd *REP) repBMInode() error {
	partition, path, err := global.GetMountedPartition(cmd.Id)
	if err != nil {
		return err
	}

	superBlock := &structures.SuperBlock{}
	if err := superBlock.ReadSuperBlock(path, int64(partition.PartStart)); err != nil {
		return err
	}

	text, err := utils.ReadFromBitMap(path, int64(superBlock.SBMInodeStart), int64(superBlock.SBMBlockStart)-1)
	if err != nil {
		return err
	}

	return cmd.generateTxt(text)
}

func (cmd *REP) repBMBlock() error {
	partition, path, err := global.GetMountedPartition(cmd.Id)
	if err != nil {
		return err
	}

	superBlock := &structures.SuperBlock{}
	if err := superBlock.ReadSuperBlock(path, int64(partition.PartStart)); err != nil {
		return err
	}

	text, err := utils.ReadFromBitMap(path, int64(superBlock.SBMBlockStart), int64(superBlock.SInodeStart)-1)
	if err != nil {
		return err
	}

	return cmd.generateTxt(text)
}

func (cmd *REP) repFile() error {
	partition, path, err := global.GetMountedPartition(cmd.Id)
	if err != nil {
		return err
	}

	superBlock := &structures.SuperBlock{}
	if err := superBlock.ReadSuperBlock(path, int64(partition.PartStart)); err != nil {
		return err
	}

	filePath := strings.Split(cmd.PathFileLs, "/")
	fileName := filePath[len(filePath)-1]

	sb := &structures.SuperBlock{}
	if err := sb.ReadSuperBlock(path, int64(partition.PartStart)); err != nil {
		return err
	}

	inodeIndex := sb.GetIndexInode(path, fileName, 0)
	if inodeIndex == -1 {
		return fmt.Errorf("file not found: %s", fileName)
	}

	content := sb.GetFile(path, inodeIndex, filePath)
	if content == "" {
		return fmt.Errorf("error reading file: %s", fileName)
	}

	return cmd.generateTxt(content)
}

func (cmd *REP) generateImage(content string) error {
	// Parse DOT content
	graph, err := graphviz.ParseBytes([]byte(content))
	if err != nil {
		return fmt.Errorf("error parsing DOT content: %w", err)
	}
	defer func() {
		if err := graph.Close(); err != nil {
			panic(fmt.Sprintf("error closing graph: %v", err))
		}
	}()

	// Create Graphviz instance
	g := graphviz.New()
	defer func() {
		if err := g.Close(); err != nil {
			panic(fmt.Sprintf("error closing Graphviz: %v", err))
		}
	}()

	// Create output directory
	dir := filepath.Dir(cmd.Path)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return fmt.Errorf("error creating directories: %w", err)
	}

	// Create output file
	outputFile, err := os.Create(cmd.Path)
	if err != nil {
		return fmt.Errorf("error creating output file: %w", err)
	}
	defer func() {
		if err := outputFile.Close(); err != nil {
			panic(fmt.Sprintf("error closing output file: %v", err))
		}
	}()

	ext := strings.ToLower(filepath.Ext(cmd.Path))
	var format graphviz.Format

	switch ext {
	case ".svg":
		format = graphviz.SVG
	case ".jpg", ".jpeg":
		format = graphviz.JPG
	case ".png":
		format = graphviz.PNG
	default:
		return fmt.Errorf("unsupported file format: %s", ext)
	}

	if err := g.Render(graph, format, outputFile); err != nil {
		return fmt.Errorf("error rendering image: %w", err)
	}

	fmt.Printf("Image generated successfully: %s\n", cmd.Path)

	return nil
}

func (cmd *REP) generateTxt(content string) error {
	ext := strings.ToLower(filepath.Ext(cmd.Path))
	if ext != ".txt" {
		return fmt.Errorf("unsupported file format: %s, only .txt files are allowed", ext)
	}

	dir := filepath.Dir(cmd.Path)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return fmt.Errorf("error creating directories: %w", err)
	}

	outputFile, err := os.Create(cmd.Path)
	if err != nil {
		return fmt.Errorf("error creating output file: %w", err)
	}
	defer func() {
		if err := outputFile.Close(); err != nil {
			panic(fmt.Sprintf("error closing output file: %v", err))
		}
	}()

	if _, err := outputFile.WriteString(content); err != nil {
		return fmt.Errorf("error writing to output file: %w", err)
	}

	fmt.Printf("Text generated successfully: %s\n", cmd.Path)

	return nil
}
