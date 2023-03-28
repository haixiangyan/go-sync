package server

import (
	"embed"
	"github.com/gin-gonic/gin"
	"github.com/haixiangyan/go-sync/server/config"
	"github.com/haixiangyan/go-sync/server/controller"
	"github.com/haixiangyan/go-sync/server/ws"
	"io/fs"
	"log"
	"net/http"
	"strings"
)

//go:embed frontend/dist/*
var FS embed.FS

func Run() {
	port := config.GetPort()

	gin.SetMode(gin.DebugMode)

	router := gin.Default()

	// 生成静态文件架构目录
	staticFiles, _ := fs.Sub(FS, "frontend/dist")
	// 静态文件前缀处理
	router.StaticFS("/static", http.FS(staticFiles))

	hub := ws.NewHub()
	go hub.Run()

	// 主路由
	router.POST("/api/v1/files", controller.FilesController)
	router.POST("/api/v1/texts", controller.TextsController)
	router.GET("/uploads/:path", controller.UploadsController)
	router.GET("/api/v1/addresses", controller.AddressesController)
	router.GET("/api/v1/qrcodes", controller.QrcodeController)
	router.GET("/ws", func(context *gin.Context) {
		ws.HttpController(context, hub)
	})

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

	router.Run(":" + port)
}
