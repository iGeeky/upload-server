package storage

import (
	// "fmt"
	"io"

	alioss "github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/iGeeky/open-account/pkg/baselib/log"
)

/**
OssClient OSS开通Region和Endpoint对照表: https://help.aliyun.com/document_detail/31837.html
华南 1 (深圳)	oss-cn-shenzhen	oss-cn-shenzhen.aliyuncs.com	oss-cn-shenzhen-internal.aliyuncs.com
**/
type OssClient struct {
	EndPoint, AccessID, AccessKey string
	Bucket                        string
	oss                           *alioss.Client
	bucket                        *alioss.Bucket
}

func NewOSS(EndPoint, AccessID, AccessKey, Bucket string) (*OssClient, error) {
	ossClient, err := alioss.New(EndPoint, AccessID, AccessKey)

	bucket, err := ossClient.Bucket(Bucket)
	if err != nil {
		log.Errorf("Get Bucket failed! err: %s", err)
		return nil, err
	}
	log.Infof("open bucket(%s) ...", Bucket)

	ossclient := &OssClient{EndPoint, AccessID, AccessKey, Bucket, ossClient, bucket}

	return ossclient, nil
}

// https://help.aliyun.com/document_detail/32147.html?spm=5176.doc31848.6.420.8SLZsH

func (client *OssClient) PutFile(key string, reader io.Reader, objectSize int64, contentType string) (err error) {
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	options := []alioss.Option{
		alioss.ObjectACL(alioss.ACLPublicRead),
		alioss.Meta("Content-Type", contentType),
	}

	return client.bucket.PutObject(key, reader, options...)
}

func (client *OssClient) FileExist(key string) (bool, error) {
	return client.bucket.IsObjectExist(key)
}

func (client *OssClient) DelFile(key string) (err error) {
	err = client.bucket.DeleteObject(key)
	return
}

func (client *OssClient) RenameFile(key string, key_new string) (err error) {
	_, err = client.bucket.CopyObject(key, key_new)
	return
}
