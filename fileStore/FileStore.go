package fileStore

import (
	"io"
	"mime/multipart"
	"os"
	"path"
	"time"

	myHelper "github.com/livegoplayer/go_helper"

	"github.com/livegoplayer/filestore-server/model"
)

//默认值
var DEFAULT_PATH string

func init() {
	DEFAULT_PATH = ""
}

//这样定义比较好
var (
	err error
)

//根据二进制文件保存文件
func SaveFileToDir(file multipart.File, newFileName string, toPath string, fileSha1 string) (*FileMeta, error) {
	//解析文件后缀，分别放到不同的文件夹
	var fileMeta = &FileMeta{}
	//如果选择使用默认路径，默认都是默认路径
	if toPath == DEFAULT_PATH {
		toPath, err = getDefaultPath(newFileName)
		if err != nil {
			return fileMeta, err
		}
	}

	//创建新文件
	createdFile, err := os.Create(path.Join(toPath, "/", newFileName))
	if err != nil {
		return fileMeta, err
	}
	defer createdFile.Close()

	//复制文件内容到新文件
	fileSize, err := io.Copy(createdFile, file)
	//复制文件
	if err != nil {
		return fileMeta, err
	}

	//初始化文件元信息
	fileMeta.FileName = newFileName
	fileMeta.FileSize = fileSize
	//文件指针重置
	_, _ = createdFile.Seek(0, 0)
	fileMeta.FileSha1 = myHelper.FileSha1(createdFile)
	fileMeta.Location = toPath + newFileName
	fileMeta.Type = myHelper.GetFileExtName(newFileName)
	fileMeta.UploadTime = time.Now()
	fileMeta.UpdateTime = time.Now()
	fileMeta.FileSha1 = fileSha1

	return fileMeta, nil
}

//为用户增加一个文件
func AddFileToUser(fileHeader *multipart.FileHeader, newFileName string, toPath string, uid int, pathId int) (fileMeta *FileMeta, err error) {
	fileMeta = nil
	err = nil

	file, err := myHelper.GetFileByHeader(fileHeader)
	if err != nil {
		return
	}
	fileSha1 := myHelper.FileSha1(file)

	//如果该文件存在
	if fileModel, exist := model.CheckFileExist(fileSha1); exist {
		fileMeta = GetFileMetaByFile(fileModel)
	} else {
		//重新获取file对象，因为file被sha1方法破坏了
		file, err = myHelper.GetFileByHeader(fileHeader)
		if err != nil {
			return
		}
		fileMeta, err = SaveFileToDir(file, newFileName, toPath, fileSha1)
		if err != nil {
			return
		}
		fileId := model.SaveFileToMysql(fileMeta.FileSha1, fileMeta.Location, fileMeta.FileSize)
		if fileId > 0 {
			_ = model.SaveFileToUser(fileId, newFileName, uid, pathId, fileMeta.FileSize)
		}
	}

	return fileMeta, err
}

//根据文件后缀名获取默认存储路径
func getDefaultPath(fileName string) (string, error) {
	ext := myHelper.GetFileExtName(fileName)
	var Path string
	if ext == "" {
		Path = path.Join("./files/", "unknown", "/")
	} else {
		Path = path.Join("./files/", ext, "/")
	}

	defaultSavePath := myHelper.PathToCommon(Path)

	//确保文件夹已经存在
	err := os.MkdirAll(defaultSavePath, 0666)
	//如果创建出错
	if err != nil {
		return "", err
	}

	return defaultSavePath, nil
}

func GetFileListByPathId(uid int, pathId int) []model.RetUserFile {
	return model.GetFileListByPath(uid, pathId)
}

func GetUserPathList(uid int) []model.UserPath {
	return model.GetUserPathList(uid)
}

//根据file获取初始化好的FileMeta对象 todo 增加user file对象
func GetFileMetaByFile(file *model.File) *FileMeta {
	fileMeta := &FileMeta{}
	fileMeta.FileSha1 = file.FileSha1
	fileMeta.FileSize = file.Size
	fileMeta.Location = file.Path
	fileMeta.Type = myHelper.GetFileExtName(file.Path)
	_, fileMeta.FileName = path.Split(file.Path)
	fileMeta.UpdateTime = myHelper.Str2Time(file.UpdateDatetime)
	fileMeta.UploadTime = myHelper.Str2Time(file.AddDatetime)

	return fileMeta
}
