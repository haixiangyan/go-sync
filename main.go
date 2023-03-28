package main

import (
	"github.com/haixiangyan/go-sync/server"
	"github.com/haixiangyan/go-sync/server/config"
	"github.com/zserge/lorca"
	"os"
	"os/signal"
	"syscall"
)

// 把 frontend/dist 打包到 .exe 可执行包

func main() {
	go server.Run()

	ui := startBrowser()

	listenToInterrupt(ui)
}

func startBrowser() (ui lorca.UI) {
	port := config.GetPort()

	// 生成新窗口
	ui, _ = lorca.New("http://127.0.0.1:"+port+"/static/index.html", "", 800, 600, "--disable-sync", "--disable-translate", "--remote-allow-origins=*")

	return
}

func listenToInterrupt(ui lorca.UI) {
	// 创建一个操作系统 Channel
	chSignal := make(chan os.Signal, 1)

	// 监听中断、中止信号
	signal.Notify(chSignal, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-ui.Done():
	case <-chSignal:
	}
}
