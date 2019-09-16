package main

import (
	"os"
	"github.com/mattn/go-gtk/gtk"
	//"fmt"
	"github.com/mattn/go-gtk/glib"
	"github.com/astaxie/beego/logs"
	"fmt"
	"time"
)

var (
	T Tools
	Log *logs.BeeLogger
)

func init()  {
	T = Tools{}
	T.InitLogger()
	Log = T.Logger
}

func start() bool {
	defer func() {
		if err := recover(); err != nil {
			// 让所有goroutine退出
			commonCtrl.ExitSignal <- true
			commonCtrl.ExitSignal <- true
			commonCtrl.ExitSignal <- true
			time.Sleep(3* time.Second)
			return
		}
	}()
	Log.Info("程序启动...")
	gtk.Init(&os.Args) //环境初始化
	Window = gtk.NewWindow(gtk.WINDOW_TOPLEVEL)
	Window.SetPosition(gtk.WIN_POS_CENTER)
	Window.SetTitle(ProjectName)
	Window.SetIconName("gtk-dialog-info")
	Window.Connect("destroy", func(ctx *glib.CallbackContext) {
		Log.Info("got destroy! %v", ctx.Data().(string))
		gtk.MainQuit()
	}, "foo")

	// View
	mainBox := InitView()
	Window.Add(mainBox)
	Window.SetSizeRequest(MainBoxWidth, MainBoxHeight)
	Window.Show()
	gtk.Main()
	return true
}

func main() {
	defer func() {
		if err := recover(); err != nil {
			Log.Error(fmt.Sprintf("程序出现崩溃请联系开发者 97236726@qq.com. err: %v", err))
		}
	}()
	start()
}
