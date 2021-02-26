package controller

import (
	"bytes"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/iGeeky/open-account/pkg/baselib/errors"
	"github.com/iGeeky/open-account/pkg/baselib/log"
	coreutils "github.com/iGeeky/open-account/pkg/baselib/utils"
	"github.com/iGeeky/upload-server/configs"
	"github.com/iGeeky/upload-server/internal/api/dao"
	"github.com/iGeeky/upload-server/internal/api/utils"
	usutils "github.com/iGeeky/upload-server/pkg/utils"
)

//去掉Content-Type后的字符编码。
func contentTypeTrim(contentType string) string {
	arr := strings.SplitN(contentType, ";", 2)
	if len(arr) > 0 {
		return arr[0]
	}
	return contentType
}

// UploadSimple 直接通过http body上传文件.
func UploadSimple(c *gin.Context) {
	ctx := NewUContetPlus(c)

	// 参数解析.
	appID := ctx.MustGet("appID").(string)
	body, _ := ctx.GetBody()

	hash := ctx.GetCustomHeader("Hash")
	contentType := ctx.GetHeader("Content-Type")
	strResourceType := strings.TrimSpace(ctx.GetCustomHeader("Type"))
	rid := ctx.GetCustomHeader("Id")

	// 目标路径.
	target := strings.TrimSpace(ctx.GetCustomHeader("Target"))
	isTest := ctx.GetCustomHeader("Test") == "1"
	contentType = contentTypeTrim(contentType)
	quality := configs.Config.ImageQuality

	// 图片处理参数(参数格式同阿里云图片处理参数兼容)
	// https://help.aliyun.com/document_detail/171050.html?spm=a2c4g.11186623.6.650.438b12ff4jFECe
	imageProcess := ctx.GetCustomHeader("ImageProcess")
	targetQuality, crop := utils.ParseImageProcessParams(imageProcess)
	if targetQuality > 0 {
		quality = targetQuality
	}

	bodyReader := bytes.NewReader(body)
	bodyLen := int64(len(body))
	fileType, suffix := usutils.GetFileTypeAndSuffix(contentType)
	if fileType == "" {
		ctx.JSONFail(400, suffix)
		return
	}
	calcHash := coreutils.Sha1hex(body)
	if hash != calcHash {
		log.Errorf("request hash is %s, calcHash is %s", hash, calcHash)
		ctx.JSONFail(400, utils.ErrHashInvalid)
		return
	}

	width := 0
	height := 0
	if fileType == "img" {
		size, err := utils.GetImageSize(body)
		if err != nil {
			log.Errorf("Get Image[%s]'s size failed! err: %v", c.Request.RequestURI, err)
		} else {
			width = size.Width
			height = size.Height
		}

		if crop != nil {
			// 设置了resize, 并且resize的大小与上传图片不一样, 需要进行resize操作.
			if (crop.Width > 0 || crop.Height > 0) && (width != crop.Width || height != crop.Height) {
				resourceName := c.Request.RequestURI
				enlargeSmaller := crop.EnlargeSmaller
				bodyNew, err := utils.ResizeBytesImgToBytes(body, resourceName, crop.Width, crop.Height, enlargeSmaller, quality)
				if err != nil {
					log.Errorf("ResizeBytesImgToBytes(filename: %s, width: %d, height: %d, enlarge: %v) failed! err: %v",
						resourceName, crop.Width, crop.Height, enlargeSmaller, err)
					ctx.JSONFail(400, errors.ErrArgsInvalid)
					return
				} else {
					log.Infof("image [%s] success resize to %dx%d", resourceName, crop.Width, crop.Height)
				}
				body = bodyNew
				// 重新计算HASH.
				hash = coreutils.Sha1hex(body)
				// 重新获取图像大小.
				size, err := utils.GetImageSize(body)
				if err != nil {
					log.Errorf("Get Image[%s]'s size failed! err: %v", c.Request.RequestURI, err)
				} else {
					width = size.Width
					height = size.Height
				}
			}
		}
	}

	if rid == "" {
		rid = utils.HashToRid(appID, hash)
		log.Infof("appID: %s, hash: %s ==> rid: %s", appID, hash, rid)
	} else {
		rid = utils.IdToRid(appID, rid)
	}

	uploadFileDao := dao.NewUploadFileDao()
	resInfo, err := uploadFileDao.GetInfoByID(rid)
	if err == nil && resInfo != nil {
		log.Infof("-------- file [fileType=%s,hash=%s,url=%s] is uploaded! -------- ", fileType, hash, resInfo.Path)
		data := gin.H{"url": resInfo.Path, "rid": resInfo.Rid, "fileType": fileType}
		ctx.JSONOK(data)
		return
	}

	var filename string
	if target == "" {
		if strResourceType == "" {
			strResourceType = fileType
		}
		filename = createFilename(appID, hash, strResourceType, suffix)
	} else {
		filename = target
	}

	if isTest {
		filename = "test/" + filename
	}

	bodyReader = bytes.NewReader(body)
	bodyLen = int64(len(body))
	url, reason := SaveFileToCloud(appID, rid, hash, filename, contentType, bodyReader, bodyLen, width, height, 0, "")
	if url == "" {
		ctx.JSONFail(500, reason)
		return
	}
	log.Infof("write file %s ok. url is: %s", filename, url)

	ctx.JSONOK(gin.H{"url": url, "rid": rid})

	return
}

// CheckExist 检查资源是否存在.
func CheckExist(c *gin.Context) {
	ctx := NewUContetPlus(c)
	appID := ctx.MustGet("appID").(string)

	var resInfo *dao.UploadFile
	var err error

	uploadFileDao := dao.NewUploadFileDao()
	URL := ctx.GetCustomHeader("URL")
	if URL != "" {
		// 如果Path参数不为空
		resInfo, err = uploadFileDao.GetInfoByURL(URL, appID)
	} else {
		rid := ctx.GetCustomHeader("ID")
		if rid != "" {
			rid = utils.IdToRid(appID, rid)
			resInfo, err = uploadFileDao.GetInfoByID(rid)
		} else {
			hash := ctx.GetCustomHeader("Hash")
			resInfo, err = uploadFileDao.GetInfoByHash(hash, appID)
		}
	}

	if err == nil {
		if resInfo != nil {
			data := gin.H{"url": resInfo.Path, "rid": resInfo.Rid}
			ctx.JSONOK(data)
		} else {
			data := gin.H{"ok": true, "reason": "NOT-EXIST"}
			ctx.JSON(200, data)
		}
	} else {
		ctx.JSONFail(500, errors.ErrServerError)
	}

	return
}
