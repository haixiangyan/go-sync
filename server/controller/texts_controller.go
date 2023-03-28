package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
)

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
