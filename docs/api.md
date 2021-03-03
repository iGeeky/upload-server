
# 签名AppKey

不同的应用,不同的资源，使用不同的appID及appKey, 具体AppID及AppKey请咨询后端开发.

# 协议说明

	请求体采用POST，直接将文件内容以二进制发送到服务器上。
	响应体均采用json格式。响应格式为ok_json。
	请求成功返回200,失败返回4xx,5xx并返回相应的JSON。

### ok_json格式说明：

ok_json主要有三个字段，ok, reason, data。其中：

* ok 为true表示服务成功。ok为false表示服务失败。
* reason中是失败原因(当ok=false时)，或为空(ok=true时)。
* data 为接口返回的数据，具体接口会不一样。可为空。
* 示例：

```
{"ok": true,
"reason": "失败原因",
"data": {}}
```

以下接口中，响应数据只列出data部分，ok_json部分不再列出。

* 公共reason值：
    * ERR_ARGS_INVALID 请求参数错误：缺少字段，或字段值不对。
    * ERR_SERVER_ERROR 服务器内部错误：如访问数据库出错了
    * ERR_SIGN_ERROR 签名错误

### SimpleHash
* 由于文件太大时，浏览器计算hash容易崩溃。SimpleHash采用以下规则计算Hash:
	* 文件大小 <= 60K 计算整个文件的md5
	* 文件大于 > 60K 计算文件头20K, 文件中部(offset=filesize/2-10*1024)20K, 及文件末尾的20K加一起的md5. 


# 接口设计

### 支持的Content-Type值：
```
image/gif
image/jpeg
image/jpg
image/png
image/x-png
image/x-png
image/bmp
video/mp4
video/x-matroska
video/x-msvideo
video/3gpp
video/x-flv
video/mpeg
video/quicktime
video/x-ms-wmv
application/vnd.apple.mpegurl
audio/wav
text/plain
application/octet-stream
application/vnd.android.package-archive
application/iphone-package-archive
text/xml
application/pdf
```

### 注意

协议中自定义请求头,默认使用前缀X-OA-, 如果服务器修改过该前缀, 客户端也应该做相应的修改. 以下文档中都会写成X-OA-开头.

### 检查文件是否存在

* URL: GET　/v1/upload/check_exist
* 是否需要签名：是,[签名方法](./Sign.md)
* 是否需要登录：否
* 请求Body: 文件内容。
* 请求头(RequestHeader)：
	* Content-Type: 需要根据Mime规范设置
	* X-OA-URL: $url, 查找指定URL资源是否存在.
	* X-OA-Hash: $filehash, 查找指定HASH资源是否存在.
	* X-OA-ID: $id, 查找指定ID的资源是否存在.

以上请求头中, `X-OA-URL`,`X-OA-Hash`, `X-OA-ID` 只需要输入一个即可.

* 响应(OK_JSON)
	* 请求成功，文件不存在，返回: {ok: true, reason: "NOT-EXIST"}
	* 请求失败，返回：{ok: false, reason: "失败原因"}

字段 | 类型 | 是否可为空 | 说明
--- | --- | ------- | ----
url | string | 否 | 文件URL.
rid | string | 否 | 文件的Rid

* reason值：
	* ERR_HASH_INVALID 文件HASH不匹配。
	* ERR_CONTENT_TYPE_INVALID 不支持的Content-Type值。

### 小文件上传接口

* URL: POST　/v1/upload/simple
* 是否需要签名：是,[签名方法](./Sign.md)
* 是否需要登录：否
* 请求Body: 文件内容。
* 请求头(RequestHeader)：
	* X-OA-Hash: $filehash `文件的sha1值`
	* X-OA-Type: $resourceType `资源的业务类型, 类型类型会作为存储路径的一部分, 可为空`
	* X-OA-ID: $id `资源在业务系统中的ID, 可以是图片ID,文件ID. 可为空`
	* X-OA-Target: $target `目标存储路径, 客户端指定存储路径, 一般建议为空, 服务器根据hash生成路径.`
	* X-OA-ImageProcess: $imageProcess `图片处理. 支持: resize,fw_300,fh_200/quality,q_80`
	* Content-Type: 需要根据Mime规范设置.

* 响应(OK_JSON)

字段 | 类型 | 是否可为空 | 说明
--- | --- | ------- | ----
url | string | 否 | 文件URL.
rid | string | 否 | 文件的Rid

* reason值：
	* ERR_HASH_INVALID 文件HASH不匹配。
	* ERR_CONTENT_TYPE_INVALID 不支持的Content-Type值。

