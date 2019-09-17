// +build darwin
package main

import (
	"syscall"
	"time"
	"fmt"
	"strconv"
)

func DarwinHandleTime (stat *syscall.Stat_t) bool {
	nowTime := time.Now()
	strTime, _ := time.ParseDuration(fmt.Sprintf("-%sh", strconv.Itoa(s.LatestDays * 24)))
	latestTime := nowTime.Add(strTime)
	switch s.SortType {
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