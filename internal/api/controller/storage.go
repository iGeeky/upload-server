package controller

import (
	"io"
	"io/ioutil"
	"os"
	"path"

	"github.com/iGeeky/open-account/pkg/baselib/errors"
	"github.com/iGeeky/open-account/pkg/baselib/log"
	"github.com/iGeeky/open-account/pkg/baselib/utils"
	"github.com/iGeeky/upload-server/configs"
	"github.com/iGeeky/upload-server/internal/api/storage"
)

type StorageConfigEx struct {
	config   *configs.StorageConfig
	instance storage.Storage
}

var storageDefInstance StorageConfigEx
var storageInstances map[string]*StorageConfigEx

func isExist(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func storageInit(storageConfig *configs.StorageConfig) storage.Storage {
	log.Infof("storageConfig.EndPoint: [%s]", storageConfig.EndPoint)
	storageInstace, err := storage.New(storageConfig.EndPoint, storageConfig.AccessID, storageConfig.AccessKey, storageConfig.Bucket)
	if err != nil {
		log.Fatalf("storage.New('%s', '%s', '******',  '%s') failed! err: %s", storageConfig.EndPoint, storageConfig.AccessID, storageConfig.Bucket, err)
		os.Exit(2)
	}
	log.Infof("storage.New(EndPoint: %s, AccessID: %s, Bucket: %s) success.",
		storageConfig.EndPoint, storageConfig.AccessID, storageConfig.Bucket)
	return storageInstace
}

func getStorageEx(appID string) *StorageConfigEx {
	storageInstace := storageInstances[appID]
	if storageInstace == nil {
		storageInstace = &storageDefInstance
	}
	return storageInstace
}

func InitUpload(isTest bool) {
	config := configs.Config

	storageDefInstance.config = &config.StorageDef
	storageDefInstance.instance = storageInit(storageDefInstance.config)
	storageInstances = make(map[string]*StorageConfigEx)

	for appID, storageConfig := range config.AppStorages {
		var storageInstace StorageConfigEx
		storageInstace.config = storageConfig
		storageInstace.instance = storageInit(storageInstace.config)
		storageInstances[appID] = &storageInstace
	}
}

func SaveFileToLocal(filename string, body []byte) error {
	fullfilename := path.Join(configs.Config.LocalDir, filename)
	log.Infof("write file [%s] to disk...", fullfilename)
	if isExist(fullfilename) {
		log.Infof("file [%s] is exist in disk!", filename)
	} else {
		dir := path.Dir(fullfilename)
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			log.Errorf("MkdirAll(%s) failed! err: %v", dir, err)
			return err
		}
		err = ioutil.WriteFile(fullfilename, body, os.ModePerm)
		if err != nil {
			log.Errorf("write file [%s] failed! err: %v", filename, err)
			return err
		}
		log.Infof("write file [%s] to disk success", fullfilename)
	}
	return nil
}

func saveToCloud(appID string, hash, filename, contentType string, bodyReader io.Reader, bodyLen int64) (url, errmsg string) {
	storageEx := getStorageEx(appID)
	storageclient := storageEx.instance
	storageconfig := storageEx.config
	defer utils.Elapsed("SaveToOSS:" + hash)()

	exist, err := storageclient.FileExist(filename)
	if err != nil {
		log.Errorf("FileExist(%s) failed! err: %s", filename, err)
		errmsg = errors.ErrServerError
		return
	}
	if !exist {
		log.Infof("put file [%s] to storage...", filename)
		err = storageclient.PutFile(filename, bodyReader, bodyLen, contentType)
		if err != nil {
			log.Errorf("PutFile failed! err: %s", err)
			errmsg = errors.ErrServerError
			return
		}
		log.Infof("put file [%s] to storage succ", filename)
	} else {
		//文件已经存在，不用处理了。
		log.Infof("**** file [%s] is in storage server! ", filename)
	}

	url = storageconfig.UrlPrefix + filename
	return
}
