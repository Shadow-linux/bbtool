package main

import (
	"github.com/mattn/go-gtk/gtk"
	"github.com/mattn/go-gtk/glib"
	"strconv"
	"fmt"
)

var (
	cMatchInput *gtk.Entry
	cInsteadInput *gtk.Entry
	cSortChangeNameInput *gtk.Entry
	// 通用控制器
	commonCtrl = CommonControl{
		Workspace: "",
		ReloadCNSrcSignal: make(chan bool, 1),
		ReloadSortSrcSignal: make(chan bool, 1),
		EmptyTargetSignal: make(chan bool, 1),
		ReloadTargetSignal: make(chan bool, 1),
		ReloadSortTargetSignal: make(chan bool, 1),
		ExitSignal: make(chan bool, 1),
	}
	// 文件变更控制器
	changeNameCtrl = ChangeFileNameControl{
		IsInstead: false,
		MatchInput: cMatchInput,
		InsteadInput: cInsteadInput,
		IsSortChangeName: false,
		SortChangeNameInput: cSortChangeNameInput,
		IsAddPrefix: false,
		CControl: &commonCtrl,
		}
	sortFileCtrl = SortFileNameControl{
		LatestDays: 1,
		SortType: 0,
		SortTypeList: []string{"最后修改时间", "最后访问时间", "创建时间"},
		FilePath: "",
		CControl: &commonCtrl,
	}
)

func InitView() (mainBox *gtk.VBox) {
	commonCtrl.GetSystemType()
	Log.Info("系统类型: %s", commonCtrl.SysType)
	// mainBox
	mainBox = gtk.NewVBox(false, 1)
	mainBox.ShowAll()

	// 顶部menu
	menuBar := gtk.NewMenuBar()
	mainBox.PackStart(menuBar, false, false, 0)
	MenuBarView(menuBar)
	menuBar.ShowAll()

	// 通用区
	commonAreaVbox := gtk.NewVBox(false, 5)
	mainBox.PackStart(commonAreaVbox, false,false, 10)
	CommonAreaView(commonAreaVbox)
	commonAreaVbox.ShowAll()

	// (Tab)
	notebook := gtk.NewNotebook()
	mainBox.PackStart(notebook, false, false, 0)
	NotebookView(notebook)
	notebook.Show()

	return
}

// 顶部
func MenuBarView(menubar *gtk.MenuBar) {
	cascadeMenu := gtk.NewMenuItemWithMnemonic("_Help")
	menubar.Append(cascadeMenu)
	submenu := gtk.NewMenu()
	cascadeMenu.SetSubmenu(submenu)
	// ------- About
	menuItem := gtk.NewMenuItemWithMnemonic("_About")
	menuItem.Connect("activate", func() {
		dialog := gtk.NewAboutDialog()
		dialog.SetName(ProjectName)
		dialog.SetPosition(gtk.WIN_POS_CENTER)
		dialog.SetProgramName(ProjectName)
		dialog.SetComments(fmt.Sprintf("Author: %s", AuthorName))
		dialog.SetAuthors(Author)
		dialog.SetVersion(Version)
		dialog.SetCopyright(CopyRight)
		dialog.Response(func() {
			dialog.Destroy()
		})
		dialog.Run()
	})
	submenu.Append(menuItem)
	// ------- 打赏
	menuItem = gtk.NewMenuItemWithMnemonic("~~客官，捧个场，感谢！")
	menuItem.Connect("activate", func() {

		dialog := gtk.NewDialog()
		dialog.SetName(ProjectName)
		dialog.SetPosition(gtk.WIN_POS_CENTER)
		dialog.SetSizeRequest(300, 410)
		dialog.SetTitle("扫一扫，谢谢支持")

		vBox := gtk.NewVBox(true, 1)

		var codeImage *gtk.Image
		switch commonCtrl.SysType {
		case "darwin":
			codeImage = gtk.NewImageFromFile(LinuxWeChatPayImg)
			break
		case "windows":
			codeImage = gtk.NewImageFromFile(WindowsChatPayImg)
			break
		}
		vBox.Add(codeImage)
		vBox.ShowAll()
		// 添加框内的对象
		dVbox := dialog.GetContentArea()
		dVbox.Add(vBox)
		dialog.Response(func() {
			dialog.Destroy()
		})
		dialog.Run()
	})

	submenu.Append(menuItem)
	// ------ Other
}

