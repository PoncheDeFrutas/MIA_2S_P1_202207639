package utils

import (
	"os"
	"syscall"
	"time"
)

func dateTimeToBytes(t time.Time) [4]byte {
	var b [4]byte
	timestamp := uint32(t.Unix())
	b[0] = byte(timestamp >> 24)
	b[1] = byte(timestamp >> 16)
	b[2] = byte(timestamp >> 8)
	b[3] = byte(timestamp)
	return b
}

func bytesToDateTime(b [4]byte) time.Time {
	timestamp := int64(b[0])<<24 | int64(b[1])<<16 | int64(b[2])<<8 | int64(b[3])
	return time.Unix(timestamp, 0)
}

func GetCreationDate(path string) ([4]byte, error) {
	info, err := os.Stat(path)
	if err != nil {
		return [4]byte{}, err
	}

	stat := info.Sys().(*syscall.Stat_t)
	creationTime := time.Unix(int64(stat.Ctim.Sec), int64(stat.Ctim.Nsec))
	return dateTimeToBytes(creationTime), nil
}

func GetModificationDate(path string) ([4]byte, error) {
	info, err := os.Stat(path)
	if err != nil {
		return [4]byte{}, err
	}

	modificationTime := info.ModTime()
	return dateTimeToBytes(modificationTime), nil
}

func GetAccessDate(path string) ([4]byte, error) {
	info, err := os.Stat(path)
	if err != nil {
		return [4]byte{}, err
	}

	stat := info.Sys().(*syscall.Stat_t)
	accessTime := time.Unix(int64(stat.Atim.Sec), int64(stat.Atim.Nsec))
	return dateTimeToBytes(accessTime), nil
}

func ReadDate(bytes [4]byte) string {
	t := bytesToDateTime(bytes)
	return t.Format("2006-01-02 15:04:05")
}
