package controller

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/livegoplsyer/filestore-server/fileStore"

	ginHelper "github.com/livegoplayer/go_gin_helper"
)

func UpLoadHandler(c *gin.Context) {
	//如果是POST请求
	//上传接口,接收文件信息流
	file, fileHeader, err := c.Request.FormFile("file")
	ginHelper.CheckError(err, "获取文件信息失败")

	//保存文件到目录
	_, err = fileStore.AddFileToUser(fileHeader, fileHeader.Filename, fileStore.DEFAULT_PATH)
	defer file.Close()
	ginHelper.CheckError(err, "保存文件信息失败")

	ginHelper.SuccessResp("ok")
}

func UploadSuccessHandler(response http.ResponseWriter, request *http.Request) {
	fmt.Print("file upload succeed !")
}