// notice warning dialog
func NoticeWarnDialog (message string)  {
	notice := gtk.NewMessageDialog(
		Window,
		gtk.DIALOG_MODAL,
		gtk.MESSAGE_WARNING,
		gtk.BUTTONS_CLOSE,
		message,
	)
	notice.Response(func() {
		notice.Destroy()
	})
	notice.Run()
}

func NoticeErrorDialog (message string)  {
	notice := gtk.NewMessageDialog(
		Window,
		gtk.DIALOG_MODAL,
		gtk.MESSAGE_ERROR,
		gtk.BUTTONS_CLOSE,
		message,
	)
	notice.Response(func() {
		notice.Destroy()
	})
	notice.Run()
}

// 通用区域
func CommonAreaView(commonAreaVbox *gtk.VBox)  {
	commonFrame := gtk.NewFrame("Common")
	commonFrame.SetSizeRequest(MainBoxWidth - 100, 60)

	// 添加一个fixed
	fixedBox1 := gtk.NewFixed()

	// ================= 添加一个选择目录的btn ===================
	chooseDirLabel := gtk.NewLabel("工作目录:")
	chooseDirEntry := gtk.NewEntry()
	chooseDirEntry.SetSizeRequest(420, 25)
	//chooseDirEntry.SetText("Hello world")
	chooseDirBtn := gtk.NewButtonFromStock(gtk.STOCK_OPEN)
	chooseDirBtn.SetSizeRequest(75, 25)
	// choose dir dialog
	chooseDirBtn.Clicked(func() {
		fileChooserDialog := gtk.NewFileChooserDialog(
			"选择工作目录（只允许选择目录）",
			chooseDirBtn.GetTopLevelAsWindow(),
			gtk.FILE_CHOOSER_ACTION_SELECT_FOLDER,
			gtk.STOCK_OK,
			gtk.RESPONSE_ACCEPT)
		fileChooserDialog.Response(func() {
			Workspace = fileChooserDialog.GetFilename()
			// 设置workspace
			chooseDirEntry.SetText(Workspace)
			// [control]
			commonCtrl.Workspace = Workspace
			// 重载文件列表
			commonCtrl.ReloadSrcFileList()
			commonCtrl.SetTargetFileListEmpty()
			// 给个信号重载src list target list
			commonCtrl.ReloadCNSrcSignal <- true
			commonCtrl.ReloadSortSrcSignal <- true
			commonCtrl.EmptyTargetSignal <- true
			fileChooserDialog.Destroy()
		})
		fileChooserDialog.Run()
	})

	// 重载当前目录按钮
	reloadBtn := gtk.NewButtonFromStock(gtk.STOCK_REFRESH)
	reloadBtn.SetSizeRequest(75, 25)
	reloadBtn.Clicked(func() {
		if len(commonCtrl.Workspace) == 0 {
			NoticeWarnDialog("请选择工作目录后再操作.")
			return
		}
		commonCtrl.ReloadSrcFileList()
		commonCtrl.ReloadCNSrcSignal <- true
		commonCtrl.ReloadSortSrcSignal <- true
 	})

	fixedBox1.Put(chooseDirLabel, 10 ,12)
	fixedBox1.Put(chooseDirEntry, 70, 10)
	fixedBox1.Put(chooseDirBtn, 510, 9)
	fixedBox1.Put(reloadBtn, 600, 9)
	fixedBox1.Add(chooseDirLabel)
	fixedBox1.Add(chooseDirEntry)
	fixedBox1.Add(chooseDirBtn)
	fixedBox1.Add(reloadBtn)
	// ============ ============ ============ ============

	commonFrame.Add(fixedBox1)
	commonAreaVbox.Add(commonFrame)
}

