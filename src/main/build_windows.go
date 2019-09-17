// +build windows
package main

import (
	"time"
	"fmt"
	"strconv"
)

func HandleTime (fileInfo os.FileInfo, latestDay int, sortType int) bool {
	nowTime := time.Now()
	strTime, _ := time.ParseDuration(fmt.Sprintf("-%sh", strconv.Itoa(latestDay * 24)))
	latestTime := nowTime.Add(strTime)
	fileSys := fileInfo.Sys().(*syscall.Win32FileAttributeData)
	Log.Info(fileInfo.Name())
	switch sortType {
	// 修改时间
	case 0:
		second := fileSys.LastWriteTime.Nanoseconds()/1e9
		if second > latestTime.Unix() {
			Log.Info(fmt.Sprintf("排序类型: [修改时间] %d > %d", second, latestTime.Unix()))
			return true
		}
		break
		// 访问时间
	case 1:
		second := fileSys.LastAccessTime.Nanoseconds()/1e9
		if second > latestTime.Unix() {
			Log.Info(fmt.Sprintf("排序类型: [访问时间] %d > %d", second, latestTime.Unix()))
			return true
		}
		break
		// 创建时间
	case 2:
		second := fileSys.CreationTime.Nanoseconds()/1e9
		if second > latestTime.Unix() {
			Log.Info(fmt.Sprintf("排序类型: [创建时间] %d > %d", T.TimespecToTime(stat.Ctimespec).Unix(), latestTime.Unix()))
			return true
		}
		break
	}
	return false
}