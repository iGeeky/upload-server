package controller

import (
	"bufio"
	"encoding/json"
	"math"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/iGeeky/open-account/pkg/baselib/errors"
	"github.com/iGeeky/open-account/pkg/baselib/log"
	coreutils "github.com/iGeeky/open-account/pkg/baselib/utils"
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
	notCompletedChunks := make([]int, 0)
	for i := 0; i < chunks; i++ {
		completedChunks = append(completedChunks, i)
	}
	data := gin.H{"completedChunks": completedChunks, "notCompletedChunks": notCompletedChunks, "url": url, "rid": rid}

	ctx.JSONOK(data)
}

func uploadChunkInitInternal(ctx *UContextPlus, contentType, hash string, filesize int) {
	appID := ctx.MustGet("appID").(string)
	fileType, _ := usutils.GetFileTypeAndSuffix(contentType)
	strResourceType := ctx.GetCustomHeader("Type")
	isTest := ctx.GetCustomHeader("Test") == "1"
	chunksize := ctx.MustGetCustomHeaderInt("ChunkSize", configs.Config.ChunkSize)

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

func saveChunkInternal(appID, rid, contentType, strResourceType string, hash string,
	filesize int64, chunkSize int, chunkindex int, chunkData []byte, isTest bool) (bool, interface{}) {

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

	ok, err := xf.HeadIsInited()
	if !ok {
		log.Errorf("file(%s) not inited! ", ulfile)
		return false, errors.ErrServerError
	}

	chunkDataLen := len(chunkData)
	log.Infof("write block, hash: %s, chunkindex: %d, chunkDataLen: %d", hash, chunkindex, chunkDataLen)

	err = xf.BlockWrite(chunkindex, chunkData, chunkDataLen)
	if err != nil {
		log.Errorf(" xf.BlockWrite(%s,block_idx:%d,write_len:%d) failed! err: %v", ulfile, chunkindex, chunkDataLen, err)
		return false, errors.ErrServerError
	} else {
		log.Infof("write block, hash: %s, chunkindex: %d ok", hash, chunkindex)

		err = xf.SetBlockUploaded(chunkindex)
		if err != nil {
			log.Errorf("xf.SetBlockUploaded(%s,block_idx: %d) failed! err: %v", ulfile, chunkindex, err)
			return false, errors.ErrServerError
		}
	}

	var status *xfile.BlockStatus
	status, err = xf.GetBlockStatus()
	if status == nil {
		log.Errorf("xf.GetBlockStatus(%s) failed! err:%v", ulfile, err)
		return false, errors.ErrServerError
	}

	var fileStatus FileStatus
	fileStatus.CompletedChunks = status.CompletedChunks
	fileStatus.NotCompletedChunks = status.NotCompletedChunks

	//所有块已经上传完成。
	if len(status.NotCompletedChunks) == 0 {
		log.Infof("file (%s) all chunk upload success!", filename)
		datafile, err := xf.GetDataFile()
		if err != nil {
			log.Errorf("open (%s) railed! err: %v", filename, err)
			return false, errors.ErrServerError
		}
		defer datafile.Close()

		var calcHash string
		bodyReader := bufio.NewReaderSize(datafile, 1024*128)
		log.Infof("begin hash_hex(%s) ...", filename)
		calcHash, _, err = usutils.SimpleHash(datafile, filesize)
		log.Infof("end hash_hex(%s) calcHash: %s", filename, calcHash)

		if err != nil {
			log.Errorf("hash_hex(%s) failed! err: %v", filename, err)
			return false, errors.ErrServerError
		}

		if hash != calcHash {
			log.Errorf("request hash is %s, calcHash is %s", hash, calcHash)
			return false, utils.ErrHashInvalid
		}

		datafile.Seek(0, 0)
		// bodyReader = bufio.NewReader(datafile)
		// TODO: 如果是视频, 获取视频信息.
		extinfo := ""
		url, reason := SaveFileToCloud(appID, rid, hash, filename, contentType, bodyReader, filesize, 0, 0, 0, extinfo)
		if url == "" {
			return false, reason
		}

		xf.Close()
		// 删除临时文件.
		xf.DeleteFiles()

		fileStatus.Rid = rid
		fileStatus.URL = url

		log.Infof("write file %s ok. url is: %s", filename, url)
	} else {
		_, err := json.Marshal(fileStatus)
		if err != nil {
			log.Infof("hash: %s, chunkindex: %d, marshal error:%s", hash, chunkindex, err.Error())
		} else {
			log.Infof("hash: %s, chunkindex: %d, completeNum:%d,NotCompleteNum:%d", hash, chunkindex, len(status.CompletedChunks), len(status.NotCompletedChunks))
		}

	}

	return true, fileStatus
}

// UploadChunkUpload 分块上传.
func UploadChunkUpload(c *gin.Context) {
	/**
	* X-OA-hash: $filehash 文件HASH(sha1)
	* X-OA-filesize: $filesize(文件大小)
	* X-OA-chunksize: $chunksize(块大小)
	* X-OA-chunkindex: $chunkindex(块索引号，从0开始)
	* X-OA-chunkhash: $chunk_hash(块HASH，sha1)
	**/
	ctx := NewUContextPlus(c)
	appID := ctx.MustGet("appID").(string)
	contentType := ctx.GetHeader("Content-Type")
	contentType = contentTypeTrim(contentType)
	strResourceType := ctx.GetCustomHeader("Type")
	hash := ctx.MustGetCustomHeader("Hash")

	filesize := ctx.MustGetCustomHeaderInt64("FileSize", -1)
	chunksize := ctx.MustGetCustomHeaderInt("ChunkSize", configs.Config.ChunkSize)
	chunkindex := ctx.MustGetCustomHeaderInt("ChunkIndex", 0)
	chunkhash := ctx.MustGetCustomHeader("ChunkHash")

	body, _ := ctx.GetBody()

	calcChunkHash := coreutils.Sha1hex(body)
	if chunkhash != calcChunkHash {
		log.Errorf("chunkhash[%v] != calcChunkHash[%v]", chunkhash, calcChunkHash)
		ctx.JSONFail(400, utils.ErrHashInvalid)
		return
	}

	isTest := ctx.GetCustomHeader("Test") == "1"
	rid := ctx.GetCustomHeader("Id")
	if rid == "" {
		rid = utils.HashToRid(appID, hash)
		log.Infof("appID: %s, hash: %s ==> rid: %s", appID, hash, rid)
	} else {
		rid = utils.IdToRid(appID, rid)
	}

	ok, status := saveChunkInternal(appID, rid, contentType, strResourceType, hash, filesize, chunksize, chunkindex, body, isTest)
	// log.Infof("Save Chunk .................")
	if ok {
		chunkStatus := status.(FileStatus)
		data := gin.H{"completedChunks": chunkStatus.CompletedChunks, "notCompletedChunks": chunkStatus.NotCompletedChunks, "url": chunkStatus.URL, "rid": chunkStatus.Rid}
		ctx.JSONOK(data)
	} else {
		// log.Errorf("error --------------------------- ")
		reason := status.(string)
		ctx.JsonFail(reason)
		return
	}
}
