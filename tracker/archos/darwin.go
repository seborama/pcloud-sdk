// +build darwin

package archos

import (
	"os"
	"syscall"
	"time"
)

func CreatedTime(fInfo os.FileInfo) time.Time {
	return time.Unix(fInfo.Sys().(*syscall.Stat_t).Ctimespec.Sec, fInfo.Sys().(*syscall.Stat_t).Ctimespec.Nsec)
}

func Device(fInfo os.FileInfo) int32 {
	return fInfo.Sys().(*syscall.Stat_t).Dev
}

func Inode(fInfo os.FileInfo) uint64 {
	return fInfo.Sys().(*syscall.Stat_t).Ino
}
