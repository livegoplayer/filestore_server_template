package model

import (
	dbHelper "github.com/livegoplayer/go_db_helper"
)

type File struct {
	Model
	Id         int
	FileSha1   string `gorm:"column:file_sha1"`
	Path       string `gorm:"column:path"`
	StoreType  int    `gorm:"column:store_type"`
	BucketName string `gorm:"column:bucket_name"`
	Size       int64  `gorm:"column:file_size"`
}

//设置表名，可以通过给struct类型定义 TableName函数，返回当前struct绑定的mysql表名是什么
func (u *File) TableName() string {
	//绑定MYSQL表名为users
	return prefix + "file"
}

type RetUserFile struct {
	Model
	ID       int    `gorm:"column:id" json:"id"` //会被自动认为是主键
	UserId   int    `gorm:"column:uid" json:"user_id"`
	FileName string `gorm:"column:filename" json:"file_name"`
	FileId   int    `gorm:"column:file_id" json:"file_id"`
	Type     int    `gorm:"column:type" json:"type"`
	FileSize int64  `gorm:"column:size" json:"size"`
	Status   int64  `gorm:"column:status" json:"status"`
	PathId   int    `gorm:"column:path_id" json:"path_id"`
}

//设置表名，可以通过给struct类型定义 TableName函数，返回当前struct绑定的mysql表名是什么
func (u *RetUserFile) TableName() string {
	//绑定MYSQL表名为users
	return prefix + "ret_user_file"
}

type UserPath struct {
	Model
	ID       int    `gorm:"column:id" json:"id" form:"path_id"` //会被自动认为是主键
	UserId   int    `gorm:"column:uid" json:"user_id" form:"uid" validate:"required,number,gt=0"`
	Status   int    `gorm:"column:status" json:"status" form:"status"`
	PathName string `gorm:"column:path_name" json:"path_name" form:"path_name" validate:"required"`
	ParentId int    `gorm:"column:parent_id" json:"parent_id" form:"parent_id"`
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

func AddUserPath(userPath UserPath) UserPath {
	db := dbHelper.GetDB()

	if err := db.Create(&userPath).Error; err != nil {
		panic(err)
	}

	return userPath
}

func UpdateUserPath(userPath UserPath, idMap []int) UserPath {
	db := dbHelper.GetDB()

	if len(idMap) > 0 {
		db = db.Model(&UserPath{}).Where("id in (?)", idMap)
	}

	err := db.Update(userPath).Error
	if err != nil {
		panic(err)
	}

	//对于单个操作，这个很有意义
	return userPath
}

func UpdateUserFile(retUserFile RetUserFile, idMap []int) RetUserFile {
	db := dbHelper.GetDB()

	if len(idMap) > 0 {
		db = db.Model(&RetUserFile{}).Where("id in (?)", idMap)
	}

	err := db.Update(retUserFile).Error
	if err != nil {
		panic(err)
	}

	return retUserFile
}

//删除文件夹下的所有文件
func DelFilesInPath(idMap []int) bool {
	db := dbHelper.GetDB()
	if len(idMap) > 0 {
		if err := db.Model(&RetUserFile{}).Where("path_id in (?)", idMap).Updates(&RetUserFile{Status: 9}).Error; err != nil {
			panic(err)
		}
	}

	return true
}

const (
	LocalStore = 1
	OSSStore   = 1
)

func SaveFileToMysql(fileSha1 string, path string, fileSize int64, fileStoreType int, bucketName string) int {

	db := dbHelper.GetDB()

	nFileModel := GetFileModelByFileMeta(fileSha1, path, fileSize, fileStoreType, bucketName)

	if err := db.Create(nFileModel).Error; err != nil {
		panic(err)
	}

	return nFileModel.Id
}

func SaveFileToUser(fileId int, fileName string, uid int, pathId int, fileSize int64, fileType int) int {
	db := dbHelper.GetDB()

	RetUserFileModel := &RetUserFile{UserId: uid, FileId: fileId, FileName: fileName, PathId: pathId, Type: fileType, FileSize: fileSize}

	if err := db.Create(RetUserFileModel).Error; err != nil {
		panic(err)
	}

	return RetUserFileModel.ID
}

func GetFileListByPath(uid int, pathId int, searchKey string) []RetUserFile {
	db := dbHelper.GetDB()

	db = db.Model(&RetUserFile{}).Where("uid = ? and path_id =? and status != 9", uid, pathId)

	if searchKey != "" {
		db = db.Where("filename like ?", "%"+searchKey+"%")
	}

	var fileList []RetUserFile
	if err := db.Find(&fileList).Error; err != nil {
		panic(err)
	}

	return fileList
}

func GetUserPathList(uid int) []UserPath {
	db := dbHelper.GetDB()

	var pathList []UserPath
	if err := db.Model(&UserPath{}).Where("uid = ? and status != 9", uid).Find(&pathList).Error; err != nil {
		panic(err)
	}

	return pathList
}

func GetUserChildPathList(uid int, pid int, searchKey string) []UserPath {
	db := dbHelper.GetDB()

	db = db.Model(&UserPath{}).Where("uid = ? and parent_id = ? and status != 9", uid, pid)

	if searchKey != "" {
		db = db.Where("path_name like ?", "%"+searchKey+"%")
	}

	var pathList []UserPath
	if err := db.Find(&pathList).Error; err != nil {
		panic(err)
	}

	return pathList
}

//根据两个实例拼接真正的FileMeta
func GetFileModelByFileMeta(fileSha1 string, path string, fileSize int64, fileStoreType int, bucketName string) *File {
	nFile := &File{}

	nFile.FileSha1 = fileSha1
	nFile.Path = path
	nFile.Size = fileSize
	nFile.StoreType = fileStoreType
	nFile.BucketName = bucketName

	return nFile
}

func GetFileByUserFileId(id int) *File {
	db := dbHelper.GetDB()

	userFile := &RetUserFile{}
	file := &File{}

	if err := db.Model(&RetUserFile{}).Where("id = ?", id).Find(&userFile).Error; err != nil {
		panic(err)
	}

	if userFile.FileId != 0 {
		if err := db.Model(&File{}).Where("id = ?", userFile.FileId).Find(&file).Error; err != nil {
			panic(err)
		}
	}

	return file
}
