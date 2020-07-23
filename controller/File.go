package controller

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/livegoplayer/filestore-server/fileStore"

	ginHelper "github.com/livegoplayer/go_gin_helper"
)

type UpLoadRequest struct {
	Uid    int `form:"uid" validate:"required"`
	PathId int `form:"path_id" validate:"required"`
}

func UpLoadHandler(c *gin.Context) {
	//如果是POST请求
	//上传接口,接收文件信息流
	file, fileHeader, err := c.Request.FormFile("file")
	ginHelper.CheckError(err, "获取文件信息失败")
	upLoadRequest := &UpLoadRequest{}
	err = c.Bind(upLoadRequest)
	ginHelper.CheckError(err, "参数校验错误")

	//保存文件到目录
	_, err = fileStore.AddFileToUser(fileHeader, fileHeader.Filename, fileStore.DEFAULT_PATH, upLoadRequest.Uid, upLoadRequest.PathId)
	defer file.Close()
	ginHelper.CheckError(err, "保存文件信息失败")

	ginHelper.SuccessResp("ok")

}

type GetFileListRequest struct {
	Uid    int `form:"uid" validate:"required, number"`
	PathId int `form:"path_id" validate:"required, number"`
}

func GetFileListHandler(c *gin.Context) {
	getFileListRequest := &GetFileListRequest{}
	err := c.Bind(getFileListRequest)
	ginHelper.CheckError(err, "参数校验错误")

	fileList := fileStore.GetFileListByPathId(getFileListRequest.Uid, getFileListRequest.PathId)
	data := make(map[string]interface{})

	data["file_list"] = fileList
	ginHelper.SuccessResp("ok", data)
}

func UploadSuccessHandler(response http.ResponseWriter, request *http.Request) {
	fmt.Print("file upload succeed !")
}
