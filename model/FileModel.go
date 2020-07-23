package model

import (
	dbHelper "github.com/livegoplayer/go_db_helper"
	myHelper "github.com/livegoplayer/go_helper"
)

type File struct {
	Model
	Id       int
	FileSha1 string `gorm:"column:file_sha1"`
	Path     string `gorm:"column:path"`
	Size     int64  `gorm:"column:file_size"`
}

//设置表名，可以通过给struct类型定义 TableName函数，返回当前struct绑定的mysql表名是什么
func (u *File) TableName() string {
	//绑定MYSQL表名为users
	return prefix + "file"
}

type RetUserFile struct {
	Model
	ID       int    //会被自动认为是主键
	UserId   int    `gorm:"column:uid" json:"user_id"`
	FileName string `gorm:"column:filename" json:"file_name"`
	FileId   int    `gorm:"column:file_id" json:"file_id"`
	Type     string `gorm:"column:type" json:"type"`
	FileSize int64  `gorm:"column:type" json:"size"`
	PathId   int    `gorm:"column:path_id" json:"path_id"`
}

//设置表名，可以通过给struct类型定义 TableName函数，返回当前struct绑定的mysql表名是什么
func (u *RetUserFile) TableName() string {
	//绑定MYSQL表名为users
	return prefix + "ret_user_file"
}

type UserPath struct {
	Model
	ID       int    //会被自动认为是主键
	UserId   int    `gorm:"column:uid" json:"user_id"`
	Type     int    `gorm:"column:type" json:"type"`
	PathName string `gorm:"column:path_name" json:"path_name"`
	ParentId int    `gorm:"column:parent_id" json:"parent_id"`
}

//设置表名，可以通过给struct类型定义 TableName函数，返回当前struct绑定的mysql表名是什么
func (u *UserPath) TableName() string {
	//绑定MYSQL表名为users
	return prefix + "user_path"
}

//检查文件是否存存在的方法
func CheckFileExist(fileSha1 string) (*File, bool) {
	db := dbHelper.GetDB()

	nFileModel := &File{}

	resObj := db.Where("file_sha1 = ?", fileSha1).Find(nFileModel)
	if resObj.RecordNotFound() {
		return nil, false
	}

	return nFileModel, true
}

func SaveFileToMysql(fileSha1 string, path string, fileSize int64) int {

	db := dbHelper.GetDB()

	nFileModel := GetFileModelByFileMeta(fileSha1, path, fileSize)

	if err := db.Create(nFileModel).Error; err != nil {
		panic(err)
	}

	return nFileModel.Id
}

func SaveFileToUser(fileId int, fileName string, uid int, pathId int, fileSize int64) int {
	db := dbHelper.GetDB()

	fileType := myHelper.GetFileExtName(fileName)
	RetUserFileModel := &RetUserFile{UserId: uid, FileId: fileId, FileName: fileName, PathId: pathId, Type: fileType, FileSize: fileSize}

	if err := db.Create(RetUserFileModel).Error; err != nil {
		panic(err)
	}

	return RetUserFileModel.ID
}

func GetFileListByPath(uid int, pathId int) []RetUserFile {
	db := dbHelper.GetDB()

	var fileList []RetUserFile
	if err := db.Model(&RetUserFile{}).Where("uid = ? and path_id =?", uid, pathId).Find(&fileList).Error; err != nil {
		panic(err)
	}

	return fileList
}

func GetUserPathList(uid int) []UserPath {
	db := dbHelper.GetDB()

	var pathList []UserPath
	if err := db.Model(&UserPath{}).Where("uid = ?", uid).Find(&pathList).Error; err != nil {
		panic(err)
	}

	return pathList
}

//根据两个实例拼接真正的FileMeta
func GetFileModelByFileMeta(fileSha1 string, path string, fileSize int64) *File {
	nFile := &File{}

	nFile.FileSha1 = fileSha1
	nFile.Path = path
	nFile.Size = fileSize

	return nFile
}
