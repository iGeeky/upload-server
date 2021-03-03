package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/iGeeky/open-account/pkg/baselib/log"
	"github.com/iGeeky/open-account/pkg/baselib/net"
	"github.com/iGeeky/upload-server/pkg/client"
)

func responseDebug(res *net.OkJson, chunkResp *client.ChunkResponse, verbose bool) {
	if verbose {
		fmt.Printf("response -------------- headers : %v\n", res.Headers)
		fmt.Printf("response ---------------- status: %d\n", res.StatusCode)
		if chunkResp != nil {
			fmt.Printf("chunk ---------- CompletedChunks: %v\n", chunkResp.CompletedChunks)
			fmt.Printf("chunk ------- NotCompletedChunks: %v\n", chunkResp.NotCompletedChunks)
		}
		fmt.Println(string(res.RawBody))
	}

	if res.StatusCode != 200 {
		fmt.Println("reason:", res.Reason)
		fmt.Println("Error:", res.Error)
		fmt.Println("reqDebug:", res.ReqDebug)
	} else {
		if res.Reason != "" {
			fmt.Println("reason:", res.Reason)
		} else {
			fmt.Println(res.Data["url"])
		}
	}
}

func main() {
	var verbose bool
	var filename string

	var url string
	var referer string
	var userAgent string

	var chunkUpload bool
	var chunkSize int

	var host string
	var appID string
	var appKey string
	var customHeaderPrefix string

	var id string
	var target string
	var fileType string
	var imageProcess string
	var checkExist string
	var timeout time.Duration
	var logLevel string
	isTest := false

	flag.StringVar(&logLevel, "logLevel", "warn", "info,warn,error")
	flag.StringVar(&host, "host", "http://127.0.0.1:2022", "指定上传服务地址")
	flag.StringVar(&appID, "appID", "img", "App Id")
	flag.StringVar(&appKey, "appKey", "super", "App Key")
	flag.StringVar(&customHeaderPrefix, "customHeaderPrefix", "X-OA-", "X-OA-|X-APP-")

	flag.StringVar(&checkExist, "checkExist", "", "Params: id=resource-id | hash=resource-hash | url=http://xxx.com/resource/url")
	flag.StringVar(&filename, "file", "", "Filename(.jpg|.png|.gif|.mp4|.avi)")
	flag.StringVar(&url, "url", "", "Url, .eg: https://test.com/test.jpg")
	flag.StringVar(&referer, "referer", "", "Referer for url download")
	flag.StringVar(&userAgent, "userAgent", "", "UserAgent for url download")
	flag.BoolVar(&chunkUpload, "chunkUpload", chunkUpload, "")
	flag.IntVar(&chunkSize, "chunkSize", chunkSize, "chunkSize: 524288")
	flag.StringVar(&id, "id", "", "Resource Id")
	flag.StringVar(&target, "target", target, "target file path")
	flag.StringVar(&fileType, "type", fileType, "file type for logical")
	flag.StringVar(&imageProcess, "imageProcess", imageProcess, "ImageProcess, .eg: resize,fw_300,fh_200/quality,q_80")

	flag.BoolVar(&verbose, "verbose", verbose, "verbose")
	flag.DurationVar(&timeout, "timeout", time.Minute*5, "upload timeout, like: 300ms, 10s, 2m, 1h")
	flag.Parse()

	log.InitLogger(logLevel, "", true, true)
	uploadClient := client.NewUploadClient(host, appID, appKey)
	uploadClient.SetCustomHeaderPrefix(customHeaderPrefix)

	if checkExist != "" {
		var id, hash, url string
		if strings.HasPrefix(checkExist, "id=") {
			id = checkExist[3:]
		} else if strings.HasPrefix(checkExist, "hash=") {
			hash = checkExist[5:]
		} else if strings.HasPrefix(checkExist, "url=") {
			url = checkExist[4:]
		} else {
			fmt.Printf("Invalid value [%+v] for checkExist", checkExist)
			os.Exit(1)
		}
		res := uploadClient.CheckFileExist(id, hash, url, timeout)
		responseDebug(res, nil, verbose)
		return
	}
	if filename != "" {
		if chunkUpload {
			// filename, id, resourceType string, chunksize int, timeout time.Duration, isTest bool
			res, chunkResp := uploadClient.ChunkUpload(filename, id, fileType, chunkSize, timeout, isTest)
			responseDebug(res, chunkResp, verbose)
		} else {
			res := uploadClient.PostFile2UploadSvr(filename, id, target, fileType, imageProcess, timeout, isTest)
			responseDebug(res, nil, verbose)
		}
		return
	}
	if url != "" {
		uploadURL := &client.UploadURLSimpleReq{
			URL:       url,
			Referer:   referer,
			UserAgent: userAgent,
		}
		res := uploadClient.UploadURL(uploadURL, id, target, fileType, imageProcess, timeout, isTest)
		responseDebug(res, nil, verbose)
		return
	}

}
