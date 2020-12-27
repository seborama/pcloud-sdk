// +build linux

package archos

import (
	"os"
	"syscall"
	"time"
)

func CreatedTime(fInfo os.FileInfo) time.Time {
	return time.Unix(int64(fInfo.Sys().(*syscall.Stat_t).Ctim.Sec), int64(fInfo.Sys().(*syscall.Stat_t).Ctim.Nsec))
}

func Device(fInfo os.FileInfo) uint64 {
	return fInfo.Sys().(*syscall.Stat_t).Dev
}

func Inode(fInfo os.FileInfo) uint64 {
	return fInfo.Sys().(*syscall.Stat_t).Ino
}
