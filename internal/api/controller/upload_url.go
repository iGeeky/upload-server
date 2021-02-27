package controller

import (
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/iGeeky/open-account/pkg/baselib/errors"
	"github.com/iGeeky/open-account/pkg/baselib/log"
	"github.com/iGeeky/open-account/pkg/baselib/net"
	"github.com/iGeeky/upload-server/configs"
	"github.com/iGeeky/upload-server/pkg/utils"
)

var defUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_10_4) AppleWebKit/600.7.12 (KHTML, like Gecko) Version/8.0.7 Safari/600.7.12"

// UploadUrlSimpleReq 通过URL上传资源.
type UploadUrlSimpleReq struct {
	URL       string `json:"url" validate:"required"`
	Referer   string `json:"referer"`
	UserAgent string `json:"userAgent"`
}

// FileDownloadSimple 文件下载.
func FileDownloadSimple(url, referer, userAgent string) (body []byte, err error) {
	headers := make(map[string]string)
	if len(referer) > 0 {
		headers["Referer"] = referer
	}
	if len(userAgent) > 0 {
		headers["User-Agent"] = userAgent
	} else {
		headers["User-Agent"] = defUserAgent
	}
	// TODO: 重试.
	res := net.HttpGet(url, headers, configs.Config.DownloadTimeout)
	if res.StatusCode != 200 {
		log.Errorf("request [%s] failed! status: %d, err: %v, cost: %v",
			res.ReqDebug, res.StatusCode, res.Error, res.Stats.All)
		err = res.Error
		if err == nil {
			err = fmt.Errorf("HttpError: %d", res.StatusCode)
		}
		return
	} else {
		log.Infof("download request [%s] success!", res.ReqDebug)
	}
	body = res.RawBody
	return
}

func getExtFromPath(strPath string) (ext string) {
	ext = path.Ext(strPath)
	if strings.HasPrefix(ext, ".") {
		ext = ext[1:]
	}
	return
}

func getExtFromURI(uri string) (ext string) {
	u, _ := url.Parse(uri)
	if u.Path != "" {
		ext = getExtFromPath(u.Path)
	}
	return
}

func getContentTypeFromParams(ctx *UContextPlus, url string) (contentType string) {
	contentType = ctx.GetHeader("Content-Type")
	if contentType != "" && !strings.Contains(contentType, "application/json") {
		return
	}

	target := strings.TrimSpace(ctx.GetCustomHeader("Target"))
	if target != "" {
		ext := getExtFromPath(target)
		contentType = utils.GetContentTypeByExt(ext)
		if contentType != "" {
			log.Infof("get contentType [%s] by target [%s] for: %s", contentType, target, url)
			return
		}
	}

	ext := getExtFromURI(url)
	contentType = utils.GetContentTypeByExt(ext)
	if contentType != "" {
		log.Infof("get contentType [%s] by ext [%s] for: %s", contentType, ext, url)
		return
	}

	contentType = "application/octet-stream"
	log.Infof("get contentType [%s] by default for: %s", contentType, url)
	return
}

// UploadURLSimple 通过URL上传图片
func UploadURLSimple(c *gin.Context) {
	ctx := NewUContextPlus(c)
	req := &UploadUrlSimpleReq{}
	ctx.ParseQueryJSONObject(req)

	body, err := FileDownloadSimple(req.URL, req.Referer, req.UserAgent)
	if err != nil {
		log.Errorf("FileDownloadSimple(%s, %s) failed! err: %+v", req.URL, req.Referer, err)
		ctx.JSONFail(500, errors.ErrServerError)
		return
	}
	debugID := "url=" + req.URL
	contentType := getContentTypeFromParams(ctx, req.URL)
	uploadInternal(ctx, body, "", contentType, debugID)
	return
}