//  Tab
func NotebookView(notebook *gtk.Notebook)  {

	changeNameTab := gtk.NewFixed()
	notebook.AppendPage(changeNameTab, gtk.NewLabel("文件更名"))
	notebook.SetSizeRequest(MainBoxWidth - 10, MainBoxHeight)
	changeNameTab.SetSizeRequest(MainBoxWidth - 10, 100)
	fixedChangeNameTabView(changeNameTab)
	changeNameTab.Show()

	fixedSortNameTab := gtk.NewFixed()
	notebook.AppendPage(fixedSortNameTab, gtk.NewLabel("文件排序"))
	fixedSortNameTabView(fixedSortNameTab)
	fixedSortNameTab.ShowAll()

}

// ========================= Tab change name =========================

func fixedChangeNameTabView (tabFixedBox *gtk.Fixed)  {
	// ================= 添加一个frame (操作区) =================
	actionFrame := gtk.NewFrame("Action")
	actionFrame.SetSizeRequest(MainBoxWidth - 10, 210)

	var (
		matchCharLabel, insteadCharLabel *gtk.Label
		matchCharInput, insteadCharInput *gtk.Entry
	)

	// Vbox
	Vbox := gtk.NewVBox(false ,1)

	// checkbox
	insteadCharCheckButton := gtk.NewCheckButtonWithLabel("匹配替换")
	insteadCharCheckButton.Connect("toggled", func() {
		if insteadCharCheckButton.GetActive() {
			Log.Info("匹配替换 => true")
			matchCharLabel.SetVisible(true)
			matchCharInput.SetVisible(true)
			insteadCharLabel.SetVisible(true)
			insteadCharInput.SetVisible(true)
			// control
			changeNameCtrl.IsInstead = true
		} else {
			Log.Info("匹配替换 => false")
			matchCharLabel.SetVisible(false)
			matchCharInput.SetVisible(false)
			insteadCharLabel.SetVisible(false)
			insteadCharInput.SetVisible(false)
			// control
			changeNameCtrl.IsInstead = false
		}
	})
	Vbox.PackStart(insteadCharCheckButton, false,false,5)
	insteadCharCheckButton.Show()
	// 匹配字符
	matchInsteadHBox := gtk.NewHBox(false ,1)
	matchInsteadHBox.Show()
	matchCharLabel = gtk.NewLabel("匹配字符:")
	matchCharInput = gtk.NewEntry()
	matchCharInput.SetSizeRequest(180, 25)
	matchInsteadHBox.PackStart(matchCharLabel, false,false,0)
	matchInsteadHBox.PackStart(matchCharInput,false,false,4)
	// 替换字符
	insteadCharLabel = gtk.NewLabel("替换字符:")
	insteadCharInput = gtk.NewEntry()
	insteadCharInput.SetSizeRequest(180, 25)
	matchInsteadHBox.PackStart(insteadCharLabel, false,false,0)
	matchInsteadHBox.PackStart(insteadCharInput, false,false,4)
	// [control]
	changeNameCtrl.MatchInput = matchCharInput
	changeNameCtrl.InsteadInput = insteadCharInput
	Vbox.PackStart(matchInsteadHBox, false,false,0)


	// 改名 cheekbox
	var (
		changeNameLabel *gtk.Label
		changeNameInput *gtk.Entry
		changeNameTipsLabel *gtk.Label
	)
	changeNameCheckButton := gtk.NewCheckButtonWithLabel("排序改名")
	changeNameCheckButton.Connect("toggled", func() {
		if changeNameCheckButton.GetActive() {
			Log.Info("排序改名 => true")
			changeNameLabel.Show()
			changeNameInput.Show()
			changeNameTipsLabel.Show()
			// [control]
			changeNameCtrl.IsSortChangeName = true
		} else {
			Log.Info("排序改名 => false")
			changeNameLabel.Hide()
			changeNameInput.Hide()
			changeNameTipsLabel.Hide()
			// [control]
			changeNameCtrl.IsSortChangeName = false
		}
	})
	Vbox.PackStart(changeNameCheckButton, false,false,0)
	changeNameCheckButton.Show()
	// 排序命名 hbox
	changeNameHBox := gtk.NewHBox(false,0)
	changeNameHBox.Show()
	// 排序命名 Label
	changeNameLabel = gtk.NewLabel("排序改名:")
	changeNameHBox.PackStart(changeNameLabel, false,false, 0)
	// 排序命名 input
	changeNameInput = gtk.NewEntry()
	changeNameInput.SetSizeRequest(180, 22)
	changeNameHBox.PackStart(changeNameInput,false,false,5)
	// 提示
	changeNameTipsLabel = gtk.NewLabel("description")
	cnTipsMarkup := "<span foreground='red' font_desc='10'>* [改名: aa] -> [过程: 1.txt => aa_1.txt]</span>"
	changeNameTipsLabel.SetMarkup(cnTipsMarkup)
	changeNameHBox.PackStart(changeNameTipsLabel,false,false,5)
	// [control]
	changeNameCtrl.SortChangeNameInput = changeNameInput
	Vbox.PackStart(changeNameHBox,false,false,0)


	// 添加前缀
	var (
		prefixLabel *gtk.Label
		prefixInput *gtk.Entry
	)
	// hbox
	prefixHBox := gtk.NewHBox(false,0)
	prefixHBox.Show()
	prefixCheckbutton := gtk.NewCheckButtonWithLabel("添加前缀")
	prefixCheckbutton.Connect("toggled", func() {
		if prefixCheckbutton.GetActive() {
			Log.Info("添加前缀 => true")
			prefixLabel.Show()
			prefixInput.Show()
			// [control]
			changeNameCtrl.IsAddPrefix = true
		} else {
			Log.Info("添加前缀 => false")
			prefixLabel.Hide()
			prefixInput.Hide()
			// [control]
			changeNameCtrl.IsAddPrefix = false
		}
	})
	Vbox.PackStart(prefixCheckbutton, false,false,5)
	prefixCheckbutton.Show()
	// 添加前缀
	prefixLabel = gtk.NewLabel("添加前缀:")
	prefixHBox.PackStart(prefixLabel, false,false,0)
	// 排序命名 下拉框
	prefixInput = gtk.NewEntry()
	prefixInput.SetSizeRequest(180, 22)
	prefixHBox.PackStart(prefixInput, false,false,5)
	// [control]
	changeNameCtrl.AddPrefixInput = prefixInput
	Vbox.PackStart(prefixHBox,false,false,4)


	// PS，代码优先级判断描述
	descPriorityLabel := gtk.NewLabel("description")
	markup := "<span foreground='red' font_desc='11'>执行优先级:  匹配替换 > 排序改名  > 添加前缀</span>"
	descPriorityLabel.SetMarkup(markup)
	tabFixedBox.Put(descPriorityLabel, 20, 525)
	tabFixedBox.Add(descPriorityLabel)
	descPriorityLabel.Show()

	// 预览按钮
	previewImg := gtk.NewImageFromStock(gtk.STOCK_ZOOM_FIT, 1)
	previewBtn := gtk.NewButton()
	previewBtn.Clicked(func() {
		// [control]
		changeNameCtrl.PreView()
		if len(commonCtrl.TargetFileList) == 0 {
			NoticeWarnDialog("未选中任何文件对象。")
		}
	})
	previewBtn.SetSizeRequest(100, 50)
	previewBtn.SetImage(previewImg)
	previewBtn.SetLabel("预览")
	tabFixedBox.Put(previewBtn,580, 360)
	tabFixedBox.Add(previewBtn)
	previewBtn.Show()

	// 确认按钮
	applyImg := gtk.NewImageFromStock(gtk.STOCK_APPLY, 1)
	applyBtn := gtk.NewButton()
	applyBtn.SetSizeRequest(100, 50)
	applyBtn.SetImage(applyImg)
	applyBtn.SetLabel("确认")

	applyBtn.Clicked(func() {
		if len(commonCtrl.TargetFileList) == 0 {
			NoticeWarnDialog("请先通过[预览]检查是否符合预期效果")
			return
		}
		changeNameCtrl.Execute()
		// [control]
		commonCtrl.ReloadSrcFileList()
		commonCtrl.SetTargetFileListEmpty()
		commonCtrl.ReloadCNSrcSignal <- true
		commonCtrl.ReloadSortSrcSignal <- true
		commonCtrl.EmptyTargetSignal <- true
		// dialog
		applyDialogMsgMarkup := "<span foreground='green' font_desc='14'>操作成功</span>"
		applyDialog := gtk.NewMessageDialog(
			applyBtn.GetTopLevelAsWindow(),
			gtk.DIALOG_MODAL,
			gtk.MESSAGE_OTHER,
			gtk.BUTTONS_CLOSE,
			"",
		)
		applyDialog.SetSizeRequest(150, 100)
		applyDialog.SetMarkup(applyDialogMsgMarkup)
		applyDialog.Response(func() {
			applyDialog.Destroy()
		})
		applyDialog.Run()
	})
	tabFixedBox.Put(applyBtn,580, 420)
	tabFixedBox.Add(applyBtn)
	applyBtn.Show()

	// 最后位置调整
	tabFixedBox.Put(actionFrame, 5, 340)
	tabFixedBox.Put(Vbox, 20 ,350)
	tabFixedBox.Add(Vbox)
	tabFixedBox.Add(actionFrame)
	actionFrame.Show()
	Vbox.Show()
	// ====================================================================
	// 源文件列表
	srcFileFrameView(tabFixedBox)
	// 目标列表
	targetFileFrameView(tabFixedBox)
}

