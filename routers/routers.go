package routers

import (
	"github.com/gin-gonic/gin"

	. "github.com/livegoplayer/filestore-server/controller"
)

func InitAppRouter(r gin.IRoutes) {
	// 设置一个get请求的路由，url为/ping, 处理函数（或者叫控制器函数）是一个闭包函数。
	r.POST("/api/file/upload", UpLoadHandler)
	r.GET("/api/file/test", TestHandler)

	r.POST("/api/user/checkToken", CheckTokenHandler)

	//获取文件列表
	r.GET("/api/file/getFileList", GetFileListHandler)
	r.GET("/api/file/getPathList", GetUserPathListHandler)
	r.GET("/api/file/getChildPathList", GetUserChildPathListHandler)
	r.POST("/api/file/saveUserPath", SaveUserPathHandler)
	r.POST("/api/file/batchDelUserPath", BatchDelUserPathHandler)
	r.POST("/api/file/batchDelUserFile", BatchDelUserFileHandler)
	r.POST("/api/file/batchMoveUserFile", BatchMoveUserFileHandler)
	r.POST("/api/file/batchMoveUserPath", BatchMoveUserPathHandler)

	//oss客户端直传相关逻辑
	r.POST("/api/file/getUploadToken", GetOSSUploadTokenHandler)
	r.POST("/api/file/ossUploadSuccessCallback", OSSUploadSuccessCallbackHandler)
}
