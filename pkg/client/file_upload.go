package client

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/iGeeky/open-account/pkg/baselib/errors"
	"github.com/iGeeky/open-account/pkg/baselib/log"
	"github.com/iGeeky/open-account/pkg/baselib/net"
	"github.com/iGeeky/open-account/pkg/baselib/utils"
	usutils "github.com/iGeeky/upload-server/pkg/utils"
)

const (
	// ErrContentTypeInvalid 请求的Content-Type不合法.
	ErrContentTypeInvalid string = "ERR_CONTENT_TYPE_INVALID"
	// ErrOpenInputFile 打开文件出错
	ErrOpenInputFile string = "ERR_OPEN_INPUT_FILE"
)

// 需要签名的请求头前缀
var gCustomHeaderPrefix = "X-OA-"

// InitCustomHeaderPrefix 初始化
func InitCustomHeaderPrefix(customHeaderPrefix string) {
	if customHeaderPrefix != "" {
		gCustomHeaderPrefix = customHeaderPrefix
	}
}

// CustomHeaderName 获取自定义的请求头名.
func CustomHeaderName(headerName string) (customHeaderName string) {
	return gCustomHeaderPrefix + headerName
}

func getContentType(filename string) string {
	ext := filepath.Ext(strings.ToLower(filename))
	if strings.HasPrefix(ext, ".") {
		ext = ext[1:]
	}
	contentType := usutils.GetContentTypeByExt(ext)
	log.Infof("### filename: %s, ext: %s, contentType: %s", filename, ext, contentType)

	// if contentType == "" {
	// 	contentType = "application/octet-stream"
	// }

	return contentType
}

// CheckFileExist 检查资源是否存在
func CheckFileExist(host, appID, appKey, id, hash, url string, timeout time.Duration) *net.OkJson {
	uri := host + "/v1/upload/check_exist"

	res := &net.OkJson{Ok: false, Reason: errors.ErrServerError}

	headers := make(map[string]string, 10)
	headers[CustomHeaderName("Platform")] = "test"
	headers[CustomHeaderName("AppID")] = appID
	if id != "" {
		headers[CustomHeaderName("ID")] = id
	}
	if hash != "" {
		headers[CustomHeaderName("Hash")] = hash
	}
	if url != "" {
		headers[CustomHeaderName("URL")] = url
	}

	res = HttpGetWithSign(uri, headers, timeout, appKey)

	return res
}

func PostBody2UploadSvr(host, filename string, content []byte, appID, appKey, id, target, fileType, imageProcess string, timeout time.Duration, isTest bool) *net.OkJson {
	uri := host + "/v1/upload/simple"
	res := &net.OkJson{Ok: false, Reason: errors.ErrServerError}

	contentType := getContentType(filename)
	if contentType == "" {
		log.Errorf("un support file type: %s", filename)
		res.Reason = ErrContentTypeInvalid
		res.StatusCode = 400
		return res
	}

	hash := utils.Sha1hex(content)
	headers := make(map[string]string, 10)
	headers["Content-Type"] = contentType
	headers[CustomHeaderName("Platform")] = "test"
	headers[CustomHeaderName("hash")] = hash
	headers[CustomHeaderName("AppID")] = appID
	if id != "" {
		headers[CustomHeaderName("ID")] = id
	}
	if fileType != "" {
		headers[CustomHeaderName("Type")] = fileType
	}
	if imageProcess != "" {
		headers[CustomHeaderName("ImageProcess")] = imageProcess
	}
	if target != "" {
		headers[CustomHeaderName("Target")] = target
	}
	if isTest {
		headers[CustomHeaderName("Test")] = "1"
	}

	res = HttpPostWithSign(uri, content, headers, timeout, appKey)

	return res
}

func PostFile2UploadSvr(host, filename, appID, appKey, id, target, fileType, imageProcess string, timeout time.Duration, isTest bool) *net.OkJson {
	bodyfile, err := os.Open(filename)
	res := &net.OkJson{Ok: false, Reason: errors.ErrServerError}
	if err != nil {
		log.Errorf("Open file failed! err: %v", err)
		res.Reason = ErrOpenInputFile
		res.StatusCode = 400
		return res
	}
	defer bodyfile.Close()

	content, _ := ioutil.ReadFile(filename)

	res = PostBody2UploadSvr(host, filename, content, appID, appKey, id, target, fileType, imageProcess, timeout, isTest)
	return res
}

type UploadURLSimpleReq struct {
	URL       string `json:"url"`
	Referer   string `json:"referer"`
	UserAgent string `json:"userAgent"`
}

func UploadURL(host, appID, appKey string, uploadURL *UploadURLSimpleReq, id, target, fileType, imageProcess string, timeout time.Duration, isTest bool) *net.OkJson {
	uri := host + "/v1/upload/url"
	res := &net.OkJson{Ok: false, Reason: errors.ErrServerError}

	headers := make(map[string]string, 10)
	headers[CustomHeaderName("Platform")] = "test"
	headers[CustomHeaderName("AppID")] = appID
	if id != "" {
		headers[CustomHeaderName("ID")] = id
	}
	if fileType != "" {
		headers[CustomHeaderName("Type")] = fileType
	}
	if imageProcess != "" {
		headers[CustomHeaderName("ImageProcess")] = imageProcess
	}
	if target != "" {
		headers[CustomHeaderName("Target")] = target
	}
	if isTest {
		headers[CustomHeaderName("Test")] = "1"
	}

	body, err := json.Marshal(uploadURL)
	if err != nil {
		res.Error = err
		return res
	}

	res = HttpPostWithSign(uri, body, headers, timeout, appKey)

	return res
}