// fixedChangeNameTab
func srcFileFrameView (fixed *gtk.Fixed)  {
	srcFileFrame := gtk.NewFrame("源列表")
	scrollWin := gtk.NewScrolledWindow(nil, nil)
	scrollWin.SetPolicy(gtk.POLICY_AUTOMATIC, gtk.POLICY_AUTOMATIC)
	scrollWin.SetSizeRequest(300, 300)

	store := gtk.NewListStore(glib.G_TYPE_BOOL, glib.G_TYPE_STRING)
	treeview := gtk.NewTreeView()
	// 设置存储
	treeview.SetModel(store)
	// 设置为多选模式
	treeview.GetSelection().SetMode(gtk.SELECTION_SINGLE)
	checkColumn := gtk.NewTreeViewColumnWithAttributes("check", gtk.NewCellRendererToggle(), "active", 0)
	nameColumn := gtk.NewTreeViewColumnWithAttributes("文件名", gtk.NewCellRendererText(), "text", 1)
	treeview.AppendColumn(checkColumn)
	treeview.AppendColumn(nameColumn)
	scrollWin.Add(treeview)
	srcFileFrame.Add(scrollWin)

	var numbers []map[string]map[string]bool

	// 重载src file list
	go func() {
		defer func() {
			if err := recover(); err != nil {
				Log.Error("重载 SrcFileList 数据出错, err: %v", err)
			}
		}()
		for {
			select {
			case _, ok := <- commonCtrl.ReloadCNSrcSignal:
				if ok {
					Log.Info("接收到重载 srcFileList 信号")
				}
			case _, ok := <- commonCtrl.ExitSignal:
				if ok {
					Log.Info("接收到退出信号")
					goto EXIT
				}
			}
			numbers = commonCtrl.SrcFileList
			// 清空列表
			store.Clear()
			Log.Info("重载 SrcFileList 数据")
			for _, filenameStatus := range numbers {
				for filename, status := range filenameStatus {
					var iter gtk.TreeIter
					store.Append(&iter)
					store.Set(&iter,
						0, status["status"],
						1, filename,
					)
				}
			}
		}
		EXIT:
	}()

	treeview.Connect("cursor-changed", func() {
		var path *gtk.TreePath
		var column *gtk.TreeViewColumn
		treeview.GetCursor(&path, &column)
		idx, _ := strconv.Atoi(path.String())

		// iter 是当前行的指针
		var iter gtk.TreeIter
		if treeview.GetModel().GetIter(&iter, path) {
			model := treeview.GetModel()
			Log.Info("chose src file %s", model.GetStringFromIter(&iter))
			// 设置是否被选中
			filenameStatus := numbers[idx]
			var realStatus bool
			for _, status := range filenameStatus{
				if status["status"] {
					realStatus = false
				} else {
					realStatus = true
				}
				status["status"] = realStatus
			}
			// 修改第一列数据
			store.SetValue(&iter, 0, realStatus)
			//fmt.Println(numbers)
		}
	})

	// 全选/取消全选
	fullBtn := gtk.NewCheckButtonWithLabel("全选")
	fullBtn.Show()
	fixed.Put(fullBtn, 282, 1)
	fixed.Add(fullBtn)
	fullBtn.Clicked(func() {
		var (
			iter gtk.TreeIter
		)
		model := treeview.GetModel()
		for idx, filenameStatus := range numbers {
			idxStr := strconv.Itoa(idx)
			model.GetIterFromString(&iter, idxStr)
			var realStatus bool
			for _, status := range filenameStatus{
				if fullBtn.GetActive() {
					realStatus = true
				} else {
					realStatus = false
				}
				status["status"] = realStatus
			}
			// 修改第一列数据
			store.SetValue(&iter, 0, realStatus)
		}
	})

	// 框整理
	fixed.Put(srcFileFrame, 25, 15)
	fixed.Add(srcFileFrame)
	srcFileFrame.ShowAll()
}

