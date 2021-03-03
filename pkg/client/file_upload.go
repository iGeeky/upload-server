package client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
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

// UploadClient 上传客户端
type UploadClient struct {
	Host               string
	AppID              string
	AppKey             string
	CustomHeaderPrefix string
}

// NewUploadClient 创建新的上传客户端
func NewUploadClient(host, appID, appKey string) (client *UploadClient) {
	client = &UploadClient{
		Host:               host,
		AppID:              appID,
		AppKey:             appKey,
		CustomHeaderPrefix: "X-OA-",
	}
	return
}

// SetCustomHeaderPrefix 设置请求头前缀.
func (client *UploadClient) SetCustomHeaderPrefix(customHeaderPrefix string) {
	client.CustomHeaderPrefix = customHeaderPrefix
}

// CustomHeaderName 获取自定义的请求头名.
func (client *UploadClient) CustomHeaderName(headerName string) (customHeaderName string) {
	return client.CustomHeaderPrefix + headerName
}

// CheckFileExist 检查资源是否存在
func (client *UploadClient) CheckFileExist(id, hash, url string, timeout time.Duration) *net.OkJson {
	uri := client.Host + "/v1/upload/check_exist"

	res := &net.OkJson{Ok: false, Reason: errors.ErrArgsInvalid}

	headers := make(map[string]string, 10)
	headers[client.CustomHeaderName("Platform")] = "test"
	headers[client.CustomHeaderName("AppID")] = client.AppID
	if id != "" {
		headers[client.CustomHeaderName("ID")] = id
	}
	if hash != "" {
		headers[client.CustomHeaderName("Hash")] = hash
	}
	if url != "" {
		headers[client.CustomHeaderName("URL")] = url
	}

	res = client.HttpGetWithSign(uri, headers, timeout)

	return res
}

// PostBody2UploadSvr 上传content
func (client *UploadClient) PostBody2UploadSvr(filename string, content []byte, id, target, resourceType, imageProcess string, timeout time.Duration, isTest bool) *net.OkJson {
	uri := client.Host + "/v1/upload/simple"
	res := &net.OkJson{Ok: false, Reason: errors.ErrArgsInvalid}

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
	headers[client.CustomHeaderName("Platform")] = "test"
	headers[client.CustomHeaderName("hash")] = hash
	headers[client.CustomHeaderName("AppID")] = client.AppID
	if id != "" {
		headers[client.CustomHeaderName("ID")] = id
	}
	if resourceType != "" {
		headers[client.CustomHeaderName("Type")] = resourceType
	}
	if imageProcess != "" {
		headers[client.CustomHeaderName("ImageProcess")] = imageProcess
	}
	if target != "" {
		headers[client.CustomHeaderName("Target")] = target
	}
	if isTest {
		headers[client.CustomHeaderName("Test")] = "1"
	}

	res = client.HttpPostWithSign(uri, content, headers, timeout)

	return res
}

// PostFile2UploadSvr 上传文件.
func (client *UploadClient) PostFile2UploadSvr(filename, id, target, resourceType, imageProcess string, timeout time.Duration, isTest bool) *net.OkJson {
	bodyfile, err := os.Open(filename)
	res := &net.OkJson{Ok: false, Reason: errors.ErrArgsInvalid}
	if err != nil {
		log.Errorf("Open file failed! err: %v", err)
		res.Reason = ErrOpenInputFile
		res.StatusCode = 400
		return res
	}
	defer bodyfile.Close()

	content, _ := ioutil.ReadFile(filename)

	res = client.PostBody2UploadSvr(filename, content, id, target, resourceType, imageProcess, timeout, isTest)
	return res
}

// UploadURLSimpleReq 上传URL请求体.
type UploadURLSimpleReq struct {
	URL       string `json:"url"`
	Referer   string `json:"referer"`
	UserAgent string `json:"userAgent"`
}

