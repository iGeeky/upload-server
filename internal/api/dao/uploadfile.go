package dao

import (
	"github.com/iGeeky/open-account/pkg/baselib/db"
	"github.com/iGeeky/upload-server/configs"
	// _ "github.com/lib/pq"
)

type UploadFile struct {
	Rid        string `gorm:"primary_key;type:varchar(64)" json:"rid"`
	AppId      string `gorm:"not null;type:varchar(32)" json:"appID"`
	Hash       string `gorm:"not null;type:varchar(64)" json:"hash"`
	Size       int64  `gorm:"" json:"size"`
	Path       string `gorm:"not null;type:varchar(256)" json:"path"`
	Width      int    `gorm:"" json:"width"`
	Height     int    `gorm:"" json:"height"`
	Duration   int    `gorm:"" json:"duration"`
	Extinfo    string `gorm:"" json:"extinfo"`
	CreateTime uint32 `gorm:"not null" json:"createTime"`
	UpdateTime uint32 `gorm:"not null" json:"updateTime"`
}

func (UploadFile) TableName() string {
	return "uploadfile"
}

type UploadFileDao struct {
	*db.BaseDao
}

func NewUploadFileDao() (dao *UploadFileDao) {
	dao = &UploadFileDao{BaseDao: db.NewBaseDao(configs.Config.UploadDB.Dialect,
		configs.Config.UploadDB.ToURL(), configs.Config.UploadDB.Debug, &UploadFile{})}
	return
}

func (u *UploadFileDao) GetInfoByID(rid string) (uploadFile *UploadFile, err error) {
	obj, err := u.GetSimple("rid=?", rid)
	if obj != nil {
		uploadFile = obj.(*UploadFile)
	}
	return
}

func (u *UploadFileDao) GetInfoByHash(hash, appID string) (uploadFile *UploadFile, err error) {
	obj, err := u.GetSimple("hash=? and app_id=?", hash, appID)
	if obj != nil {
		uploadFile = obj.(*UploadFile)
	}
	return
}

func (u *UploadFileDao) MustGetInfoByID(rid string) (uploadFile *UploadFile) {
	obj := u.MustGet("rid=?", rid)
	if obj != nil {
		uploadFile = obj.(*UploadFile)
	}
	return
}

func (u *UploadFileDao) GetInfoByURL(url, appID string) (uploadFile *UploadFile, err error) {
	obj, err := u.GetSimple("path = ? and app_id=?", url, appID)
	if obj != nil {
		uploadFile = obj.(*UploadFile)
	}
	return
}