// target file list
func targetFileFrameView (fixed *gtk.Fixed)  {
	tragetFileFrame := gtk.NewFrame("目标列表")
	scrollWin := gtk.NewScrolledWindow(nil, nil)
	scrollWin.SetPolicy(gtk.POLICY_AUTOMATIC, gtk.POLICY_AUTOMATIC)
	scrollWin.SetSizeRequest(300, 300)

	store := gtk.NewListStore(glib.G_TYPE_STRING, )
	treeview := gtk.NewTreeView()
	// 设置存储
	treeview.SetModel(store)
	// 设置为多选模式
	treeview.GetSelection().SetMode(gtk.SELECTION_MULTIPLE)
	nameColumn := gtk.NewTreeViewColumnWithAttributes("修改预览", gtk.NewCellRendererText(), "text", 0)
	treeview.AppendColumn(nameColumn)
	scrollWin.Add(treeview)
	tragetFileFrame.Add(scrollWin)

	var numbers []string
	go func() {
		defer func() {
			if err := recover(); err != nil {
				Log.Error("重载TargetFileList数据出错, err: %v", err)
			}
		}()
		for {
			select {
			case _, ok := <- commonCtrl.EmptyTargetSignal:
				if ok {
					Log.Info("接收到清空目标列表信号.")
				}
			case _, ok := <- commonCtrl.ReloadTargetSignal:
				if ok {
					Log.Info("接收到重载目标列表信号.")
				}
			case _, ok := <- commonCtrl.ExitSignal:
				if ok {
					Log.Info("接收到退出信号")
					goto EXIT
				}
			}
			numbers = commonCtrl.TargetFileList
			Log.Info("重载 TargetFileList 数据")
			store.Clear()
			for _, filename := range numbers {
				var iter gtk.TreeIter
				store.Append(&iter)
				store.Set(&iter,
					0, filename,
				)
			}
		}
		EXIT:
	}()

	// 框整理
	fixed.Put(tragetFileFrame, 365, 15)
	fixed.Add(tragetFileFrame)
	tragetFileFrame.ShowAll()
}

