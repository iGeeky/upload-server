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

func responseDebug(res *net.OkJson, verbose bool) {
	if verbose {
		fmt.Printf("response -------------- headers : %v\n", res.Headers)
		fmt.Printf("response ---------------- status: %d\n", res.StatusCode)
		fmt.Println(string(res.RawBody))
	}

	if res.StatusCode != 200 {
		fmt.Println("reason:", res.Reason)
		fmt.Println("Error:", res.Error)
		fmt.Println("reqDebug:", res.ReqDebug)
	} else {
		fmt.Println(string(res.RawBody))
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
	var host string
	var appID string
	var appKey string
	var id string
	var imageProcess string
	var fileType string
	var target string
	var checkExist string
	var customHeaderPrefix string
	var timeout time.Duration
	var logLevel string
	isTest := false

	flag.StringVar(&logLevel, "logLevel", "warn", "info,warn,error")
	flag.StringVar(&customHeaderPrefix, "customHeaderPrefix", "X-OA-", "X-OA-|X-APP-")
	flag.StringVar(&checkExist, "checkExist", "", "Params: id=resource-id | hash=resource-hash | url=http://xxx.com/resource/url")
	flag.StringVar(&filename, "file", "", "Filename(.jpg|.png|.gif|.mp4|.avi)")
	flag.StringVar(&appID, "appID", "img", "App Id")
	flag.StringVar(&appKey, "appKey", "super", "App Key")
	flag.StringVar(&id, "id", "", "Resource Id")
	flag.StringVar(&imageProcess, "imageProcess", imageProcess, "ImageProcess, .eg: resize,fw_300,fh_200/quality,q_80")
	flag.StringVar(&fileType, "type", fileType, "file type for logical")
	flag.StringVar(&target, "target", target, "target file path")
	flag.BoolVar(&verbose, "verbose", verbose, "verbose")
	flag.DurationVar(&timeout, "timeout", time.Minute*5, "upload timeout, like: 300ms, 10s, 2m, 1h")
	flag.StringVar(&host, "host", "http://127.0.0.1:2022", "指定上传服务地址")
	flag.Parse()

	log.InitLogger(logLevel, "", true, true)
	client.InitCustomHeaderPrefix(customHeaderPrefix)
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
		res := client.CheckFileExist(host, appID, appKey, id, hash, url, timeout)
		responseDebug(res, verbose)
		return
	}
	if filename != "" {
		res := client.PostFile2UploadSvr(host, filename, appID, appKey, id, target, fileType, imageProcess, timeout, isTest)
		responseDebug(res, verbose)
	}

}
