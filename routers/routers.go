package routers

import (
	"github.com/gin-gonic/gin"
	ginHelper "github.com/livegoplayer/go_gin_helper"

	. "github.com/livegoplayer/filestore-server/controller"
)

func InitAppRouter(r *gin.Engine) {

	userGroup := r.Group("/api/user")
	{
		userGroup.POST("/checkToken", CheckTokenHandler)
		userGroup.GET("/logout", LogoutHandler)
	}

	//各种中间件调用顺序不能变
	//根据需求开关验证逻辑，如果需要postman测试 接口的话，建议关闭此选项
	fileGroup := r.Group("/api/file", ginHelper.AuthenticationMiddleware(CommonCheckTokenHandler))
	{
		// 设置一个get请求的路由，url为/ping, 处理函数（或者叫控制器函数）是一个闭包函数。
		fileGroup.POST("/upload", UpLoadHandler)
		fileGroup.GET("/test", TestHandler)

		//获取文件列表
		fileGroup.GET("/getFileList", GetFileListHandler)
		fileGroup.GET("/getPathList", GetUserPathListHandler)
		fileGroup.GET("/getChildPathList", GetUserChildPathListHandler)
		fileGroup.POST("/saveUserPath", SaveUserPathHandler)
		fileGroup.POST("/batchDelUserPath", BatchDelUserPathHandler)
		fileGroup.POST("/batchDelUserFile", BatchDelUserFileHandler)
		fileGroup.POST("/batchMoveUserFile", BatchMoveUserFileHandler)
		fileGroup.POST("/batchMoveUserPath", BatchMoveUserPathHandler)

		//oss客户端直传相关逻辑
		fileGroup.POST("/getUploadToken", GetOSSUploadTokenHandler)
		fileGroup.POST("/ossUploadSuccessCallback", OSSUploadSuccessCallbackHandler)
	}
}
