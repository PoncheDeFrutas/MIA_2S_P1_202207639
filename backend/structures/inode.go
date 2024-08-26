package structures

import (
	"backend/utils"
	"fmt"
	"time"
)

type Inode struct {
	IuId   int32
	IGid   int32
	ISize  int32
	IAtime float32
	ICTime float32
	IMTime float32
	IBlock [15]int32
	IType  byte
	IPerm  [3]byte
}

func (i *Inode) DefaultValue(blockCount int32) {
	i.IuId = 1
	i.IGid = 1
	i.ISize = 0
	i.IAtime = float32(time.Now().Unix())
	i.ICTime = float32(time.Now().Unix())
	i.IMTime = float32(time.Now().Unix())
	i.IBlock[0] = blockCount
	for j := 1; j < 15; j++ {
		i.IBlock[j] = -1
	}
	i.IType = '0'
	for j := 0; j < 3; j++ {
		i.IPerm[j] = '7'
	}
}

func (i *Inode) WriteInode(path string, offset int64, maxSize int64) error {
	if err := utils.WriteToFile(path, offset, maxSize, i); err != nil {
		return err
	}
	return nil
}

func (i *Inode) ReadInode(path string, offset int64) error {
	if err := utils.ReadFromFile(path, offset, i); err != nil {
		return err
	}
	return nil
}

func (i *Inode) Print() {
	fmt.Printf("IuId: %d\n", i.IuId)
	fmt.Printf("IGid: %d\n", i.IGid)
	fmt.Printf("ISize: %d\n", i.ISize)
	fmt.Printf("IAtime: %s\n", time.Unix(int64(i.IAtime), 0))
	fmt.Printf("ICTime: %s\n", time.Unix(int64(i.ICTime), 0))
	fmt.Printf("IMTime: %s\n", time.Unix(int64(i.IMTime), 0))
	fmt.Printf("IBlock: %v\n", i.IBlock)
	fmt.Printf("IType: %c\n", i.IType)
	fmt.Printf("IPerm: %s\n", string(i.IPerm[:]))
}
