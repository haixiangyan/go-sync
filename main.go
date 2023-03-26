package main

import (
	"embed"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/zserge/lorca"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strings"
	"syscall"
)

// 把 frontend/dist 打包到 .exe 可执行文伯

//go:embed frontend/dist/*
var FS embed.FS

func main() {
	go func() {
		gin.SetMode(gin.DebugMode)

		router := gin.Default()

		// 生成静态文件架构目录
		staticFiles, _ := fs.Sub(FS, "frontend/dist")
		// 静态文件前缀处理
		router.StaticFS("/static", http.FS(staticFiles))

		// 主路由
		router.POST("/api/v1/texts", TextsController)

		// 404
		router.NoRoute(func(c *gin.Context) {
			path := c.Request.URL.Path

			// 如果为 static 开头，那么返回固定 index.html
			if strings.HasPrefix(path, "/static") {
				reader, err := staticFiles.Open("index.html")
				if err != nil {
					log.Fatal(err)
				}

				defer reader.Close()

				stat, err := reader.Stat()
				if err != nil {
					log.Fatal(err)
				}

				c.DataFromReader(http.StatusOK, stat.Size(), "text/html", reader, nil)
			} else {
				c.Status(http.StatusNotFound)
			}
		})

		router.Run(":8080")
	}()

	// 生成新窗口
	var ui lorca.UI
	ui, _ = lorca.New("http://127.0.0.1:8080/static/index.html", "", 800, 600, "--disable-sync", "--disable-translate", "--remote-allow-origins=*")

	// 创建一个操作系统 Channel
	chSignal := make(chan os.Signal, 1)

	// 监听中断、中止信号
	signal.Notify(chSignal, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-ui.Done():
	case <-chSignal:
	}
}

func TextsController(c *gin.Context) {
	var json struct {
		Raw string `json:"raw"`
	}

	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	} else {
		// 当前执行文件的路径
		exe, err := os.Executable()

		if err != nil {
			log.Fatal(err)
		}

		// 创建 /uploads 目录
		dir := filepath.Dir(exe)
		uploads := filepath.Join(dir, "uploads")
		err = os.MkdirAll(uploads, os.ModePerm)

		if err != nil {
			log.Fatal(err)
		}

		// 生成随机路径
		filename := uuid.New().String()
		fullpath := path.Join("uploads", filename+".txt")

		// 写文件
		err = os.WriteFile(filepath.Join(dir, fullpath), []byte(json.Raw), 0644)

		if err != nil {
			log.Fatal(err)
		}

		c.JSON(http.StatusOK, gin.H{"url": "/" + fullpath})
	}
}
