
## Introductions

一个通用文件上传服务. 文件存储后端支持: 亚马逊S3/阿里云OSS/MinIO.


## Technologies

* 语言: golang
* Web框架: gin
* 数据库: MySQL


## Features

* [x] 支持直接上传文件
* [x] 支持URL方式上传.
* [x] 支持对大文件的分块上传.
* [ ] 支持图片处理
  * [ ] 裁剪
  * [ ] 缩放
* [x] 存储后端支持:
  * [x] S3
  * [x] OSS
  * [x] MinIO
## API Document

[API文档](./docs/api.md)

## License

[MIT](./LICENSE)