// ===================================================================

// Tab sort name
func fixedSortNameTabView (fixed *gtk.Fixed)  {
	sortNameVBox := gtk.NewVBox(false, 2 )

	actionFrame := gtk.NewFrame("Action")
	actionFrame.SetSizeRequest(MainBoxWidth - 20, 80)
	SortNameActionView(sortNameVBox)

	tableFrame := gtk.NewFrame("Table")
	tableFrame.SetSizeRequest(MainBoxWidth - 20, 450)
	SortNameTableView(sortNameVBox)

	fixed.Put(actionFrame,10, 5)
	fixed.Add(actionFrame)
	fixed.Put(sortNameVBox,25, 30)
	fixed.Add(sortNameVBox)
	fixed.Put(tableFrame,10, 100)
	fixed.Add(tableFrame)
}

func SortNameActionView(vBox *gtk.VBox)  {
	hBox := gtk.NewHBox(false, 5)

	// 时间范围
	timeRangeLabel := gtk.NewLabel("最近(x)天:")
	hBox.PackStart(timeRangeLabel,false,false,0)
	timeRangeInput := gtk.NewEntry()
	// [control]
	timeRangeInput.SetText(strconv.Itoa(sortFileCtrl.LatestDays))
	hBox.PackStart(timeRangeInput, false,false,5)

	// 排序label
	sortTypeLabel := gtk.NewLabel("排序类型:")
	hBox.PackStart(sortTypeLabel, false,false,5)
	// combo
	comboBoxEntry := gtk.NewComboBoxText()
	comboBoxEntry.SetSizeRequest(120, 23)
	// [control]
	for _, item := range sortFileCtrl.SortTypeList {
		comboBoxEntry.AppendText(item)
	}
	comboBoxEntry.SetActive(0)
	hBox.PackStart(comboBoxEntry,false,false,0)

	// 排序按钮
	sortBtnImg := gtk.NewImageFromStock(gtk.STOCK_SELECT_ALL, 1)
	sortBtn := gtk.NewButton()
	sortBtn.SetLabel("排序")
	sortBtn.SetImage(sortBtnImg)
	sortBtn.SetSizeRequest(100, 28)
	// [control]
	sortBtn.Clicked(func() {
		if len(commonCtrl.Workspace) == 0 {
			NoticeWarnDialog("请选择目录后再进行文件排序.")
			return
		}
		latestDayStr := timeRangeInput.GetText()
		latestDay, err := strconv.Atoi(latestDayStr)
		if  err != nil {
			NoticeWarnDialog(fmt.Sprintf("请输入数字。\n Error: %v", err))
			return
		}
		sortFileCtrl.LatestDays = latestDay
		sortFileCtrl.SortType = comboBoxEntry.GetActive()
		sortFileCtrl.Sort()
	})
	hBox.PackStart(sortBtn, false,false,135)

	// 添加hbox
	vBox.PackStart(hBox, false, false, 5)
}

