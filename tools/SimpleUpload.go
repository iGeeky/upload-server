package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/iGeeky/open-account/pkg/baselib/net"
)

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
	var timeout time.Duration
	isTest := false

	flag.StringVar(&filename, "file", "test.jpg", "Filename(.jpg|.png|.gif|.mp4|.avi)")
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

	res := net.PostFile2UploadSvr(host, filename, appID, appKey, id, target, fileType, imageProcess, timeout, isTest)
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
		fmt.Println(res.Data["url"])
	}

}
