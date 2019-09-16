package main

import (
	"github.com/astaxie/beego/logs"
	"syscall"
	"time"
	"fmt"
)

/*
Util
*/

type Tools struct {
	Logger *logs.BeeLogger
}

func (t *Tools) InitLogger () {
	t.Logger = logs.NewLogger(1)
	t.Logger.SetLogger("console")
	//jsonConfig := map[string]interface{}{
	//	"filename": LogFile,
	//}
	//configStr, _ := json.Marshal(jsonConfig)
	//t.Logger.SetLogger("file", string(configStr))
	// 设置函数和行号
	t.Logger.EnableFuncCallDepth(true)
	t.Logger.SetLevel(logs.LevelDebug)
	t.Logger.Info("Initialize logger succeed.")
}

func (t *Tools) TimespecToTime(ts syscall.Timespec) time.Time  {
	return time.Unix(int64(ts.Sec), int64(ts.Nsec))
}

func (t *Tools) ConcatFilePath (sysType, dirname, filename string) string  {
	switch sysType {
	case "darwin":
		return fmt.Sprintf("%s/%s", dirname, filename)
	case "windows":
		return fmt.Sprintf("%s\\%s", dirname, filename)
	}
	return ""
}
