package model

import (
	"github.com/livegoplsyer/filestore-server/dbHelper"
)

type File struct {
	Id             int
	FileSha1       string `gorm:"column:file_sha1"`
	Path           string `gorm:"column:path"`
	Size           int64  `gorm:"column:file_size"`
	AddDatetime    int64  `gorm:"column:add_datetime;-"`
	UpdateDatetime int64  `gorm:"column:update_datetime;-"`
}

//设置表名，可以通过给struct类型定义 TableName函数，返回当前struct绑定的mysql表名是什么
func (u *File) TableName() string {
	//绑定MYSQL表名为users
	return prefix + "file"
}

type RetUserFile struct {
	Model
	ID       int    //会被自动认为是主键
	UserId   int    `gorm:"column:uid"`
	FileName string `gorm:"column:filename"`
	FileId   int    `gorm:"column:file_id"`
}

//设置表名，可以通过给struct类型定义 TableName函数，返回当前struct绑定的mysql表名是什么
func (u *RetUserFile) TableName() string {
	//绑定MYSQL表名为users
	return prefix + "ret_user_file"
}

//检查文件是否存存在的方法
func CheckFileExist(fileSha1 string) (*File, bool) {
	db := dbHelper.GetDB()

	nFileModel := &File{}

	resObj := db.Where("file_sha1 = ?", fileSha1).Find(nFileModel)
	if resObj.RecordNotFound() {
		return &File{FileSha1: fileSha1}, false
	}

	return nFileModel, true
}

//todo 这个函数将来需要加入retuserfile对象
func SaveFileToMysql(fileSha1 string, path string, fileSize int64) error {

	db := dbHelper.GetDB()

	nFileModel := GetFileModelByFileMeta(fileSha1, path, fileSize)

	if err := db.Create(nFileModel).Error; err != nil {
		return err
	}

	return nil
}

//根据两个实例拼接真正的FileMeta
func GetFileModelByFileMeta(fileSha1 string, path string, fileSize int64) *File {
	nFile := &File{}

	nFile.FileSha1 = fileSha1
	nFile.Path = path
	nFile.Size = fileSize

	return nFile
}
