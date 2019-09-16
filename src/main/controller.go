package main

import (
	"fmt"
	"github.com/mattn/go-gtk/gtk"
	"io/ioutil"
	"strings"
	"strconv"
	"os"
	"time"
	"syscall"
	"runtime"
	"os/exec"
	"path/filepath"
)

/*
Control
	文件更名: ChangeFileNameControl
	文件排序: SortFileNameControl
*/

type CommonControl struct {
	Workspace string
	// 源文件列表
	SrcFileList []map[string]map[string]bool
	// 目标文件列表
	TargetFileList []string
	// 源sort tab 源文件列表
	SortTabSrcFileList []string
	// 目标 sort tab 目标文件列表
	SortTabTargetFileList []string
	// 重载源文件列表
	ReloadCNSrcSignal chan bool
	// 重载目标文件列表
	ReloadTargetSignal chan bool
	// 清空目标文件列表
	EmptyTargetSignal chan bool

	// 重载SortTab源文件列表
	ReloadSortSrcSignal chan bool
	// 重载SortTab目标文件列表
	ReloadSortTargetSignal chan bool

	// 退出信号
	ExitSignal chan bool

}

func (c *CommonControl) GetSystemType() string {
	sysType := runtime.GOOS
	return sysType
}

func (c *CommonControl) ReloadSrcFileList () {
	sysType := c.GetSystemType()
	c.SrcFileList = []map[string]map[string]bool{}
	c.SortTabSrcFileList = []string{}
	// 读入该目录下的文件
	Log.Info("重载: %s", c.Workspace)
	files, err := ioutil.ReadDir(c.Workspace)
	if err == nil {
		for _, f := range files {
			tmpMap := map[string]map[string]bool{
				f.Name(): {
					"status": false,
				},
			}
			c.SrcFileList = append(c.SrcFileList, tmpMap)
			switch sysType {
			case "darwin":
				c.SortTabSrcFileList = append(c.SortTabSrcFileList, fmt.Sprintf("%s/%s",c.Workspace, f.Name()))
				break
			case "windows":
				c.SortTabSrcFileList = append(c.SortTabSrcFileList, fmt.Sprintf("%s\\%s",c.Workspace, f.Name()))
			}
		}
	}
	//fmt.Println(this.SrcFileList)
}

func (c *CommonControl) SetSrcFileListEmpty () {
	c.SrcFileList = []map[string]map[string]bool{}
}


func (c *CommonControl) SetTargetFileListEmpty () {
	c.TargetFileList = []string{}
}


type ChangeFileNameControl struct {
	// 是否执行替换
	IsInstead bool
	MatchInput *gtk.Entry
	InsteadInput *gtk.Entry
	// 是否执行排序命名
	IsSortChangeName bool
	SortChangeNameInput *gtk.Entry
	// 是否执行添加前缀
	IsAddPrefix bool
	AddPrefixInput *gtk.Entry
	// 添加通用的模块
	CControl *CommonControl
	// 待处理list
	ProcessHandleList []string
	// 原始列表
	originHandleList []string

}

// 过滤出被选中的 file list
func (c *ChangeFileNameControl) filterSrcFileList () {
	c.ProcessHandleList = []string{}
	for _, filenameStatus := range c.CControl.SrcFileList {
		for filename, status := range filenameStatus {
			if status["status"] {
				c.ProcessHandleList = append(c.ProcessHandleList, filename)
			}
		}
	}
	c.originHandleList = c.ProcessHandleList
}

func (c *ChangeFileNameControl) Instead() {
	matchText := strings.TrimSpace(c.MatchInput.GetText())
	insteadText := strings.TrimSpace(c.InsteadInput.GetText())
	Log.Info("ChangeFileNameControl Instead")
	Log.Info("matchInput: %v \t insteadInput: %v ",matchText, insteadText)
	for _, filename := range c.ProcessHandleList {
		newFilename := strings.Replace(filename, matchText, insteadText, 100)
		c.CControl.TargetFileList = append(c.CControl.TargetFileList, newFilename)
	}
}

func (c *ChangeFileNameControl) SortChangeName() {
	input := strings.TrimSpace(c.SortChangeNameInput.GetText())
	Log.Info("ChangeFileNameControl SortChangeName")
	Log.Info("input: %s \n", input)
	n := 01
	for _, filename := range c.ProcessHandleList  {
		filenameL := strings.Split(filename, ".")
		newFilename := fmt.Sprintf("%s_%s.%s", input, strconv.Itoa(n), filenameL[1])
		c.CControl.TargetFileList = append(c.CControl.TargetFileList, newFilename)
		n++
	}
}

