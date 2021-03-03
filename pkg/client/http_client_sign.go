package client

import (
	"fmt"
	"net/url"
	"time"

	"github.com/iGeeky/open-account/pkg/baselib/ginplus"
	"github.com/iGeeky/open-account/pkg/baselib/net"
)

func ParseArgs(uri string) (args map[string]string) {
	u, err := url.Parse(uri)
	if err != nil {
		return
	}
	values, err := url.ParseQuery(u.RawQuery)
	if err == nil {
		args = make(map[string]string)
		for k, arrValue := range values {
			args[k] = arrValue[0] // 有多个参数的, 只取第一个.所以请不要传入多个相同的参数, 会导致签名错误.
		}
		return
	}
	return
}

func (client *UploadClient) HttpPostWithSign(uri string, body []byte, headers map[string]string, timeout time.Duration) *net.OkJson {
	signature, SignStr := "", ""
	if client.AppKey != "" {
		args := ParseArgs(uri)
		signature, SignStr = ginplus.SignSimple(uri, args, headers, body, client.AppKey)
		headers[client.CustomHeaderName("SIGN")] = signature
	}

	res := net.HttpPostJson(uri, body, headers, timeout)

	if res.StatusCode == 401 {
		fmt.Printf("signature [%s] SigStr [[\n%s\n]]", signature, SignStr)
	}
	return res
}

func (client *UploadClient) HttpGetWithSign(uri string, headers map[string]string, timeout time.Duration) *net.OkJson {
	signature, SignStr := "", ""
	if client.AppKey != "" {
		args := ParseArgs(uri)
		signature, SignStr = ginplus.SignSimple(uri, args, headers, nil, client.AppKey)
		headers[client.CustomHeaderName("SIGN")] = signature
	}

	res := net.HttpGetJson(uri, headers, timeout)

	if res.StatusCode == 401 {
		fmt.Printf("signature [%s] SigStr [[\n%s\n]]", signature, SignStr)
	}
	return res
}
