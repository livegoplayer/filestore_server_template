package fileStore

import (
	"time"
)

//文件信息结构
type FileMeta struct {
	FileName   string
	FileSha1   string
	FileSize   int64 //复制的字节数，主要是io.copy返回的第一个参数
	Location   string
	UploadTime time.Time
	UpdateTime time.Time
}
