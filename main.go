package main

import (
	"fmt"
	"github.com/StackExchange/wmi"
	"github.com/lxn/walk"
	"log"
	"strconv"
	"syscall"
	"time"
)

var kernel = syscall.NewLazyDLL("Kernel32.dll")

type memoryStatusEx struct {
	cbSize                  uint32
	dwMemoryLoad            uint32
	ullTotalPhys            uint64 // in bytes
	ullAvailPhys            uint64
	ullTotalPageFile        uint64
	ullAvailPageFile        uint64
	ullTotalVirtual         uint64
	ullAvailVirtual         uint64
	ullAvailExtendedVirtual uint64
}

type cpuInfo struct {
	Name           string
	NumberOfCores  uint32
	ThreadCount    uint32
	LoadPercentage uint16
}

func main() {

	//GlobalMemoryStatusEx := kernel.NewProc("GlobalMemoryStatusEx")
	//var memInfo memoryStatusEx
	//memInfo.cbSize = uint32(unsafe.Sizeof(memInfo))
	//mem, _, _ := GlobalMemoryStatusEx.Call(uintptr(unsafe.Pointer(&memInfo)))
	//if mem == 0 {
	//	return
	//}
	////var total float64
	//total := float64(memInfo.ullTotalPhys) / 1024 / 1024 / 1024
	////total = memInfo.ullTotalPhys/1024/1024/1024
	//fmt.Println("total=:", total)
	//fmt.Println("free=:", memInfo.ullAvailPhys)

	//GuiInit()

	cpuUserData := make(chan uint16)

	go GetCpuPercentage(&cpuUserData)
	GuiInit(&cpuUserData)
}

func GetCpuPercentage(ch *chan uint16) {
	for {
		//time.Sleep(time.Second * 2)
		var cpuinfo []cpuInfo

		err := wmi.Query("Select * from Win32_Processor", &cpuinfo)
		if err != nil {
			return
		}
		// fmt.Sprintf("%.2f",float64(100))
		if len(cpuinfo) > 0 {
			thisCpu := cpuinfo[0]
			data := thisCpu.LoadPercentage
			*ch <- data
		}
	}

}

func GuiInit(ch *chan uint16) {
	mw, err := walk.NewMainWindow()
	if err != nil {
		log.Fatal(err)
	}
	//托盘图标文件
	icon, err := walk.Resources.Icon("./icon/1.ico")
	if err != nil {
		log.Fatal(err)
	}
	ni, err := walk.NewNotifyIcon(mw)
	if err != nil {
		log.Fatal(err)
	}
	defer ni.Dispose()
	if err := ni.SetIcon(icon); err != nil {
		log.Fatal(err)
	}
	speed := 0
	go func() {
		for data := range *ch {
			fmt.Println(data)

			if data <= 30 {
				speed = 100000
			} else if data <= 60 {
				speed = 10000
			} else {
				speed = 1000
			}
		}
	}()

	go func() {
		for {
			if speed != 0 {
				for num := 1; num <= 6; num++ {
					time.Sleep(time.Microsecond * time.Duration(speed))
					numString := strconv.Itoa(num)
					icon, err := walk.Resources.Icon("./run/" + numString + ".ico")
					if err != nil {
						log.Fatal(err)
					}
					if err := ni.SetIcon(icon); err != nil {
						log.Fatal(err)
					}
				}
			}

		}

	}()

	if err := ni.SetToolTip("鼠标在icon上悬浮的信息."); err != nil {
		log.Fatal(err)
	}
	//ni.MouseDown().Attach(func(x, yint, button walk.MouseButton) {
	//	if button != walk.LeftButton {
	//		return
	//	}
	//	if err := ni.ShowCustom("Walk 任务栏通知标题","walk 任务栏通知内容"); err != nil {
	//		og.Fatal(err)
	//	}
	//})
	exitAction := walk.NewAction()
	if err := exitAction.SetText("右键icon的菜单按钮"); err != nil {
		log.Fatal(err)
	}
	//Exit 实现的功能
	exitAction.Triggered().Attach(func() { walk.App().Exit(0) })
	if err := ni.ContextMenu().Actions().Add(exitAction); err != nil {
		log.Fatal(err)
	}
	if err := ni.SetVisible(true); err != nil {
		log.Fatal(err)
	}
	if err := ni.ShowInfo("Walk NotifyIcon Example", "Click the icon to show again."); err != nil {
		log.Fatal(err)
	}
	mw.Run()
}
