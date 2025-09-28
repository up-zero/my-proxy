package util

import (
	"github.com/gin-gonic/gin"
	"io"
)

// FormFileReadAll 读取表单文件内容
func FormFileReadAll(c *gin.Context, fileName string) ([]byte, error) {
	file, err := c.FormFile(fileName)
	if err != nil {
		return nil, err
	}
	src, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()
	fileData, err := io.ReadAll(src)
	if err != nil {
		return nil, err
	}
	return fileData, nil
}
