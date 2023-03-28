package main

import (
	"embed"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/skip2/go-qrcode"
	"github.com/zserge/lorca"
	"io/fs"
	"log"
	"net"
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
	port := "27149"

	go func() {
		gin.SetMode(gin.DebugMode)

		router := gin.Default()

		// 生成静态文件架构目录
		staticFiles, _ := fs.Sub(FS, "frontend/dist")
		// 静态文件前缀处理
		router.StaticFS("/static", http.FS(staticFiles))

		// 主路由
		router.POST("/api/v1/files", FilesController)
		router.POST("/api/v1/texts", TextsController)
		router.GET("/uploads/:path", UploadsController)
		router.GET("/api/v1/addresses", AddressesController)
		router.GET("/api/v1/qrcodes", QrcodeController)

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
	}()

	// 生成新窗口
	var ui lorca.UI
	ui, _ = lorca.New("http://127.0.0.1:"+port+"/static/index.html", "", 800, 600, "--disable-sync", "--disable-translate", "--remote-allow-origins=*")

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

func AddressesController(c *gin.Context) {
	addrs, _ := net.InterfaceAddrs()

	var result []string

	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				result = append(result, ipnet.IP.String())
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"addresses": result})
}

func GetUploadsDir() (uploads string) {
	exe, err := os.Executable()

	if err != nil {
		log.Fatal(err)
	}

	dir := filepath.Dir(exe)
	uploads = filepath.Join(dir, "uploads")
	return
}

func UploadsController(c *gin.Context) {
	if path := c.Param("path"); path != "" {
		target := filepath.Join(GetUploadsDir(), path)

		c.Header("Content-Description", "File Transfer")
		c.Header("Content-Transfer-Encoding", "binary")
		c.Header("Content-Disposition", "attachment; filename="+path)
		c.Header("Content-Type", "application/octet-stream")
		c.File(target)
	} else {
		c.Status(http.StatusNotFound)
	}
}

func QrcodeController(c *gin.Context) {
	if content := c.Query("content"); content != "" {
		png, err := qrcode.Encode(content, qrcode.Medium, 256)

		if err != nil {
			log.Fatal(err)
		}

		c.Data(http.StatusOK, "image/png", png)
	} else {
		c.Status(http.StatusBadRequest)
	}
}

func FilesController(c *gin.Context) {
	// 获取前端上传文件
	file, err := c.FormFile("raw")
	if err != nil {
		log.Fatal(err)
	}

	exe, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}

	dir := filepath.Dir(exe)

	filename := uuid.New().String()
	uploads := filepath.Join(dir, "uploads")

	err = os.MkdirAll(uploads, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

	fullpath := path.Join("uploads", filename+filepath.Ext(file.Filename))
	fileErr := c.SaveUploadedFile(file, filepath.Join(dir, fullpath))

	if fileErr != nil {
		log.Fatal(err)
	}

	c.JSON(http.StatusOK, gin.H{"url": "/" + fullpath})
}
