package controller

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"awesomeProject/fileStore"
	"awesomeProject/helper"
)

func UpLoadHandler(c *gin.Context) {
	//如果是POST请求
	//上传接口,接收文件信息流
	file, fileHeader, err := c.Request.FormFile("file")
	helper.CheckError(err, "获取文件信息失败")

	//保存文件到目录
	_, err = fileStore.AddFileToUser(fileHeader, fileHeader.Filename, fileStore.DEFAULT_PATH)
	defer file.Close()
	helper.CheckError(err, "保存文件信息失败")

	helper.SuccessResp(c, "ok")
}

func UploadSuccessHandler(response http.ResponseWriter, request *http.Request) {
	fmt.Print("file upload succeed !")
}