func SortNameTableView(vBox *gtk.VBox)  {
	hBox := gtk.NewHBox(false, 5)

	// ======  file list ====
	sVBox1 := gtk.NewVBox(false,5)
	scrollWin := gtk.NewScrolledWindow(nil, nil)
	scrollWin.SetPolicy(gtk.POLICY_AUTOMATIC, gtk.POLICY_AUTOMATIC)
	scrollWin.SetSizeRequest(550, 420)
	store := gtk.NewListStore(glib.G_TYPE_STRING)
	treeview := gtk.NewTreeView()
	// 设置存储
	treeview.SetModel(store)
	treeview.GetSelection().SetMode(gtk.SELECTION_SINGLE)
	nameColumn := gtk.NewTreeViewColumnWithAttributes("文件名(递归扫描)", gtk.NewCellRendererText(), "text", 0)
	treeview.AppendColumn(nameColumn)
	scrollWin.Add(treeview)
	// 选中
	treeview.Connect("cursor-changed", func() {
		var path *gtk.TreePath
		var column *gtk.TreeViewColumn
		treeview.GetCursor(&path, &column)
		idx, _ := strconv.Atoi(path.String())
		sortFileCtrl.FilePath = commonCtrl.SortTabTargetFileList[idx]
	})
	// 重载src file list
	go func() {
		for {
			select {
			case _, ok := <- commonCtrl.ReloadSortSrcSignal:
				if ok {
					Log.Info("接收到sort src signal")
					commonCtrl.SortTabTargetFileList = commonCtrl.SortTabSrcFileList
				}
			case _, ok := <- commonCtrl.ReloadSortTargetSignal:
				if ok {
					Log.Info("接收到sort target signal")
				}
			case _, ok := <- commonCtrl.ExitSignal:
				if ok {
					Log.Info("接收到退出信号")
					goto EXIT
				}
			}
			// 清空列表
			store.Clear()
			for _, filename := range commonCtrl.SortTabTargetFileList {
				var iter gtk.TreeIter
				store.Append(&iter)
				store.Set(&iter,
					0, filename,
				)
			}
		}
		EXIT:
	}()
	sVBox1.PackStart(scrollWin, false,false, 0)
	hBox.PackStart(sVBox1,false,false,0)

	// ====== Open button =======
	sVBox := gtk.NewVBox(false,0)
	openBtnImg := gtk.NewImageFromStock(gtk.STOCK_JUMP_TO, 1)
	openBtn := gtk.NewButton()
	openBtn.SetSizeRequest(80, 80)
	openBtn.SetImage(openBtnImg)
	openBtn.SetLabel("打开文件")
	openBtn.SetUSize(90, 80)
	openBtn.Clicked(func() {
		// [control]
		if sortFileCtrl.FilePath == "" {
			NoticeWarnDialog("请选择文件后才打开文件.")
			return
		}
		if err := sortFileCtrl.OpenFile(sortFileCtrl.FilePath); err != nil {
			Log.Error("打开文件失败: %s, err: %v", sortFileCtrl.FilePath, err)
			NoticeErrorDialog(fmt.Sprintf("打开文件失败.\nError: %v", err))
		}
	})
	sVBox.PackStart(openBtn, false,false,0)

	openDirBtnImg := gtk.NewImageFromStock(gtk.STOCK_OPEN, 1)
	openDirBtn := gtk.NewButton()
	openDirBtn.SetSizeRequest(90, 80)
	openDirBtn.SetImage(openDirBtnImg)
	openDirBtn.SetLabel("打开目录")
	openDirBtn.Clicked(func() {
		// [control]
		if sortFileCtrl.FilePath == "" {
			NoticeWarnDialog("请选择文件后才打开文件.")
			return
		}
		if err := sortFileCtrl.OpenDir(sortFileCtrl.FilePath); err != nil {
			Log.Error("打开文件目录: %s, err: %v", sortFileCtrl.FilePath, err)
			NoticeErrorDialog(fmt.Sprintf("打开文件目录.\nError: %v", err))
		}
	})
	sVBox.PackStart(openDirBtn,false,false,10)


	hBox.PackStart(sVBox,false,false,10)

	// 添加hbox
	vBox.PackStart(hBox, false,false,50)
}


// ======================================================================
