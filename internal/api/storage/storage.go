package storage

import (
	"io"
	"strings"
)

// Storage 存储类接口
type Storage interface {
	PutFile(key string, reader io.Reader, objectSize int64, contentType string) (err error)
	DelFile(key string) (err error)
	FileExist(key string) (exist bool, err error)
	RenameFile(key string, newKey string) (err error)
}

// New 根据配置创建存储实现
func New(EndPoint, AccessID, AccessKey, Bucket string) (storage Storage, err error) {
	if strings.Contains(EndPoint, "aliyuncs.com") {
		storage, err = NewOSS(EndPoint, AccessID, AccessKey, Bucket)
	} else {
		storage, err = NewS3(EndPoint, AccessID, AccessKey, Bucket)
	}
	return
}
