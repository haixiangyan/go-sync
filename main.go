package main

import (
	"github.com/gin-gonic/gin"
	"github.com/zserge/lorca"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	go func() {
		gin.SetMode(gin.DebugMode)

		router := gin.Default()

		router.GET("/", func(c *gin.Context) {
			c.Writer.WriteString("123")
		})

		router.Run(":8080")
	}()

	var ui lorca.UI

	ui, _ = lorca.New("http://127.0.0.1:8080", "", 800, 600, "--disable-sync", "--disable-translate", "--remote-allow-origins=*")

	// 创建一个操作系统 Channel
	chSignal := make(chan os.Signal, 1)

	// 监听中断、中止信号
	signal.Notify(chSignal, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-ui.Done():
	case <-chSignal:
	}
}
