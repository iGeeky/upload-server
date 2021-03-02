package utils

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"

	blutils "github.com/iGeeky/open-account/pkg/baselib/utils"
)

// SimpleHashFile 简易Hash算法.
func SimpleHashFile(path string) (string, int64, error) {
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		return "", 0, err
	}

	fi, err := file.Stat()
	if err != nil {
		return "", 0, err
	}

	return SimpleHash(file, fi.Size())
}

// SimpleHash 简易HASH算法: 小于60K的文件，计算整个文件的hash,大于60K的文件，计算文件头20K及文件中部20K和文件尾部20K的数据。
func SimpleHash(reader io.ReadSeeker, filesize int64) (string, int64, error) {
	var simpleMD5Block = int64(1024 * 20)
	if filesize <= simpleMD5Block*3 {
		return blutils.Md5Reader(reader)
	}

	h := md5.New()
	_, err := io.CopyN(h, reader, simpleMD5Block)
	if err != nil {
		return "", 0, err
	}
	//中部前端
	halfSize := int64(filesize / 2)
	halfBlockSize := int64(simpleMD5Block / 2)
	_, err = reader.Seek(halfSize-halfBlockSize, io.SeekStart)
	if err != nil {
		return "", 0, err
	}
	_, err = io.CopyN(h, reader, simpleMD5Block)
	if err != nil {
		return "", 0, err
	}
	//尾部
	_, err = reader.Seek(simpleMD5Block*-1, io.SeekEnd)
	if err != nil {
		return "", 0, err
	}

	_, err = io.CopyN(h, reader, simpleMD5Block)
	if err != nil {
		return "", 0, err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), filesize, nil
}
