package main

import (
	"github.com/mattn/go-gtk/gtk"
)

// ================================= Constant =================================
const ProjectName = "BB 文件管理工具"
const Version = "v1.0.0"
const CopyRight = "@Created by Golang GTK"
const AuthorName = "Shadow-YD"
// wechat pay
const WeChatPayImg = "./wechat.jpg"
// 主体窗口size
const MainBoxWidth = 700
const MainBoxHeight = 690
// 日志文件
const LogFile = "./bb_tool.log"


// ================================= Variable =================================
var (
	Author = []string{"Shadow-YD", "972367265@qq.com"}
	Window *gtk.Window
	Workspace string
)

