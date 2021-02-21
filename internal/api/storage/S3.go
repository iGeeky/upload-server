package storage

import (
	"io"
	"strings"

	"github.com/iGeeky/open-account/pkg/baselib/log"

	minio "github.com/minio/minio-go"
)

type S3Client struct {
	EndPoint, AccessID, AccessKey string
	Bucket                        string
	client                        *minio.Client
}

func NewS3(endPoint, accessID, accessKey, bucket string) (client *S3Client, err error) {
	minioClient, err := minio.New(endPoint, accessID, accessKey, false)
	if err != nil {
		log.Errorf("minio.New failed! err: %v", err)
		return
	}

	location := ""
	err = minioClient.MakeBucket(bucket, location)
	if err != nil {
		exists, err2 := minioClient.BucketExists(bucket)
		if err2 != nil || !exists {
			err = err2
			log.Errorf("MakeBucket(bucket:%s, location:%s) failed! err: %s", bucket, location, err)
			return
		}
		err = nil
	}
	log.Infof("open bucket(%s) ...", bucket)

	client = &S3Client{endPoint, accessID, accessKey, bucket, minioClient}

	return
}

func (client *S3Client) PutFile(key string, reader io.Reader, objectSize int64, contentType string) (err error) {
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	opts := minio.PutObjectOptions{
		ContentType: contentType,
	}

	if objectSize == 0 {
		objectSize = -1
	}

	n, err := client.client.PutObject(client.Bucket, key, reader, objectSize, opts)
	if err != nil {
		log.Errorf("client.PutObject(Bucket: %s, key: %s) failed! err:", client.Bucket, key)
	} else if n != int64(objectSize) {
		log.Errorf("client.PutObject(Bucket: %s, key: %s) failed! writedSize(%d) != objectSize(%d)", client.Bucket, key, n, objectSize)
	}
	return
}

func (client *S3Client) FileExist(key string) (exist bool, err error) {
	// return client.bucket.IsObjectExist(key)
	opts := minio.StatObjectOptions{}
	var info minio.ObjectInfo
	info, err = client.client.StatObject(client.Bucket, key, opts)
	// log.Infof("info: %+v, err: %v", info, err)
	exist = err == nil && info.Key != ""

	if err != nil && strings.Contains(err.Error(), "not exist") {
		err = nil
	}

	return
}

func (client *S3Client) DelFile(key string) (err error) {
	err = client.client.RemoveObject(client.Bucket, key)
	return
}

func (client *S3Client) RenameFile(key string, newKey string) (err error) {
	src := minio.NewSourceInfo(client.Bucket, key, nil)

	// Destination object
	dst, err := minio.NewDestinationInfo(client.Bucket, newKey, nil, nil)
	if err != nil {
		log.Errorf("NewDestinationInfo(%s, %s) failed! err: %v", client.Bucket, newKey)
		return
	}

	// Copy object call
	err = client.client.CopyObject(dst, src)
	if err != nil {
		log.Errorf("CopyObject(dst:%s, src:%s) failed! err: %v", newKey, key, err)
		return
	}

	return
}
