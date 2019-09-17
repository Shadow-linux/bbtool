// +build darwin
package main

import (
	"syscall"
	"time"
	"fmt"
	"strconv"
	"os"
)

func HandleTime (fileInfo os.FileInfo, latestDay int, sortType int) bool {
	stat := fileInfo.Sys().(*syscall.Stat_t)
	Log.Info(fileInfo.Name())
	Log.Info(fmt.Sprintf("访问时间: %v", T.TimespecToTime(stat.Atimespec)))
	Log.Info(fmt.Sprintf("创建时间: %v", T.TimespecToTime(stat.Ctimespec)))
	Log.Info(fmt.Sprintf("修改时间: %v", T.TimespecToTime(stat.Mtimespec)))
	nowTime := time.Now()
	strTime, _ := time.ParseDuration(fmt.Sprintf("-%sh", strconv.Itoa(latestDay * 24)))
	latestTime := nowTime.Add(strTime)
	switch sortType {
	// 修改时间
	case 0:
		if T.TimespecToTime(stat.Mtimespec).Unix() > latestTime.Unix() {
			Log.Info(fmt.Sprintf("排序类型: [修改时间] %d > %d", T.TimespecToTime(stat.Mtimespec).Unix(), latestTime.Unix()))
			return true
		}
		break
		// 访问时间
	case 1:
		if T.TimespecToTime(stat.Atimespec).Unix() > latestTime.Unix() {
			Log.Info(fmt.Sprintf("排序类型: [访问时间] %d > %d", T.TimespecToTime(stat.Atimespec).Unix(), latestTime.Unix()))
			return true
		}
		break
		// 创建时间
	case 2:
		if T.TimespecToTime(stat.Ctimespec).Unix() > latestTime.Unix() {
			Log.Info(fmt.Sprintf("排序类型: [创建时间] %d > %d", T.TimespecToTime(stat.Ctimespec).Unix(), latestTime.Unix()))
			return true
		}
		break
	}
	return false
}