### 通过URL上传文件.

* URL: POST　/v1/upload/url
* 是否需要签名：是,[签名方法](./Sign.md)
* 是否需要登录：否
* 请求头(RequestHeader)：
	* X-OA-Type: $resourceType `资源的业务类型, 类型类型会作为存储路径的一部分, 可为空`
	* X-OA-ID: $id `资源在业务系统中的ID, 可以是图片ID,文件ID. 可为空`
	* X-OA-Target: $target `目标存储路径, 客户端指定存储路径, 一般建议为空, 服务器根据hash生成路径.`
	* X-OA-ImageProcess: $imageProcess `图片处理. 支持: resize,fw_300,fh_200/quality,q_80`
	* Content-Type: 需要根据Mime规范设置.
* 请求Body: POST + JSON

字段 | 类型 | 是否可为空 | 说明
--- | --- | ------ | ---
url | string | 否 | URL
referer | string | 是 | 请求referer
userAgent | string | 是 | 请求userAgent

* 响应(OK_JSON)

字段 | 类型 | 是否可为空 | 说明
--- | --- | ------- | ----
url | string | 否 | 上传的文件URL
rid | string | 否 | 文件的Rid

### 文件上传(分块上传)
分块上传中，请求Body中存储文件内容，附加信息统一通过HTTP请求头传递。

##### 上传初始化
初始化，或检查文件(HASH)是否上传过，上传了多少块。

* URL: POST /v1/upload/chunk/init
* 是否需要签名：是,[签名方法](./Sign.md)
* 是否需要登录：否
* 请求Body: 空
* 请求头(RequestHeader)：
	* X-OA-Hash: $fileHash 文件HASH, [SimpleHash](#SimpleHash)
	* X-OA-FileSize: $fileSize, 文件大小
    * X-OA-ChunkSize: $chunkSize, 块大小, 可以不传, 不传时使用服务器配置大小.
	* X-OA-Type: $resourceType `资源的业务类型, 类型类型会作为存储路径的一部分, 可为空`
	* X-OA-ID: $id `资源在业务系统中的ID, 可以是图片ID,文件ID. 可为空`
	* Content-Type: 需要根据Mime规范设置.
* 响应(OK_JSON)

字段名|类型|是否可为空| 备注
----|----|----|----
completedChunks | list(int) | Y　| 已经上传成功的块ID列表
notCompletedChunks | list(int) | Y　| 未上传的块ID列表
url | string | Y | 资源URL
rid | string | Y | 资源ID
chunkSize | int | Y | 块大小（后面按该大小分块上传，不同的文件，块大小可能不同）

该接口返回值分三种情况：
假设文件有6块。
* 从未上传过的文件：
	* completedChunks 为空
	* notCompletedChunks 为所有块的列表，如：0,1,2,3,4,5
	* url 空
	* rid 空
* 上传过一部分块的文件：
	* completedChunks 返回已经上传的块列表，如 0,1,2
	* notCompletedChunks 为所有块的列表，如：3,4,5
	* url 空
	* rid 空
* 上传完成的文件
	* completedChunks 返回已经上传的块列表，如 0,1,2,3,4,5
	* notCompletedChunks 为空。
	* url http://res.host.com/path/to/file.jpg
	* rid VX2432342342

##### 文件块内容上传

* URL: POST /v1/upload/chunk/upload
* 是否需要签名：是,[签名方法](./Sign.md)
* 是否需要登录：否
* 请求Body: 该块的内容。
* 请求头(RequestHeader)：
	* X-OA-Hash: $fileHash 文件HASH, [SimpleHash](#SimpleHash)
	* X-OA-FileSize: $fileSize, 文件大小
    * X-OA-ChunkSize: $chunkSize, 块大小, 可以不传, 不传时使用服务器配置大小.
    * X-OA-ChunkIndex: $chunkIndex, 块索引, 从0开始.
    * X-OA-ChunkHash: $chunkHash, 块Hash值得, (SHA1)
	* X-OA-Type: $resourceType `资源的业务类型, 类型类型会作为存储路径的一部分, 可为空`
	* X-OA-ID: $id `资源在业务系统中的ID, 可以是图片ID,文件ID. 可为空`
	* Content-Type: 需要根据Mime规范设置.
* 响应(OK_JSON)

字段名|类型|是否可为空| 备注
----|----|----|----
completedChunks | list(int) | Y　| 已经上传成功的块ID列表
notCompletedChunks | list(int) | Y　| 未上传的块ID列表
url | string | Y | 资源URL
rid | string | Y | 资源ID
