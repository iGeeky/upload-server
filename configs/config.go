package configs

import (
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/iGeeky/open-account/pkg/baselib/log"

	"gopkg.in/yaml.v2"
)

type StorageConfig struct {
	EndPoint  string `yaml:"end_point"`
	AccessID  string `yaml:"access_id"`
	AccessKey string `yaml:"access_key"`
	Bucket    string `yaml:"bucket"`
	UrlPrefix string `yaml:"url_prefix"` //上传后的URL前缀(如果是通过CDN访问, 需要配置成CDN的域名)
}

// CORS 跨域请求配置参数
type CORS struct {
	Enable           bool   `yaml:"enable"`
	AllowOrigins     string `yaml:"allow_origins"`
	AllowMethods     string `yaml:"allow_methods"`
	AllowHeaders     string `yaml:"allow_headers"`
	AllowCredentials bool   `yaml:"allow_credentials"`
	MaxAge           int    `yaml:"max_age"`
}

// DatabaseConfig 数据库配置.
type DatabaseConfig struct {
	Dialect      string `yaml:"dialect"`
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	Database     string `yaml:"database"`
	User         string `yaml:"user"`
	Password     string `yaml:"password"`
	MaxOpenConns int    `yaml:"max_open_conns"` //最大链接数.
	MaxIdleConns int    `yaml:"max_idle_conns"` //最大闲置链接数
	Debug        bool
}

// ToURL 获取dns链接
func (p *DatabaseConfig) ToURL() (url string) {
	if p.Dialect == "postgres" {
		url = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			p.Host, p.Port, p.User, p.Password, p.Database)
	} else {
		url = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			p.User, p.Password, p.Host, p.Port, p.Database)
	}
	return
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Listen            string
	LogLevel          string `yaml:"log_level"`
	LogFileName       string `yaml:"log_filename"`
	CheckSign         bool   `yaml:"check_sign"`
	Debug             bool   `yaml:"debug"`
	DisableStacktrace bool   `yaml:"disable_stacktrace"`
	URLSign           bool   `yaml:"url_sign"`
	ServerEnv         string `yaml:"server_env"`

	// 超级签名, 仅用于测试, 只在Debug=true时有效
	SuperSignKey string `yaml:"super_sign_key"`

	ReqDebug     bool          `yaml:"req_debug"`      //是否开启请求日志
	ReqDebugHost string        `yaml:"req_debug_host"` //host值, 默认为: https://127.0.0.1:2021
	ReqDebugDir  string        `yaml:"req_debug_dir"`  //目录值, 默认为: /data/logs/req_debug/
	FdExpiretime time.Duration `yaml:"fd_expiretime"`  //文件句柄过期时间,默认为10分钟.

	UploadDB *DatabaseConfig `yaml:"upload_database"` //mysql 链接配置.

	CORS CORS `yaml:"cors"`

	ChunkTmpDir string //临时文件存储目录
	ChunkSize   int    //标准块大小

	RequestPerSecond int           //爬虫每秒请求数。
	MaxTasks         int           //爬虫最大的同时下载数
	DownloadTimeout  time.Duration //下载文件的超时时间。
	LocalDir         string        //下载的文件的临时存储目录
	ImageQuality     int           //jpeg压缩质量.

	// appID及对应的sign key
	AppKeys     map[string]string `yaml:"app_keys"`
	SignHeaders []string          `yaml:"sign_headers"` //需要参与签名的请求头列表

	CustomHeaderPrefix string `yaml:"custom_header_prefix"` //自定义请求头的前缀(默认为X-OA-)

	// 默认的存储配置.
	StorageDef StorageConfig `yaml:"storage_def"`
	// appid对应的存储配置，当找不到时，使用默认的存储配置。
	AppStorages map[string]*StorageConfig `yaml:"storages"`
}

// Config 全局配置
var Config *ServerConfig

// GetSignKey 获取应用的signKey
func (a *ServerConfig) GetSignKey(appID string) string {
	return a.AppKeys[appID]
}

// LoadConfig 加载解析配置.
func LoadConfig(configFilename string) (config *ServerConfig, err error) {
	Config = &ServerConfig{
		Listen:       ":2021",
		LogLevel:     "debug",
		LogFileName:  "./logs/upload-server.log",
		CheckSign:    true,
		Debug:        false,
		URLSign:      false,
		ServerEnv:    "dev",
		ReqDebug:     true,
		ReqDebugHost: "http://127.0.0.1:2022",
		ReqDebugDir:  "/data/logs/req_debug/",
		FdExpiretime: time.Minute * 10,

		CustomHeaderPrefix: "X-OA-",

		UploadDB: &DatabaseConfig{
			Dialect:      "mysql",
			Host:         "127.0.0.1",
			Port:         3306,
			Database:     "upload",
			User:         "upload",
			Password:     "123456",
			Debug:        true,
			MaxOpenConns: 50,
			MaxIdleConns: 20,
		},

		ChunkTmpDir:      "/data/upload/chunk/",
		ChunkSize:        1024 * 512,
		DownloadTimeout:  time.Duration(time.Second * 65),
		RequestPerSecond: 5,
		MaxTasks:         5,
		LocalDir:         "/data/upload/local",
		ImageQuality:     80,
	}
	config = Config

	config.AppKeys = make(map[string]string)

	var buf []byte
	buf, err = ioutil.ReadFile(configFilename)
	if err != nil {
		log.Fatalf("LoadConfig(%s) failed! err: %v", configFilename, err)
		return
	}
	err = yaml.Unmarshal(buf, Config)
	if err != nil {
		log.Fatalf("Unmarshal yaml config failed! err: %v", configFilename, err)
		return
	}

	if !strings.HasSuffix(Config.ChunkTmpDir, "/") {
		Config.ChunkTmpDir = Config.ChunkTmpDir + "/"
	}

	return
}
