# 需要添加的请求头：

* X-OA-AppID: $appID
* X-OA-Sign: $Signature
* X-OA-Platform: $platform
* X-OA-Version: $client_version
* X-OA-Channel: $channel

注意: 自定义请求头的前缀`X-OA-`是默认值, 如果服务器修改了该前缀, 客户端也要相应的修改.

```
$appID为分配给应用的App ID与AppKey对应。
$Signature 根据本文描述的算法，生成的签名值。
$platform 为请求对应的平台，值可以为: ios,android,pc
$client_version 为客户端版本号
$channel 为渠道ID。
```

### 签名步骤：
##### Step1：组成签名字符串SignStr：

```
<CanonicalURI>\n
<CanonicalArgs>\n
<CanonicalHeaders>\n
<SignedHeaders>\n
<Sha1Body>\n
<AppKey>
```

其中：
* CanonicalURI URL_ENCODE编码过的URI。特别说明：/不需要转意。编码方法见：uri-encode编码一节。
* CanonicalArgs 将参数按`参数名`排序后，参数名与值之间使用=连接起来，参数之间使用&连接起来。参数名与值都需要使用uri-encode编码。
    * 如果有同名参数，需要按值的字母序排好序。如："x=a&x=c&x=a2"变成：x=a&x=a2&x=c
* CanonicalHeaders 是签名的请求头列表，其中字段名变成小写，字段值去掉左右空格。然后字段名与字段值以`:`号连接，不同的请求头之间以`\n`连接。
    * 如果包含多个的字段，值按字母序排序，然后以`,`连接。
    * 请求头不要求所有字段都列入签名范围。必须签名的字段包括：
        * 以X-开头的字段(不包含X-Sign)
* SignedHeaders 是所有签名字段列表。所有字段名变成小写，按字母序排序，然后以`;`连接。
* Sha1Body 为请求体的Sha1值(HEX)。如果请求体为空，值为：HEX(SHA1(""))
* AppKey 分配给应用的AppKey。

##### Step2: 签名
```
Signature=HEX(sha1(SignStr))
```

##### Step3: 添加签名请求头：X-Sign
计算好请求头之后，在请求时需要添加请求头：
X-Sign: ${Signature}

##### uri-encode 方法
```
function is_encoded(str)
    local pattern = "%%%x%x"
    if string.find(str, pattern) then
        return true
    end
    return false
end

function uri_encode(arg, encodeSlash)
    if is_encoded(arg) then
        return arg
    end
    if encodeSlash == nil then
        encodeSlash = true
    end

    local chars = {}
    for i = 1,string.len(arg) do
        local ch = string.sub(arg, i,i)
        if (ch >= 'A' and ch <= 'Z') or (ch >= 'a' and ch <= 'z') or 
            (ch >= '0' and ch <= '9') or ch == '_' or ch == '-' or     
            ch == '~' or ch == '.' then
            table.insert(chars, ch)
         elseif ch == '/' then
            if encodeSlash then
                table.insert(chars, "%2F")
            else 
                table.insert(chars, ch)
            end
        else
            table.insert(chars, string.format("%%%02X", string.byte(ch)))
        end
    end
    return table.concat(chars)
end

function URL_ENCODE(arg)
    return uri_encode(arg, false)
end
```


# 示例
-------------
#### 输入
```
------------- 示例 ---------------
app_key:16317d117c6eceb8b1b0ebb40e506617
uri:/path/test/~-_/99@/中文.doc
args:dest=mongo&DEST=MongoEx&aBo=d9&aBo=Ads&name&aBo=a09&aBo=030
headers: Host: www.test.com
Content-Type: application/text
Content-Length: 16
range: 0-1000
date: Fri, 18 Dec 2015 06:17:47 GMT
X-Token: test-token
X-AppId: test
X-rid: 001
X-FOO: Dest 
X-FOo: Ads
X-Foo: Abort
X-foo: 099

body:This is the body
```
#### 输出
------------- 结果 ----------------
```
signature: 51425c7fd23bfaca3581334b5905d5b5b5d4b1ac
signStr: [[[/path/test/~-_/99%40/%E4%B8%AD%E6%96%87.doc
DEST=MongoEx&aBo=030&aBo=Ads&aBo=a09&aBo=d9&dest=mongo&name=
x-appid:test
x-foo:099,Abort,Ads,Dest
x-rid:001
x-token:test-token
x-appid;x-foo;x-rid;x-token
8e91dd971a7b7ed3797b4794da78df4f25225377
16317d117c6eceb8b1b0ebb40e506617]]]
```
* CURL命令:

```
 curl "http://127.0.0.1:82/path/test/~-_/99@/中文.doc?dest=mongo&DEST=MongoEx&aBo=d9&aBo=Ads&name&aBo=a09&aBo=030" -d"This is the body" -H"Host: www.yf.com" -H"Content-Type: application/text" -H"Content-Length: 16" -H"range: 0-1000" -H"date: Fri, 18 Dec 2015 06:17:47 GMT" -H"X-Token: test-token" -H"X-AppId: test" -H"X-rid: 001" -H"X-FOO: Dest " -H"X-FOo: Ads" -H"X-Foo: Abort" -H"X-foo: 099" -H"X-Sign: 719d7d49eebc533d3480d25685199d38fdc430ef"
```