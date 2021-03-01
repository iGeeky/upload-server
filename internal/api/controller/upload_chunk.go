package controller

import (
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/iGeeky/open-account/pkg/baselib/errors"
	"github.com/iGeeky/open-account/pkg/baselib/log"
	"github.com/iGeeky/upload-server/configs"
	"github.com/iGeeky/upload-server/internal/api/dao"
	"github.com/iGeeky/upload-server/internal/api/utils"
	usutils "github.com/iGeeky/upload-server/pkg/utils"
	"github.com/iGeeky/upload-server/pkg/xfile"
)

// FileStatus 文件上传状态.
type FileStatus struct {
	xfile.BlockStatus
	URL string
	Rid string
	// Md5       string
	ChunkSize int
}

func getCurrentChunkSize(filesize int, chunksize int, chunkindex int) (contentChunkSize int, chunks int) {
	chunks = int(math.Ceil(float64(filesize) / float64(chunksize)))
	fullchunks := filesize / chunksize
	if chunkindex >= 0 && chunkindex < fullchunks {
		contentChunkSize = chunksize
	} else if chunkindex == fullchunks {
		contentChunkSize = filesize % chunksize
	}
	return
}

func initChunk(appID, contentType, strResourceType string, hash string, filesize int, chunkSize int, isTest bool) (bool, interface{}) {
	fileType, suffix := usutils.GetFileTypeAndSuffix(contentType)
	if fileType == "" {
		return false, suffix
	}
	if strResourceType == "" {
		strResourceType = fileType
	}

	var filename = createFilename(appID, hash, strResourceType, suffix)
	if isTest {
		filename = "test/" + filename
	}
	var ulfile = configs.Config.ChunkTmpDir + filename + ".data"

	var xf, err = xfile.Open(ulfile)
	if xf == nil {
		log.Errorf("xfile.Open(%s) failed! err:%v", ulfile, err)
		return false, errors.ErrServerError
	}
	defer xf.Close()

	var inited bool
	inited, err = xf.HeadIsInited()
	if err != nil {
		log.Errorf("xf.HeadIsInited(%s) failed! err: %v", ulfile, err)
		return false, errors.ErrServerError
	}
	if !inited {
		var createTime = uint32(time.Now().Unix())
		var err = xf.HeadInit(int64(filesize), chunkSize, createTime)
		if err != nil {
			log.Errorf("xf.HeadInit(%s,filesize:%d,chunkSize:%d) failed! err:%v", ulfile, filesize, chunkSize, err)
			return false, errors.ErrServerError
		}
	}

	var status *xfile.BlockStatus
	status, err = xf.GetBlockStatus()
	if status == nil {
		log.Errorf("xf.GetBlockStatus(%s) failed! err:%v", ulfile, err)
		return false, errors.ErrServerError
	}

	head, err := xf.GetHead()
	if err != nil {
		log.Errorf("xf.GetHead(%s) failed! err:%v", ulfile, err)
		return false, errors.ErrServerError
	}
	var fileStatus FileStatus
	fileStatus.ChunkSize = int(head.BlockSize)
	fileStatus.CompletedChunks = status.CompletedChunks
	fileStatus.NotCompletedChunks = status.NotCompletedChunks

	// 所有块都已经上传完成，但是未上传到阿里云上，由于处理过程复杂，先简单告诉客户端，一个块未完成。让其重传。
	if len(fileStatus.NotCompletedChunks) == 0 && len(fileStatus.CompletedChunks) > 0 {
		lastIdx := len(fileStatus.CompletedChunks) - 1
		lastEle := fileStatus.CompletedChunks[lastIdx]
		fileStatus.CompletedChunks = fileStatus.CompletedChunks[0 : lastIdx-1]
		fileStatus.NotCompletedChunks = append(fileStatus.NotCompletedChunks, lastEle)
	}

	return true, fileStatus
}

func returnChunkInfo(ctx *UContextPlus, fileType string, chunks int, url string, rid string, extinfo string) {
	completedChunks := make([]int, 0)
	notcompletedChunks := make([]int, 0)
	for i := 0; i < chunks; i++ {
		completedChunks = append(completedChunks, i)
	}
	data := gin.H{"completedChunks": completedChunks, "notCompletedChunks": notcompletedChunks, "url": url, "rid": rid}

	ctx.JSONOK(data)
}

func uploadChunkInitInternal(ctx *UContextPlus, contentType, hash string, filesize int) {
	appID := ctx.MustGet("appID").(string)
	fileType, _ := usutils.GetFileTypeAndSuffix(contentType)
	strResourceType := strings.TrimSpace(ctx.GetCustomHeader("Type"))

	chunksize := configs.Config.ChunkSize
	isTest := ctx.GetCustomHeader("Test") == "1"

	strChunkSize := ctx.GetCustomHeader("ChunkSize")
	if strChunkSize != "" {
		reqChunkSize, err := strconv.Atoi(strChunkSize)
		if err == nil {
			chunksize = reqChunkSize
		}
	}

	//生成rid, 检查在数据库里面有没有
	rid := ctx.GetCustomHeader("Id")
	if rid == "" {
		rid = utils.HashToRid(appID, hash)
	} else {
		rid = utils.IdToRid(appID, rid)
	}

	uploadFileDao := dao.NewUploadFileDao()
	resInfo, err := uploadFileDao.GetInfoByID(rid)
	if err == nil && resInfo != nil {
		log.Infof("-------- file [fileType=%s,hash=%s,url=%s] is uploaded! -------- ", fileType, hash, resInfo.Path)
		chunks := int(math.Ceil(float64(filesize) / float64(chunksize)))
		returnChunkInfo(ctx, fileType, chunks, resInfo.Path, resInfo.Rid, resInfo.Extinfo)
		return
	}

	log.Infof("initChunk(appID:%s, contentType: %s, hash: %s, filesize: %v, chunksize: %d, isTest: %v",
		appID, contentType, hash, filesize, chunksize, isTest)
	ok, status := initChunk(appID, contentType, strResourceType, hash, filesize, chunksize, isTest)

	if ok {
		chunkStatus := status.(FileStatus)
		if chunkStatus.ChunkSize > 0 {
			chunksize = chunkStatus.ChunkSize
		}
		data := gin.H{
			"completedChunks":    chunkStatus.CompletedChunks,
			"notCompletedChunks": chunkStatus.NotCompletedChunks,
			"url":                "", "rid": "", "chunksize": chunksize,
		}
		ctx.JSONOK(data)
	} else {
		reason := status.(string)
		ctx.JsonFail(reason)
	}
}

// UploadChunkInit 分成上传初始化
func UploadChunkInit(c *gin.Context) {
	ctx := NewUContextPlus(c)
	hash := ctx.GetCustomHeader("Hash")
	strFileSize := ctx.GetCustomHeader("FileSize")
	i, err := strconv.ParseInt(strFileSize, 10, 32)
	if err != nil {
		log.Errorf("The value <%s> of request header <%s> is invalid", strFileSize, ctx.CustomHeaderName("FileSize"))
		ctx.JsonFail(errors.ErrArgsInvalid)
		return
	}
	filesize := int(i)

	contentType := ctx.GetHeader("Content-Type")
	contentType = contentTypeTrim(contentType)

	uploadChunkInitInternal(ctx, contentType, hash, filesize)
}
