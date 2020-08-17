package controller

import (
	"crypto/md5"
	"io/ioutil"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	myHelper "github.com/livegoplayer/go_helper"
	myLogger "github.com/livegoplayer/go_logger"
	"github.com/spf13/viper"

	"github.com/livegoplayer/filestore-server/fileStore"
	"github.com/livegoplayer/filestore-server/model"

	ginHelper "github.com/livegoplayer/go_gin_helper"
)

type UpLoadRequest struct {
	Uid    int `form:"uid" validate:"required"`
	PathId int `form:"path_id" validate:"number"`
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

func SaveUserPathHandler(c *gin.Context) {
	addPathRequest := &model.UserPath{}
	err := c.Bind(addPathRequest)
	ginHelper.CheckError(err, "参数校验错误")

	//保存文件到目录
	userPathInfo := fileStore.SaveUserFilePath(addPathRequest)

	ginHelper.SuccessResp("ok", userPathInfo)
}

type IdsRequest struct {
	IdMap []int `form:"id_map" validate:"required,dive,gt=0"`
	Uid   int   `form:"uid" validate:"required,gt=0"`
}

func BatchDelUserPathHandler(c *gin.Context) {
	delIdsRequest := &IdsRequest{}
	err := c.Bind(delIdsRequest)
	ginHelper.CheckError(err, "参数校验错误")

	//检索所有文件夹列表的子文件
	//第一步，获取该用户的所有文件夹列表,拿出来遍历
	userPathList := fileStore.GetUserPathList(delIdsRequest.Uid)
	idList := []int{}
	for _, id := range delIdsRequest.IdMap {
		idList = append(idList, fileStore.GetChildPathIdList(id, userPathList, []int{})...)
	}

	success := fileStore.DelUserPath(idList)

	data := make(map[string]interface{})

	data["success"] = success
	ginHelper.SuccessResp("删除文件夹成功", data)
}

type moveIdsRequest struct {
	IdMap    []int `form:"id_map" validate:"required,dive,gt=0"`
	Uid      int   `form:"uid" validate:"required,gt=0"`
	ParentId int   `form:"parent_id" validate:"required,gt=0"`
}

//批量操作用户path相关，不涉及子文件操作
func BatchMoveUserPathHandler(c *gin.Context) {
	moveIdsRequest := &moveIdsRequest{}
	err := c.Bind(moveIdsRequest)
	ginHelper.CheckError(err, "参数校验错误")
	//检查是否把父文件夹移动到了子文件夹或者当前文件夹
	userPathList := fileStore.GetUserPathList(moveIdsRequest.Uid)
	idList := []int{}
	for _, id := range moveIdsRequest.IdMap {
		idList = append(idList, fileStore.GetChildPathIdList(id, userPathList, []int{})...)
	}
	if exists, _ := myHelper.InArray(moveIdsRequest.ParentId, idList); exists {
		ginHelper.ErrorResp(1, "不能移动文件夹到其子文件夹")
	}

	//更新
	success := fileStore.UpdateUserPath(moveIdsRequest.IdMap, model.UserPath{ParentId: moveIdsRequest.ParentId})

	data := make(map[string]interface{})
	data["success"] = success
	ginHelper.SuccessResp("移动文件夹成功", data)
}

func BatchMoveUserFileHandler(c *gin.Context) {
	moveIdsRequest := &moveIdsRequest{}
	err := c.Bind(moveIdsRequest)
	ginHelper.CheckError(err, "参数校验错误")

	success := fileStore.UpdateUserFile(moveIdsRequest.IdMap, model.RetUserFile{PathId: moveIdsRequest.ParentId})

	data := make(map[string]interface{})
	data["success"] = success
	ginHelper.SuccessResp("移动文件成功", data)
}

func BatchDelUserFileHandler(c *gin.Context) {
	delIdsRequest := &IdsRequest{}
	err := c.Bind(delIdsRequest)
	ginHelper.CheckError(err, "参数校验错误")

	//保存文件到目录
	success := fileStore.DelUserFile(delIdsRequest.IdMap)

	data := make(map[string]interface{})

	data["success"] = success
	ginHelper.SuccessResp("删除文件成功", data)
}

type GetFileListRequest struct {
	Uid       int    `form:"uid" validate:"number,gt=0" json:"uid"`
	PathId    int    `form:"path_id" validate:"number" json:"path_id"`
	SearchKey string `form:"search_key" json:"search_key"`
}

func GetFileListHandler(c *gin.Context) {
	getFileListRequest := &GetFileListRequest{}
	err := c.BindQuery(getFileListRequest)
	ginHelper.CheckError(err, "参数校验错误")

	fileList := fileStore.GetFileListByPathId(getFileListRequest.Uid, getFileListRequest.PathId, getFileListRequest.SearchKey)
	data := make(map[string]interface{})

	data["file_list"] = fileList
	data["path_id"] = getFileListRequest.PathId
	ginHelper.SuccessResp("ok", data)
}

type GetUserPathList struct {
	Uid int `form:"uid" validate:"number,gt=0" json:"uid"`
}

func GetUserPathListHandler(c *gin.Context) {
	getUserPathList := &GetUserPathList{}
	err := c.BindQuery(getUserPathList)
	ginHelper.CheckError(err, "参数校验错误")

	pathList := fileStore.GetUserPathList(getUserPathList.Uid)
	data := make(map[string]interface{})

	data["path_list"] = pathList
	ginHelper.SuccessResp("ok", data)
}

type GetUserChildPathList struct {
	Uid       int    `form:"uid" validate:"number,gt=0" json:"uid"`
	Pid       int    `form:"parent_id" validate:"number" json:"parent_id"`
	SearchKey string `form:"search_key" json:"search_key"`
}

//获取子目录列表
func GetUserChildPathListHandler(c *gin.Context) {
	getUserChildPathList := &GetUserChildPathList{}
	err := c.BindQuery(getUserChildPathList)
	ginHelper.CheckError(err, "参数校验错误")

	pathList := fileStore.GetUserChildPathList(getUserChildPathList.Uid, getUserChildPathList.Pid, getUserChildPathList.SearchKey)
	data := make(map[string]interface{})

	data["path_list"] = pathList
	ginHelper.SuccessResp("ok", data)
}

//获取oss直传token
type GetOSSUploadTokeRequest struct {
	FileName string `form:"file_name" validate:"required" json:"file_name"`
	FileSha1 string `form:"file_sha1" validate:"required" json:"file_sha1"`
	Uid      int    `form:"uid" validate:"required" json:"uid"`
	PathId   int    `form:"path_id" json:"path_id"`
	FileSize int64  `form:"file_size" validate:"required" json:"file_size"`
}

//获取oss客户端直传的配置
func GetOSSUploadTokenHandler(c *gin.Context) {
	getOSSUploadTokeRequest := &GetOSSUploadTokeRequest{}
	err := c.Bind(getOSSUploadTokeRequest)
	ginHelper.CheckError(err, "参数校验错误")
	data := make(map[string]interface{})

	bucketName := viper.GetString("oss.bucketName")
	//获取保存的目录名
	pathToSave := filepath.ToSlash(fileStore.GetDefaultPath(getOSSUploadTokeRequest.FileName))

	//检查文件是否存在，如果已经存在
	isUpload := false
	file, exists := fileStore.CheckFileExists(getOSSUploadTokeRequest.FileSha1)
	if exists {
		//直接调用给用户添加的方法
		id := fileStore.AddExistOSSFileToUser(file.Id, getOSSUploadTokeRequest.FileName, getOSSUploadTokeRequest.Uid, getOSSUploadTokeRequest.PathId, file.Size)
		if id > 0 {
			isUpload = true
			data["token"] = ""
			data["is_upload"] = isUpload
			ginHelper.SuccessResp("ok", data)
		} else {
			ginHelper.ErrorResp(1, "上传失败")
		}
	}

	callbackParam := fileStore.CallbackParam{
		CallbackUrl: viper.GetString("app_host") + "/api/file/ossUploadSuccessCallback",
	}

	md5Time := md5.Sum([]byte(strconv.Itoa(int(time.Now().Unix()))))
	fileExt := myHelper.GetFileExtName(getOSSUploadTokeRequest.FileName)
	realFileName := myHelper.Substring(getOSSUploadTokeRequest.FileName, 0, strings.LastIndex(getOSSUploadTokeRequest.FileName, "."))
	fileSsoName := realFileName + string(md5Time[:]) + "." + fileExt

	v := url.Values{}
	v.Add("bucket_name", bucketName)
	v.Add("file_sso_name", fileSsoName)
	v.Add("file_name", getOSSUploadTokeRequest.FileName)
	v.Add("file_sha1", getOSSUploadTokeRequest.FileSha1)
	v.Add("uid", strconv.Itoa(getOSSUploadTokeRequest.Uid))
	v.Add("path_id", strconv.Itoa(getOSSUploadTokeRequest.PathId))
	v.Add("file_size", strconv.Itoa(int(getOSSUploadTokeRequest.FileSize)))
	v.Add("file_path", pathToSave)
	//v := &OSSUploadSuccessCallbackHandlerRequest{
	//	BucketName:  bucketName,
	//	FileOSSName: fileSsoName,
	//	FileName:    getOSSUploadTokeRequest.FileName,
	//	FileSha1:    getOSSUploadTokeRequest.FileSha1,
	//	Uid:         getOSSUploadTokeRequest.Uid,
	//	PathId:      getOSSUploadTokeRequest.PathId,
	//	FileSize:    getOSSUploadTokeRequest.FileSize,
	//	FileOSSPath: pathToSave,
	//}

	//callbackParam.CallbackBody = "filename=${object}&size=${size}&mimeType=${mimeType}&height=${imageInfo.height}&width=${imageInfo.width}"
	callbackParam.CallbackBody = v.Encode()
	//callbackBody, err := json.Marshal(v)
	//callbackParam.CallbackBody = string(callbackBody)

	token := fileStore.GetPolicyToken(int64(time.Minute*5/time.Millisecond), pathToSave, callbackParam, viper.GetString("oss.bucketName"))

	data["token"] = token
	data["is_upload"] = isUpload
	ginHelper.SuccessResp("ok", data)
}

type OSSUploadSuccessCallbackHandlerRequest struct {
	BucketName  string `form:"bucket_name" validate:"required" json:"bucket_name"`
	FileOSSName string `form:"file_sso_name" validate:"required" json:"file_sso_name"`
	FileName    string `form:"file_name" validate:"required" json:"file_name"`
	FileOSSPath string `form:"file_path" validate:"required" json:"file_path"`
	FileSha1    string `form:"file_sha1" validate:"required" json:"file_sha1"`
	Uid         int    `form:"uid" validate:"required" json:"uid"`
	PathId      int    `form:"path_id" validate:"required" json:"path_id"`
	FileSize    int64  `form:"file_size" validate:"required" json:"file_size"`
}

//处理oss上传成功 由oss服务器回调函数
func OSSUploadSuccessCallbackHandler(c *gin.Context) {
	// Get PublicKey bytes
	bytePublicKey, err := fileStore.GetPublicKey(c.Request)
	ginHelper.CheckError(err)

	// Get Authorization bytes : decode from Base64String
	byteAuthorization, err := fileStore.GetAuthorization(c.Request)
	ginHelper.CheckError(err)

	// Get MD5 bytes from Newly Constructed Authrization String.
	byteMD5, err := fileStore.GetMD5FromNewAuthString(c.Request)
	ginHelper.CheckError(err)

	// verifySignature and response to client
	content, _ := ioutil.ReadAll(c.Request.Body)
	myLogger.Info(content)
	if fileStore.VerifySignature(bytePublicKey, byteMD5, byteAuthorization) {
		// 这里存放callback代码
		request := &OSSUploadSuccessCallbackHandlerRequest{}
		myLogger.Info(c.Request.Header.Get("content-type"))
		if c.Request.Header.Get("content-type") == "application/json" || c.Request.Header.Get("content-type") == "application/json" {
			err := c.Bind(request)
			ginHelper.CheckError(err)
		} else if c.Request.Header.Get("content-type") == "text/plain" {
			content, _ := ioutil.ReadAll(c.Request.Body)
			urlMap, err := url.ParseQuery(string(content))
			ginHelper.CheckError(err)
			request.BucketName = strings.Join(urlMap["bucket_name"], "")
			request.FileOSSName = strings.Join(urlMap["file_sso_name"], "")
			request.FileName = strings.Join(urlMap["file_name"], "")
			request.FileOSSPath = strings.Join(urlMap["file_path"], "")
			request.FileSha1 = strings.Join(urlMap["file_sha1"], "")
			request.Uid, _ = strconv.Atoi(strings.Join(urlMap["uid"], ""))
			request.PathId, _ = strconv.Atoi(strings.Join(urlMap["path_id"], ""))
			fileSize, _ := strconv.Atoi(strings.Join(urlMap["file_size"], ""))
			request.FileSize = int64(fileSize)
		}
		myLogger.Info(*request)

		id := fileStore.AddOSSFileToUser(request.BucketName, request.FileOSSName, request.FileName, request.FileOSSPath, request.FileSha1, request.Uid, request.PathId, request.FileSize)

		if id > 0 {
			data := make(map[string]interface{})
			data["new_id"] = id
			ginHelper.SuccessResp("ok", data) // response OK : 200
		} else {
			ginHelper.ErrorResp(1, "保存失败")
		}
	} else {
		ginHelper.ErrorResp(1, "验证失败") // response FAILED : 400
	}
}
