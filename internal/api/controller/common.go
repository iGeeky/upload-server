package controller

import (
	"fmt"
	"io"
	"time"

	"github.com/iGeeky/open-account/pkg/baselib/ginplus"
	"github.com/iGeeky/open-account/pkg/baselib/utils"
	"github.com/iGeeky/upload-server/internal/api/dao"

	"github.com/gin-gonic/gin"
)

// UContextPlus 扩展的上下文.
type UContextPlus struct {
	*ginplus.ContextPlus
}

// JSONFail 返回出错信息
func (c *UContextPlus) JSONFail(status int, reason string) {
	response := gin.H{"ok": false, "reason": reason}
	c.JSON(status, response)
}

// JSONOK 返回成功消息.
func (c *UContextPlus) JSONOK(data gin.H) {
	response := gin.H{"ok": true, "reason": "", "data": data}
	c.JSON(200, response)
}

// NewUContextPlus 创建一个扩展上下文.
func NewUContextPlus(c *gin.Context) (context *UContextPlus) {
	context = &UContextPlus{
		ContextPlus: ginplus.NewContetPlus(c),
	}
	return
}

func createFilename(appID, hash, prefix, suffix string) (filename string) {
	filename = fmt.Sprintf("%s/%s/%s/%s", prefix, hash[0:2], hash[2:4], hash)
	if suffix != "" {
		filename = filename + "." + suffix
	}
	return
}

// APIPing ping测试接口
func APIPing(c *gin.Context) {
	ctx := NewUContextPlus(c)
	data := gin.H{"buildTime": utils.BuildTime, "GitBranch": utils.GitBranch, "GitCommit": utils.GitCommit, "now": utils.Datetime()}
	ctx.JsonOk(data)
}

func saveFileInfo(info *dao.UploadFile) {
	uploadFileDao := dao.NewUploadFileDao()
	defer utils.Elapsed("SaveFileinfo:" + info.Hash)()
	uploadFileDao.Upsert(info)
}

func SaveFileToCloud(appID, rid_, hash, filename, contentType string, bodyReader io.Reader, bodyLen int64,
	width, height, duration int, extInfo string) (url string, errmsg string) {
	defer utils.Elapsed("ToOSS&ToDB:" + hash)()
	url, errmsg = saveToCloud(appID, hash, filename, contentType, bodyReader, bodyLen)

	//保存文件信息。
	now := uint32(time.Now().Unix())
	info := &dao.UploadFile{}
	info.Rid = rid_
	info.AppId = appID
	info.Hash = hash
	info.Size = bodyLen
	info.Path = url
	info.Width = width
	info.Height = height
	info.Duration = duration
	info.Extinfo = extInfo
	info.CreateTime = now
	info.UpdateTime = now

	saveFileInfo(info)

	return
}