// UploadURL 通过URL上传文件
func (client *UploadClient) UploadURL(uploadURL *UploadURLSimpleReq, id, target, resourceType, imageProcess string, timeout time.Duration, isTest bool) *net.OkJson {
	uri := client.Host + "/v1/upload/url"
	res := &net.OkJson{Ok: false, Reason: errors.ErrArgsInvalid}

	headers := make(map[string]string, 10)
	headers[client.CustomHeaderName("Platform")] = "test"
	headers[client.CustomHeaderName("AppID")] = client.AppID
	if id != "" {
		headers[client.CustomHeaderName("ID")] = id
	}
	if resourceType != "" {
		headers[client.CustomHeaderName("Type")] = resourceType
	}
	if imageProcess != "" {
		headers[client.CustomHeaderName("ImageProcess")] = imageProcess
	}
	if target != "" {
		headers[client.CustomHeaderName("Target")] = target
	}
	if isTest {
		headers[client.CustomHeaderName("Test")] = "1"
	}

	body, err := json.Marshal(uploadURL)
	if err != nil {
		res.Error = err
		return res
	}

	res = client.HttpPostWithSign(uri, body, headers, timeout)

	return res
}

// ChunkResponse 分块上传返回结构
type ChunkResponse struct {
	CompletedChunks    []int  `json:"completedChunks"`    //已经上传成功的块ID列表
	NotCompletedChunks []int  `json:"notCompletedChunks"` //未上传的块ID列表
	URL                string `json:"url"`                //资源URL
	Rid                string `json:"rid"`                //资源ID
	ChunkSize          int    `json:"chunkSize"`          //块大小（后面按该大小分块上传，不同的文件，块大小可能不同）
}

func chunkResponseParse(body map[string]interface{}) (resp *ChunkResponse, err error) {
	resp = &ChunkResponse{}
	buf, err := json.Marshal(body)
	if err != nil {
		log.Errorf("Marshal(%v) failed! err: %v", body, err)
		return nil, err
	}

	err = json.Unmarshal(buf, resp)
	if err != nil {
		resp = nil
		return
	}
	return
}

func (client *UploadClient) chunkInit(headers map[string]string, timeout time.Duration) (*net.OkJson, *ChunkResponse) {
	uri := client.Host + "/v1/upload/chunk/init"

	res := client.HttpPostWithSign(uri, []byte(""), headers, timeout)
	if res.StatusCode != 200 {
		log.Errorf("request [%s] failed! err: %v", res.ReqDebug, res.Error)
		if res.Error == nil {
			res.Error = fmt.Errorf("http-error: %d", res.StatusCode)
		}
		return res, nil
	}

	if !res.Ok {
		log.Errorf("request [%s] failed! reason: %v", res.ReqDebug, res.Reason)
		return res, nil
	}

	resp, err := chunkResponseParse(res.Data)
	if err != nil {
		log.Errorf("request [%s] parse response [%s] failed! err: %v", res.ReqDebug, string(res.RawBody), err)
		res.Error = err
		return res, nil
	}

	return res, resp
}

func getOffsetAndChunksize(chunkSize, chunkIndex int, filesize int64) (int64, int) {
	offset := int64(0)
	offset = int64(chunkSize) * int64(chunkIndex)
	chunks := int(math.Ceil(float64(filesize) / float64(chunkSize)))
	if chunkIndex == chunks-1 { //最后一块。
		chunkSizeNew := filesize % int64(chunkSize)
		if chunkSizeNew > 0 {
			chunkSize = int(chunkSizeNew)
		}
	}

	return offset, chunkSize
}

