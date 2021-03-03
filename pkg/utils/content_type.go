package utils

// [Content-Type] = {文件扩展名，路径前缀}
var gContentType map[string][]string
var gExtContentTypes map[string]string

func init() {
	gContentType = make(map[string][]string)
	// 图片
	gContentType["image/gif"] = []string{"gif", "img"}
	gContentType["image/jpeg"] = []string{"jpeg", "img"}
	gContentType["image/jpg"] = []string{"jpg", "img"}
	gContentType["image/png"] = []string{"png", "img"}
	gContentType["image/x-png"] = []string{"png", "img"}
	gContentType["image/x-png"] = []string{"png", "img"}
	gContentType["image/bmp"] = []string{"bmp", "img"}

	// 视频
	gContentType["video/mp4"] = []string{"mp4", "video"}
	gContentType["video/x-matroska"] = []string{"mkv", "video"}
	gContentType["video/x-msvideo"] = []string{"avi", "video"}
	gContentType["video/3gpp"] = []string{"3gp", "video"}
	gContentType["video/x-flv"] = []string{"flv", "video"}
	gContentType["video/mpeg"] = []string{"mpg", "video"}
	gContentType["video/quicktime"] = []string{"mov", "video"}
	gContentType["video/x-ms-wmv"] = []string{"wmv", "video"}
	gContentType["application/vnd.apple.mpegurl"] = []string{"m3u8", "video"}

	//音频
	gContentType["audio/wav"] = []string{"wav", "audio"}

	// 文件
	gContentType["text/plain"] = []string{"txt", "file"}
	gContentType["application/octet-stream"] = []string{"", "file"}
	gContentType["application/vnd.android.package-archive"] = []string{"apk", "file"}
	gContentType["application/iphone-package-archive"] = []string{"ipa", "file"}
	gContentType["text/xml"] = []string{"plist", "file"}
	gContentType["application/pdf"] = []string{"pdf", "file"}

	gExtContentTypes = make(map[string]string)
	for contentType, varr := range gContentType {
		ext := varr[0]
		gExtContentTypes[ext] = contentType
	}
}

// GetFileTypeAndSuffix 根据contentType获取文件类型及后缀. fileType, suffix
func GetFileTypeAndSuffix(contentType string) (string, string) {
	info := gContentType[contentType]
	if len(info) == 0 {
		return "file", ""
	}

	return info[1], info[0]
}

// GetContentTypeByExt 根据后缀,获取content-type
func GetContentTypeByExt(ext string) string {
	return gExtContentTypes[ext]
}