func (c *ChangeFileNameControl) AddPrefix() {
	input := strings.TrimSpace(c.AddPrefixInput.GetText())
	Log.Info("ChangeFileNameControl AddPrefix")
	Log.Info("input: %s \n", input)
	for _, filename := range c.ProcessHandleList {
		c.CControl.TargetFileList = append(c.CControl.TargetFileList, input + filename)
	}
}

// 最终执行时的必走流程
func (c *ChangeFileNameControl) commonExecute ()  {
	c.filterSrcFileList()
	// 清空target 列表
	c.CControl.SetTargetFileListEmpty()
	if c.IsInstead {
		c.Instead()
		// 处理完成后，赋值给下一个用
		c.ProcessHandleList = c.CControl.TargetFileList
		c.CControl.SetTargetFileListEmpty()
	}
	if c.IsSortChangeName {
		c.SortChangeName()
		// 处理完成后，赋值给下一个用
		c.ProcessHandleList = c.CControl.TargetFileList
		c.CControl.SetTargetFileListEmpty()
	}
	if c.IsAddPrefix {
		c.AddPrefix()
		// 处理完成后，赋值给下一个用
		c.ProcessHandleList = c.CControl.TargetFileList
		c.CControl.SetTargetFileListEmpty()
	}
	c.CControl.TargetFileList = c.ProcessHandleList
	c.CControl.ReloadTargetSignal <- true
}

func (c *ChangeFileNameControl) PreView() {
	Log.Info("ChangeFileNameControl PreView")
	c.commonExecute()
}

func (c *ChangeFileNameControl) Execute() {
	Log.Info("ChangeFileNameControl Execute")
	c.commonExecute()
	var oldFilename, newFilename string
	for idx, filename := range c.originHandleList {
		oldFilename = fmt.Sprintf("%s/%s",c.CControl.Workspace, filename)
		newFilename = fmt.Sprintf("%s/%s",c.CControl.Workspace, c.CControl.TargetFileList[idx])
		os.Rename(oldFilename, newFilename)
	}

}


type SortFileNameControl struct {
	LatestDays int
	SortType int
	SortTypeList []string
	FilePath string
	// 添加通用的模块
	CControl *CommonControl
}

func (s *SortFileNameControl) handleTime (stat *syscall.Stat_t) bool {
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

func (s *SortFileNameControl) Sort () {
	Log.Info("进行文件排序操作, 最近 %d 天", s.LatestDays)
	s.CControl.SortTabTargetFileList = []string{}
	for _, file := range s.CControl.SortTabSrcFileList {
		fileInfo, err := os.Stat(file)
		if err != nil {
			Log.Error("获取文件状态失败, err: %v", err)
		}
		stat := fileInfo.Sys().(*syscall.Stat_t)
		Log.Info(file)
		Log.Info(fmt.Sprintf("访问时间: %v", T.TimespecToTime(stat.Atimespec)))
		Log.Info(fmt.Sprintf("创建时间: %v", T.TimespecToTime(stat.Ctimespec)))
		Log.Info(fmt.Sprintf("修改时间: %v", T.TimespecToTime(stat.Mtimespec)))
		if s.handleTime(stat) {
			s.CControl.SortTabTargetFileList = append(s.CControl.SortTabTargetFileList, file)
		}
	}
	s.CControl.ReloadSortTargetSignal <- true
}

func (s *SortFileNameControl) OpenFile (filePath string) error {
	var (
		cmd *exec.Cmd
		err error
		sysType = s.CControl.GetSystemType()
	)
	Log.Info("打开文件: %s", filePath)
	switch sysType {
	case "darwin":
		cmd = exec.Command("open", filePath)
		cmd.Stdout = os.Stdout
	case "windows":
		cmd = exec.Command("start", filePath)
		cmd.Stdout = os.Stdout
	default:
		return err
	}
	err = cmd.Start()
	if cmd.Stdout != nil {
		Log.Warning("%v", cmd.Stdout)
	}
	if err != nil {
		return fmt.Errorf(" [sys]: %v", err)
	}
	return err
}

func (s *SortFileNameControl) OpenDir (filePath string) error {
	var (
		cmd *exec.Cmd
		err error
		sysType = s.CControl.GetSystemType()
	)
	dirPath := filepath.Dir(filePath)
	Log.Info("打开文件目录: %s", dirPath)
	switch sysType {
	case "darwin":
		cmd = exec.Command("open", dirPath)
		cmd.Stdout = os.Stdout
	case "windows":
		cmd = exec.Command("start", filePath)
		cmd.Stdout = os.Stdout
	default:
		return err
	}
	err = cmd.Start()
	if cmd.Stdout != nil {
		Log.Warning("%v", cmd.Stdout)
	}
	if err != nil {
		return fmt.Errorf(" [sys]: %v", err)
	}
	return err
}