func (client *UploadClient) chunkUpload(bodyfile *os.File, headers map[string]string, filesize int64, chunkSize, chunkIndex int, timeout time.Duration) (*net.OkJson, *ChunkResponse) {
	res := &net.OkJson{Ok: false, Reason: errors.ErrArgsInvalid}
	uri := client.Host + "/v1/upload/chunk/upload"
	filename := headers[client.CustomHeaderName("FileName")]

	offset, chunkSizeReal := getOffsetAndChunksize(chunkSize, chunkIndex, filesize)
	if offset >= filesize {
		log.Errorf("chunkSize(%d) x chunkIndex(%d) : offset(%d) >= filesize(%d)", chunkSize, chunkIndex, offset, filesize)
		res.Error = fmt.Errorf("chunk invalid")
		return res, nil
	}

	buf := make([]byte, chunkSizeReal)
	_, err := bodyfile.ReadAt(buf, offset)
	if err != nil {
		log.Errorf("ReadAt(%s, offset: %d, len: %d) failed! err: %v", filename, offset, chunkSizeReal, err)
		res.Error = err
		return res, nil
	}
	chunkhash := utils.Sha1hex(buf)

	headers[client.CustomHeaderName("ChunkSize")] = strconv.Itoa(chunkSizeReal)
	headers[client.CustomHeaderName("ChunkIndex")] = strconv.Itoa(chunkIndex)
	headers[client.CustomHeaderName("ChunkHash")] = chunkhash

	res = client.HttpPostWithSign(uri, buf, headers, timeout)
	if res.StatusCode != 200 {
		log.Errorf("request [%s] failed! status: %d, err: %v", res.ReqDebug, res.StatusCode, res.Error)
		if res.Error == nil {
			res.Error = fmt.Errorf("http-error: %d", res.StatusCode)
		}
		return res, nil
	}

	if !res.Ok {
		log.Errorf("request [%s] failed! reason: %v", res.ReqDebug, res.Reason)
		return res, nil
	}

	log.Infof("--- res.Data: %+v", res.Data)
	resp, err := chunkResponseParse(res.Data)
	if err != nil {
		log.Errorf("request [%s] parse response [%s] failed! err: %v", res.ReqDebug, string(res.RawBody), err)
		res.Error = err
		return res, nil
	}
	return res, resp
}

// ChunkUpload 分块上传.
func (client *UploadClient) ChunkUpload(filename, id, resourceType string, chunkSize int, timeout time.Duration, isTest bool) (*net.OkJson, *ChunkResponse) {
	res := &net.OkJson{Ok: false, Reason: errors.ErrArgsInvalid}
	hash, filesize, err := usutils.SimpleHashFile(filename)
	if err != nil {
		log.Errorf("Md5File(%s) failed! err: %v", filename, err)
		res.Error = err
		return res, nil
	}

	headers := make(map[string]string, 10)
	headers["Content-Type"] = getContentType(filename)
	headers[client.CustomHeaderName("AppID")] = client.AppID
	headers[client.CustomHeaderName("Hash")] = hash
	headers[client.CustomHeaderName("FileName")] = url.QueryEscape(filename)
	headers[client.CustomHeaderName("FileSize")] = strconv.FormatInt(filesize, 10)
	if resourceType != "" {
		headers[client.CustomHeaderName("Type")] = resourceType
	}
	if chunkSize > 0 {
		headers[client.CustomHeaderName("ChunkSize")] = strconv.Itoa(chunkSize)
	}
	if id != "" {
		headers[client.CustomHeaderName("ID")] = id
	}
	if isTest {
		headers[client.CustomHeaderName("Test")] = "1"
	}

	res, chunkResp := client.chunkInit(headers, timeout)
	if res.Error != nil {
		log.Errorf("chunkInit(headers:%v) failed! err: %v", headers, res.Error)
		return res, chunkResp
	}
	if len(chunkResp.NotCompletedChunks) == 0 && chunkResp.URL != "" {
		return res, chunkResp
	}

	if chunkResp.ChunkSize > 0 {
		chunkSize = chunkResp.ChunkSize
	}

	bodyfile, err := os.Open(filename)
	if err != nil {
		log.Errorf("Open file failed! err:%v", err)
		res = &net.OkJson{Ok: false, Reason: errors.ErrArgsInvalid}
		res.Error = err
		return res, nil
	}
	defer bodyfile.Close()

	NotCompletedChunks := chunkResp.NotCompletedChunks
	for i := 0; i < 3; i++ {
		for _, chunkIndex := range NotCompletedChunks {
			res, chunkResp = client.chunkUpload(bodyfile, headers, filesize, chunkSize, chunkIndex, timeout)
			if res.Error != nil || res.Reason != "" {
				log.Errorf("chunkUpload(headers: %v) failed! err: %v, Reason: %s", headers, res.Error, res.Reason)
				//TODO: 出错重试。
				return res, chunkResp
			}
		}

		NotCompletedChunks = chunkResp.NotCompletedChunks
		if len(NotCompletedChunks) < 1 {
			break
		}
	}
	if len(NotCompletedChunks) > 0 {
		if res.Error == nil {
			res.Error = fmt.Errorf("Upload-Failed: Too many errors")
		}
		if res.Reason == "" {
			res.Reason = "ERR_UPLOAD_FAILED"
		}
	}
	return res, chunkResp
}
