package oss

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	. "gopkg.in/check.v1"
)

type OssConnSuite struct{}

var _ = Suite(&OssConnSuite{})

func (s *OssConnSuite) TestURLMarker(c *C) {
	um := urlMaker{}
	um.Init("docs.github.com", true, false)
	c.Assert(um.Type, Equals, urlTypeCname)
	c.Assert(um.Scheme, Equals, "http")
	c.Assert(um.NetLoc, Equals, "docs.github.com")

	c.Assert(um.getURL("bucket", "object", "params").String(), Equals, "http://docs.github.com/object?params")
	c.Assert(um.getURL("bucket", "object", "").String(), Equals, "http://docs.github.com/object")
	c.Assert(um.getURL("", "object", "").String(), Equals, "http://docs.github.com/object")

	var conn Conn
	conn.config = getDefaultOssConfig()
	conn.config.AuthVersion = AuthV1
	c.Assert(conn.getResource("bucket", "object", "subres"), Equals, "/bucket/object?subres")
	c.Assert(conn.getResource("bucket", "object", ""), Equals, "/bucket/object")
	c.Assert(conn.getResource("", "object", ""), Equals, "/")

	um.Init("https://docs.github.com", true, false)
	c.Assert(um.Type, Equals, urlTypeCname)
	c.Assert(um.Scheme, Equals, "https")
	c.Assert(um.NetLoc, Equals, "docs.github.com")

	um.Init("http://docs.github.com", true, false)
	c.Assert(um.Type, Equals, urlTypeCname)
	c.Assert(um.Scheme, Equals, "http")
	c.Assert(um.NetLoc, Equals, "docs.github.com")

	um.Init("docs.github.com:8080", false, true)
	c.Assert(um.Type, Equals, urlTypeAliyun)
	c.Assert(um.Scheme, Equals, "http")
	c.Assert(um.NetLoc, Equals, "docs.github.com:8080")

	c.Assert(um.getURL("bucket", "object", "params").String(), Equals, "http://bucket.docs.github.com:8080/object?params")
	c.Assert(um.getURL("bucket", "object", "").String(), Equals, "http://bucket.docs.github.com:8080/object")
	c.Assert(um.getURL("", "object", "").String(), Equals, "http://docs.github.com:8080/")
	c.Assert(conn.getResource("bucket", "object", "subres"), Equals, "/bucket/object?subres")
	c.Assert(conn.getResource("bucket", "object", ""), Equals, "/bucket/object")
	c.Assert(conn.getResource("", "object", ""), Equals, "/")

	um.Init("https://docs.github.com:8080", false, true)
	c.Assert(um.Type, Equals, urlTypeAliyun)
	c.Assert(um.Scheme, Equals, "https")
	c.Assert(um.NetLoc, Equals, "docs.github.com:8080")

	um.Init("127.0.0.1", false, true)
	c.Assert(um.Type, Equals, urlTypeIP)
	c.Assert(um.Scheme, Equals, "http")
	c.Assert(um.NetLoc, Equals, "127.0.0.1")

	um.Init("http://127.0.0.1", false, false)
	c.Assert(um.Type, Equals, urlTypeIP)
	c.Assert(um.Scheme, Equals, "http")
	c.Assert(um.NetLoc, Equals, "127.0.0.1")
	c.Assert(um.getURL("bucket", "object", "params").String(), Equals, "http://127.0.0.1/bucket/object?params")
	c.Assert(um.getURL("", "object", "params").String(), Equals, "http://127.0.0.1/?params")

	um.Init("https://127.0.0.1:8080", false, false)
	c.Assert(um.Type, Equals, urlTypeIP)
	c.Assert(um.Scheme, Equals, "https")
	c.Assert(um.NetLoc, Equals, "127.0.0.1:8080")

	um.Init("http://[2401:b180::dc]", false, false)
	c.Assert(um.Type, Equals, urlTypeIP)
	c.Assert(um.Scheme, Equals, "http")
	c.Assert(um.NetLoc, Equals, "[2401:b180::dc]")

	um.Init("https://[2401:b180::dc]:8080", false, false)
	c.Assert(um.Type, Equals, urlTypeIP)
	c.Assert(um.Scheme, Equals, "https")
	c.Assert(um.NetLoc, Equals, "[2401:b180::dc]:8080")

	um.InitExt("https://docs.github.com:8080", false, false, true)
	c.Assert(um.Type, Equals, urlTypePathStyle)
	c.Assert(um.Scheme, Equals, "https")
	c.Assert(um.NetLoc, Equals, "docs.github.com:8080")
	c.Assert(um.getURL("bucket", "object", "params").String(), Equals, "https://docs.github.com:8080/bucket/object?params")
	c.Assert(um.getURL("", "object", "params").String(), Equals, "https://docs.github.com:8080/?params")

	um.InitExt("docs.github.com", false, false, true)
	c.Assert(um.Type, Equals, urlTypePathStyle)
	c.Assert(um.Scheme, Equals, "http")
	c.Assert(um.NetLoc, Equals, "docs.github.com")

	c.Assert(um.getURL("bucket", "object", "params").String(), Equals, "http://docs.github.com/bucket/object?params")
	c.Assert(um.getURL("bucket", "object", "").String(), Equals, "http://docs.github.com/bucket/object")
	c.Assert(um.getURL("", "object", "").String(), Equals, "http://docs.github.com/")
}

func (s *OssConnSuite) TestAuth(c *C) {
	endpoint := "https://github.com/"
	cfg := getDefaultOssConfig()
	cfg.AuthVersion = AuthV1
	um := urlMaker{}
	um.Init(endpoint, false, false)
	conn := Conn{cfg, &um, nil}
	uri := um.getURL("bucket", "object", "")
	req := &http.Request{
		Method:     "PUT",
		URL:        uri,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     make(http.Header),
		Host:       uri.Host,
	}

	req.Header.Set("Content-Type", "text/html")
	req.Header.Set("Date", "Thu, 17 Nov 2005 18:49:58 GMT")
	req.Header.Set("Host", endpoint)
	req.Header.Set("X-OSS-Meta-Your", "your")
	req.Header.Set("X-OSS-Meta-Author", "foo@bar.com")
	req.Header.Set("X-OSS-Magic", "abracadabra")
	req.Header.Set("Content-Md5", "ODBGOERFMDMzQTczRUY3NUE3NzA5QzdFNUYzMDQxNEM=")

	var akIf Credentials
	credProvider := conn.config.CredentialsProvider
	akIf = credProvider.(CredentialsProvider).GetCredentials()
	conn.signHeader(req, conn.getResource("bucket", "object", ""), akIf)
	testLogger.Println("AUTHORIZATION:", req.Header.Get(HTTPHeaderAuthorization))
}

func (s *OssConnSuite) TestAuthV1Header(c *C) {
	endpoint := "http://oss-cn-hangzhou.aliyuncs.com/"
	cfg := getDefaultOssConfig()
	cfg.AuthVersion = AuthV1
	cfg.AccessKeyID = "ak"
	cfg.AccessKeySecret = "sk"
	um := urlMaker{}
	um.Init(endpoint, false, false)
	conn := Conn{cfg, &um, nil}
	uri := um.getURL("examplebucket", "", "")
	req := &http.Request{
		Method:     "PUT",
		URL:        uri,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     make(http.Header),
		Host:       uri.Host,
	}
	req.Header.Add("Content-MD5", "eB5eJF1ptWaXm4bijSPyxw==")
	req.Header.Add("Content-Type", "text/html")
	req.Header.Add("x-oss-meta-author", "alice")
	req.Header.Add("x-oss-meta-magic", "abracadabra")
	req.Header.Add("x-oss-date", "Wed, 28 Dec 2022 10:27:41 GMT")
	req.Header.Add("Date", "Wed, 28 Dec 2022 10:27:41 GMT")
	var akIf Credentials
	credProvider := conn.config.CredentialsProvider
	akIf = credProvider.(CredentialsProvider).GetCredentials()
	conn.signHeader(req, conn.getResource("examplebucket", "nelson", ""), akIf)
	c.Assert("OSS ak:kSHKmLxlyEAKtZPkJhG9bZb5k7M=", Equals, req.Header.Get(HTTPHeaderAuthorization))

	uri = um.getURL("examplebucket", "", "acl")
	req = &http.Request{
		Method:     "PUT",
		URL:        uri,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     make(http.Header),
		Host:       uri.Host,
	}
	req.Header.Add("Content-MD5", "eB5eJF1ptWaXm4bijSPyxw==")
	req.Header.Add("Content-Type", "text/html")
	req.Header.Add("x-oss-meta-author", "alice")
	req.Header.Add("x-oss-meta-magic", "abracadabra")
	req.Header.Add("x-oss-date", "Wed, 28 Dec 2022 10:27:41 GMT")
	req.Header.Add("Date", "Wed, 28 Dec 2022 10:27:41 GMT")
	conn.signHeader(req, conn.getResource("examplebucket", "nelson", "acl"), akIf)
	c.Assert("OSS ak:/afkugFbmWDQ967j1vr6zygBLQk=", Equals, req.Header.Get(HTTPHeaderAuthorization))

	uri = um.getURL("examplebucket", "", "resourceGroup&non-resousce=null")
	req = &http.Request{
		Method:     "GET",
		URL:        uri,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     make(http.Header),
		Host:       uri.Host,
	}
	req.Header.Add("x-oss-date", "Wed, 28 Dec 2022 10:27:41 GMT")
	req.Header.Add("Date", "Wed, 28 Dec 2022 10:27:41 GMT")
	conn.signHeader(req, conn.getResource("examplebucket", "", "resourceGroup"), akIf)
	c.Assert("OSS ak:vkQmfuUDyi1uDi3bKt67oemssIs=", Equals, req.Header.Get(HTTPHeaderAuthorization))

	uri = um.getURL("examplebucket", "", "resourceGroup&acl")
	req = &http.Request{
		Method:     "GET",
		URL:        uri,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     make(http.Header),
		Host:       uri.Host,
	}
	req.Header.Add("x-oss-date", "Wed, 28 Dec 2022 10:27:41 GMT")
	req.Header.Add("Date", "Wed, 28 Dec 2022 10:27:41 GMT")
	params := map[string]interface{}{}
	params["resourceGroup"] = nil
	params["acl"] = nil
	sub := conn.getSubResource(params)
	conn.signHeader(req, conn.getResource("examplebucket", "", sub), akIf)
	c.Assert("OSS ak:x3E5TgOvl/i7PN618s5mEvpJDYk=", Equals, req.Header.Get(HTTPHeaderAuthorization))
}

func (s *OssConnSuite) TestAuthV1Query(c *C) {
	endpoint := "http://oss-cn-hangzhou.aliyuncs.com/"
	cfg := getDefaultOssConfig()
	cfg.AuthVersion = AuthV1
	cfg.AccessKeyID = "ak"
	cfg.AccessKeySecret = "sk"
	um := urlMaker{}
	um.Init(endpoint, false, false)
	conn := Conn{cfg, &um, nil}
	uri := um.getURL("bucket", "key", "versionId=versionId")
	req := &http.Request{
		Method:     "GET",
		URL:        uri,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     make(http.Header),
		Host:       uri.Host,
	}
	req.Header.Add("Date", "Sun, 12 Nov 2023 16:43:40 GMT")
	signTime, _ := http.ParseTime("Sun, 12 Nov 2023 16:43:40 GMT")
	params := map[string]interface{}{}
	params["versionId"] = "versionId"
	signUrl, err := conn.signURL("GET", "bucket", "key", signTime.Unix(), params, nil)
	c.Assert(err, IsNil)
	c.Assert("http://bucket.oss-cn-hangzhou.aliyuncs.com/key?Expires=1699807420&OSSAccessKeyId=ak&Signature=dcLTea%2BYh9ApirQ8o8dOPqtvJXQ%3D&versionId=versionId", Equals, signUrl)

	cfg.SecurityToken = "token"
	uri = um.getURL("bucket", "key+123", "versionId=versionId")
	req = &http.Request{
		Method:     "GET",
		URL:        uri,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     make(http.Header),
		Host:       uri.Host,
	}
	req.Header.Add("Date", "Sun, 12 Nov 2023 16:56:44 GMT")
	signTime, _ = http.ParseTime("Sun, 12 Nov 2023 16:56:44 GMT")
	params = map[string]interface{}{}
	params["versionId"] = "versionId"
	signUrl, err = conn.signURL("GET", "bucket", "key+123", signTime.Unix(), params, nil)
	c.Assert(err, IsNil)
	c.Assert("http://bucket.oss-cn-hangzhou.aliyuncs.com/key%2B123?Expires=1699808204&OSSAccessKeyId=ak&Signature=jzKYRrM5y6Br0dRFPaTGOsbrDhY%3D&security-token=token&versionId=versionId", Equals, signUrl)
}

func (s *OssConnSuite) TestConnToolFunc(c *C) {
	err := CheckRespCode(202, []int{})
	c.Assert(err, NotNil)

	err = CheckRespCode(202, []int{404})
	c.Assert(err, NotNil)

	err = CheckRespCode(202, []int{202, 404})
	c.Assert(err, IsNil)

	srvErr, err := serviceErrFromXML([]byte(""), 312, "")
	c.Assert(err, NotNil)
	c.Assert(srvErr.StatusCode, Equals, 0)

	srvErr, err = serviceErrFromXML([]byte("ABC"), 312, "")
	c.Assert(err, NotNil)
	c.Assert(srvErr.StatusCode, Equals, 0)

	srvErr, err = serviceErrFromXML([]byte("<Error></Error>"), 312, "")
	c.Assert(err, IsNil)
	c.Assert(srvErr.StatusCode, Equals, 312)

	unexpect := UnexpectedStatusCodeError{[]int{200}, 202}
	c.Assert(len(unexpect.Error()) > 0, Equals, true)
	c.Assert(unexpect.Got(), Equals, 202)

	fd, err := os.Open("../sample/BingWallpaper-2015-11-07.jpg")
	c.Assert(err, IsNil)
	fd.Close()
	var out ProcessObjectResult
	err = jsonUnmarshal(fd, &out)
	c.Assert(err, NotNil)
}

func (s *OssConnSuite) TestSignRtmpURL(c *C) {
	cfg := getDefaultOssConfig()

	um := urlMaker{}
	um.Init(endpoint, false, false)
	conn := Conn{cfg, &um, nil}

	//Anonymous
	channelName := "test-sign-rtmp-url"
	playlistName := "playlist.m3u8"
	expiration := time.Now().Unix() + 3600
	signedRtmpURL := conn.signRtmpURL(bucketName, channelName, playlistName, expiration)
	playURL := getPublishURL(bucketName, channelName)
	hasPrefix := strings.HasPrefix(signedRtmpURL, playURL)
	c.Assert(hasPrefix, Equals, true)

	//empty playlist name
	playlistName = ""
	signedRtmpURL = conn.signRtmpURL(bucketName, channelName, playlistName, expiration)
	playURL = getPublishURL(bucketName, channelName)
	hasPrefix = strings.HasPrefix(signedRtmpURL, playURL)
	c.Assert(hasPrefix, Equals, true)
}

func (s *OssConnSuite) TestGetRtmpSignedStr(c *C) {
	cfg := getDefaultOssConfig()
	um := urlMaker{}
	um.Init(endpoint, false, false)
	conn := Conn{cfg, &um, nil}

	akIf := conn.config.GetCredentials()

	//Anonymous
	channelName := "test-get-rtmp-signed-str"
	playlistName := "playlist.m3u8"
	expiration := time.Now().Unix() + 3600
	params := map[string]interface{}{}
	signedStr := conn.getRtmpSignedStr(bucketName, channelName, playlistName, expiration, akIf.GetAccessKeySecret(), params)
	c.Assert(signedStr, Equals, "")
}

func (s *OssConnSuite) TestAuthV4Header(c *C) {
	endpoint := "http://oss-cn-hangzhou.aliyuncs.com/"
	cfg := getDefaultOssConfig()
	cfg.AuthVersion = AuthV4
	cfg.AccessKeyID = "ak"
	cfg.AccessKeySecret = "sk"
	cfg.Region = "cn-hangzhou"
	um := urlMaker{}
	um.Init(endpoint, false, false)
	conn := Conn{cfg, &um, nil}
	uri := um.getURL("bucket", "1234+-/123/1.txt", "")
	req := &http.Request{
		Method:     "PUT",
		URL:        uri,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     make(http.Header),
		Host:       uri.Host,
	}
	req.Header.Add("x-oss-head1", "value")
	req.Header.Add("abc", "value")
	req.Header.Add("ZAbc", "value")
	req.Header.Add("XYZ", "value")
	req.Header.Add("content-type", "text/plain")
	req.Header.Add("x-oss-content-sha256", "UNSIGNED-PAYLOAD")
	signTime := time.Unix(1702743657, 0).UTC()
	req.Header.Add("x-oss-date", signTime.Format(timeFormatV4))
	req.Header.Add("Date", "Sat, 16 Dec 2023 16:20:57 GMT")

	params := map[string]interface{}{}
	params["param1"] = "value1"
	params["+param1"] = "value3"
	params["|param1"] = "value4"
	params["+param2"] = nil
	params["|param2"] = nil
	params["param2"] = nil
	sub := conn.getSubResource(params)
	var akIf Credentials
	credProvider := conn.config.CredentialsProvider
	akIf = credProvider.(CredentialsProvider).GetCredentials()
	conn.signHeader(req, conn.getResourceV4("bucket", "1234+-/123/1.txt", sub), akIf)
	c.Assert("OSS4-HMAC-SHA256 Credential=ak/20231216/cn-hangzhou/oss/aliyun_v4_request,Signature=e21d18daa82167720f9b1047ae7e7f1ce7cb77a31e8203a7d5f4624fa0284afe", Equals, req.Header.Get(HTTPHeaderAuthorization))

}

func (s *OssConnSuite) TestAuthV4HeaderToken(c *C) {
	endpoint := "http://oss-cn-hangzhou.aliyuncs.com/"
	cfg := getDefaultOssConfig()
	cfg.AuthVersion = AuthV4
	cfg.AccessKeyID = "ak"
	cfg.AccessKeySecret = "sk"
	cfg.SecurityToken = "token"
	cfg.Region = "cn-hangzhou"
	um := urlMaker{}
	um.Init(endpoint, false, false)
	conn := Conn{cfg, &um, nil}
	uri := um.getURL("bucket", "1234+-/123/1.txt", "")
	req := &http.Request{
		Method:     "PUT",
		URL:        uri,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     make(http.Header),
		Host:       uri.Host,
	}
	req.Header.Add("x-oss-head1", "value")
	req.Header.Add("abc", "value")
	req.Header.Add("ZAbc", "value")
	req.Header.Add("XYZ", "value")
	req.Header.Add("content-type", "text/plain")
	req.Header.Add("x-oss-content-sha256", "UNSIGNED-PAYLOAD")
	signTime := time.Unix(1702784856, 0).UTC()
	req.Header.Add("x-oss-date", signTime.Format(timeFormatV4))
	req.Header.Add("Date", signTime.Format(http.TimeFormat))
	params := map[string]interface{}{}
	params["param1"] = "value1"
	params["+param1"] = "value3"
	params["|param1"] = "value4"
	params["+param2"] = nil
	params["|param2"] = nil
	params["param2"] = nil
	sub := conn.getSubResource(params)
	var akIf Credentials
	credProvider := conn.config.CredentialsProvider
	akIf = credProvider.(CredentialsProvider).GetCredentials()
	if cfg.SecurityToken != "" {
		req.Header.Set(HTTPHeaderOssSecurityToken, akIf.GetSecurityToken())
	}
	conn.signHeader(req, conn.getResourceV4("bucket", "1234+-/123/1.txt", sub), akIf)
	c.Assert("OSS4-HMAC-SHA256 Credential=ak/20231217/cn-hangzhou/oss/aliyun_v4_request,Signature=b94a3f999cf85bcdc00d332fbd3734ba03e48382c36fa4d5af5df817395bd9ea", Equals, req.Header.Get(HTTPHeaderAuthorization))

}

func (s *OssConnSuite) TestAuthV4HeaderWithAdditionalHeaders(c *C) {
	endpoint := "http://oss-cn-hangzhou.aliyuncs.com/"
	cfg := getDefaultOssConfig()
	cfg.AuthVersion = AuthV4
	cfg.AccessKeyID = "ak"
	cfg.AccessKeySecret = "sk"
	cfg.Region = "cn-hangzhou"
	cfg.AdditionalHeaders = []string{"ZAbc", "abc"}
	um := urlMaker{}
	um.Init(endpoint, false, false)
	conn := Conn{cfg, &um, nil}
	uri := um.getURL("bucket", "1234+-/123/1.txt", "")
	req := &http.Request{
		Method:     "PUT",
		URL:        uri,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     make(http.Header),
		Host:       uri.Host,
	}
	req.Header.Add("x-oss-head1", "value")
	req.Header.Add("abc", "value")
	req.Header.Add("ZAbc", "value")
	req.Header.Add("XYZ", "value")
	req.Header.Add("content-type", "text/plain")
	req.Header.Add("x-oss-content-sha256", "UNSIGNED-PAYLOAD")
	signTime := time.Unix(1702747512, 0).UTC()
	req.Header.Add("x-oss-date", signTime.Format(timeFormatV4))
	req.Header.Add("Date", signTime.Format(http.TimeFormat))
	params := map[string]interface{}{}
	params["param1"] = "value1"
	params["+param1"] = "value3"
	params["|param1"] = "value4"
	params["+param2"] = nil
	params["|param2"] = nil
	params["param2"] = nil
	sub := conn.getSubResource(params)
	var akIf Credentials
	credProvider := conn.config.CredentialsProvider
	akIf = credProvider.(CredentialsProvider).GetCredentials()
	if cfg.SecurityToken != "" {
		req.Header.Set(HTTPHeaderOssSecurityToken, akIf.GetSecurityToken())
	}
	conn.signHeader(req, conn.getResourceV4("bucket", "1234+-/123/1.txt", sub), akIf)
	c.Assert("OSS4-HMAC-SHA256 Credential=ak/20231216/cn-hangzhou/oss/aliyun_v4_request,AdditionalHeaders=abc;zabc,Signature=4a4183c187c07c8947db7620deb0a6b38d9fbdd34187b6dbaccb316fa251212f", Equals, req.Header.Get(HTTPHeaderAuthorization))
}

func (s *OssConnSuite) TestAuthV4Query(c *C) {
	endpoint := "http://oss-cn-hangzhou.aliyuncs.com/"
	cfg := getDefaultOssConfig()
	cfg.AuthVersion = AuthV4
	cfg.AccessKeyID = "ak"
	cfg.AccessKeySecret = "sk"
	cfg.Region = "cn-hangzhou"
	um := urlMaker{}
	um.Init(endpoint, false, false)
	conn := Conn{cfg, &um, nil}
	uri := um.getURL("bucket", "1234+-/123/1.txt", "")
	req := &http.Request{
		Method:     "PUT",
		URL:        uri,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     make(http.Header),
		Host:       uri.Host,
	}
	req.Header.Add("x-oss-head1", "value")
	req.Header.Add("abc", "value")
	req.Header.Add("ZAbc", "value")
	req.Header.Add("XYZ", "value")
	req.Header.Add("content-type", "application/octet-stream")
	signTime := time.Unix(1702781677, 0).UTC()
	expiresTime := time.Unix(1702782276, 0).UTC()
	params := map[string]interface{}{}
	params["param1"] = "value1"
	params["+param1"] = "value3"
	params["|param1"] = "value4"
	params["+param2"] = nil
	params["|param2"] = nil
	params["param2"] = nil
	var akIf Credentials
	credProvider := conn.config.CredentialsProvider
	akIf = credProvider.(CredentialsProvider).GetCredentials()
	if akIf.GetSecurityToken() != "" {
		params[HTTPParamOssSecurityToken] = akIf.GetSecurityToken()
	}

	expires := expiresTime.Unix() - signTime.Unix()
	product := conn.config.GetSignProduct()
	region := conn.config.GetSignRegion()
	strDay := signTime.Format(shortTimeFormatV4)
	additionalList, _ := conn.getAdditionalHeaderKeys(req)

	params[HTTPParamSignatureVersion] = signingAlgorithmV4
	params[HTTPParamCredential] = fmt.Sprintf("%s/%s/%s/%s/aliyun_v4_request", akIf.GetAccessKeyID(), strDay, region, product)
	params[HTTPParamDate] = signTime.Format(timeFormatV4)
	params[HTTPParamExpiresV2] = strconv.FormatInt(expires, 10)
	if len(additionalList) > 0 {
		params[HTTPParamAdditionalHeadersV2] = strings.Join(additionalList, ";")
	}
	bucketName := "bucket"
	objectName := "1234+-/123/1.txt"
	subResource := conn.getSubResource(params)
	canonicalizedResource := conn.getResourceV4(bucketName, objectName, subResource)
	authorizationStr := conn.getSignedStrV4(req, canonicalizedResource, akIf.GetAccessKeySecret(), &signTime)
	params[HTTPParamSignatureV2] = authorizationStr
	urlParams := conn.getURLParams(params)
	signUrl := conn.url.getSignURL(bucketName, objectName, urlParams)
	c.Assert(strings.Contains(signUrl, "x-oss-signature-version=OSS4-HMAC-SHA256"), Equals, true)
	c.Assert(strings.Contains(signUrl, "x-oss-date=20231217T025437Z"), Equals, true)
	c.Assert(strings.Contains(signUrl, "x-oss-expires=599"), Equals, true)
	c.Assert(strings.Contains(signUrl, "x-oss-credential="+url.QueryEscape("ak/20231217/cn-hangzhou/oss/aliyun_v4_request")), Equals, true)
	c.Assert(strings.Contains(signUrl, "x-oss-signature=a39966c61718be0d5b14e668088b3fa07601033f6518ac7b523100014269c0fe"), Equals, true)
	c.Assert("http://bucket.oss-cn-hangzhou.aliyuncs.com/1234%2B-%2F123%2F1.txt?%2Bparam1=value3&%2Bparam2&param1=value1&param2&x-oss-credential=ak%2F20231217%2Fcn-hangzhou%2Foss%2Faliyun_v4_request&x-oss-date=20231217T025437Z&x-oss-expires=599&x-oss-signature=a39966c61718be0d5b14e668088b3fa07601033f6518ac7b523100014269c0fe&x-oss-signature-version=OSS4-HMAC-SHA256&%7Cparam1=value4&%7Cparam2", Equals, signUrl)
}

func (s *OssConnSuite) TestAuthV4QueryToken(c *C) {
	endpoint := "http://oss-cn-hangzhou.aliyuncs.com/"
	cfg := getDefaultOssConfig()
	cfg.AuthVersion = AuthV4
	cfg.AccessKeyID = "ak"
	cfg.AccessKeySecret = "sk"
	cfg.SecurityToken = "token"
	cfg.Region = "cn-hangzhou"
	um := urlMaker{}
	um.Init(endpoint, false, false)
	conn := Conn{cfg, &um, nil}
	uri := um.getURL("bucket", "1234+-/123/1.txt", "")
	req := &http.Request{
		Method:     "PUT",
		URL:        uri,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     make(http.Header),
		Host:       uri.Host,
	}
	req.Header.Add("x-oss-head1", "value")
	req.Header.Add("abc", "value")
	req.Header.Add("ZAbc", "value")
	req.Header.Add("XYZ", "value")
	req.Header.Add("content-type", "application/octet-stream")
	signTime := time.Unix(1702785388, 0).UTC()
	expiresTime := time.Unix(1702785987, 0).UTC()
	params := map[string]interface{}{}
	params["param1"] = "value1"
	params["+param1"] = "value3"
	params["|param1"] = "value4"
	params["+param2"] = nil
	params["|param2"] = nil
	params["param2"] = nil
	var akIf Credentials
	credProvider := conn.config.CredentialsProvider
	akIf = credProvider.(CredentialsProvider).GetCredentials()
	if akIf.GetSecurityToken() != "" {
		params[HTTPParamOssSecurityToken] = akIf.GetSecurityToken()
	}

	expires := expiresTime.Unix() - signTime.Unix()
	product := conn.config.GetSignProduct()
	region := conn.config.GetSignRegion()
	strDay := signTime.Format(shortTimeFormatV4)
	additionalList, _ := conn.getAdditionalHeaderKeys(req)

	params[HTTPParamSignatureVersion] = signingAlgorithmV4
	params[HTTPParamCredential] = fmt.Sprintf("%s/%s/%s/%s/aliyun_v4_request", akIf.GetAccessKeyID(), strDay, region, product)
	params[HTTPParamDate] = signTime.Format(timeFormatV4)
	params[HTTPParamExpiresV2] = strconv.FormatInt(expires, 10)
	if len(additionalList) > 0 {
		params[HTTPParamAdditionalHeadersV2] = strings.Join(additionalList, ";")
	}
	bucketName := "bucket"
	objectName := "1234+-/123/1.txt"
	subResource := conn.getSubResource(params)
	canonicalizedResource := conn.getResourceV4(bucketName, objectName, subResource)
	authorizationStr := conn.getSignedStrV4(req, canonicalizedResource, akIf.GetAccessKeySecret(), &signTime)
	params[HTTPParamSignatureV2] = authorizationStr
	urlParams := conn.getURLParams(params)
	signUrl := conn.url.getSignURL(bucketName, objectName, urlParams)
	c.Assert(strings.Contains(signUrl, "x-oss-signature-version=OSS4-HMAC-SHA256"), Equals, true)
	c.Assert(strings.Contains(signUrl, "x-oss-date=20231217T035628Z"), Equals, true)
	c.Assert(strings.Contains(signUrl, "x-oss-expires=599"), Equals, true)
	c.Assert(strings.Contains(signUrl, "x-oss-credential="+url.QueryEscape("ak/20231217/cn-hangzhou/oss/aliyun_v4_request")), Equals, true)
	c.Assert(strings.Contains(signUrl, "x-oss-signature=3817ac9d206cd6dfc90f1c09c00be45005602e55898f26f5ddb06d7892e1f8b5"), Equals, true)
	c.Assert("http://bucket.oss-cn-hangzhou.aliyuncs.com/1234%2B-%2F123%2F1.txt?%2Bparam1=value3&%2Bparam2&param1=value1&param2&x-oss-credential=ak%2F20231217%2Fcn-hangzhou%2Foss%2Faliyun_v4_request&x-oss-date=20231217T035628Z&x-oss-expires=599&x-oss-security-token=token&x-oss-signature=3817ac9d206cd6dfc90f1c09c00be45005602e55898f26f5ddb06d7892e1f8b5&x-oss-signature-version=OSS4-HMAC-SHA256&%7Cparam1=value4&%7Cparam2", Equals, signUrl)
}

func (s *OssConnSuite) TestAuthV4QueryWithAdditionalHeaders(c *C) {
	endpoint := "http://oss-cn-hangzhou.aliyuncs.com/"
	cfg := getDefaultOssConfig()
	cfg.AuthVersion = AuthV4
	cfg.AccessKeyID = "ak"
	cfg.AccessKeySecret = "sk"
	cfg.Region = "cn-hangzhou"
	cfg.AdditionalHeaders = []string{"ZAbc", "abc"}
	um := urlMaker{}
	um.Init(endpoint, false, false)
	conn := Conn{cfg, &um, nil}
	uri := um.getURL("bucket", "1234+-/123/1.txt", "")
	req := &http.Request{
		Method:     "PUT",
		URL:        uri,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     make(http.Header),
		Host:       uri.Host,
	}
	req.Header.Add("x-oss-head1", "value")
	req.Header.Add("abc", "value")
	req.Header.Add("ZAbc", "value")
	req.Header.Add("XYZ", "value")
	req.Header.Add("content-type", "application/octet-stream")
	signTime := time.Unix(1702783809, 0).UTC()
	expiresTime := time.Unix(1702784408, 0).UTC()
	params := map[string]interface{}{}
	params["param1"] = "value1"
	params["+param1"] = "value3"
	params["|param1"] = "value4"
	params["+param2"] = nil
	params["|param2"] = nil
	params["param2"] = nil
	var akIf Credentials
	credProvider := conn.config.CredentialsProvider
	akIf = credProvider.(CredentialsProvider).GetCredentials()
	if akIf.GetSecurityToken() != "" {
		params[HTTPParamOssSecurityToken] = akIf.GetSecurityToken()
	}

	expires := expiresTime.Unix() - signTime.Unix()
	product := conn.config.GetSignProduct()
	region := conn.config.GetSignRegion()
	strDay := signTime.Format(shortTimeFormatV4)
	additionalList, _ := conn.getAdditionalHeaderKeys(req)

	params[HTTPParamSignatureVersion] = signingAlgorithmV4
	params[HTTPParamCredential] = fmt.Sprintf("%s/%s/%s/%s/aliyun_v4_request", akIf.GetAccessKeyID(), strDay, region, product)
	params[HTTPParamDate] = signTime.Format(timeFormatV4)
	params[HTTPParamExpiresV2] = strconv.FormatInt(expires, 10)
	if len(additionalList) > 0 {
		params[HTTPParamAdditionalHeadersV2] = strings.Join(additionalList, ";")
	}
	bucketName := "bucket"
	objectName := "1234+-/123/1.txt"
	subResource := conn.getSubResource(params)
	canonicalizedResource := conn.getResourceV4(bucketName, objectName, subResource)
	authorizationStr := conn.getSignedStrV4(req, canonicalizedResource, akIf.GetAccessKeySecret(), &signTime)
	params[HTTPParamSignatureV2] = authorizationStr
	urlParams := conn.getURLParams(params)
	signUrl := conn.url.getSignURL(bucketName, objectName, urlParams)
	c.Assert(strings.Contains(signUrl, "x-oss-signature-version=OSS4-HMAC-SHA256"), Equals, true)
	c.Assert(strings.Contains(signUrl, "x-oss-date=20231217T033009Z"), Equals, true)
	c.Assert(strings.Contains(signUrl, "x-oss-expires=599"), Equals, true)
	c.Assert(strings.Contains(signUrl, "x-oss-credential="+url.QueryEscape("ak/20231217/cn-hangzhou/oss/aliyun_v4_request")), Equals, true)
	c.Assert(strings.Contains(signUrl, "x-oss-signature=6bd984bfe531afb6db1f7550983a741b103a8c58e5e14f83ea474c2322dfa2b7"), Equals, true)
	c.Assert(strings.Contains(signUrl, "x-oss-additional-headers="+url.QueryEscape("abc;zabc")), Equals, true)
	c.Assert("http://bucket.oss-cn-hangzhou.aliyuncs.com/1234%2B-%2F123%2F1.txt?%2Bparam1=value3&%2Bparam2&param1=value1&param2&x-oss-additional-headers=abc%3Bzabc&x-oss-credential=ak%2F20231217%2Fcn-hangzhou%2Foss%2Faliyun_v4_request&x-oss-date=20231217T033009Z&x-oss-expires=599&x-oss-signature=6bd984bfe531afb6db1f7550983a741b103a8c58e5e14f83ea474c2322dfa2b7&x-oss-signature-version=OSS4-HMAC-SHA256&%7Cparam1=value4&%7Cparam2", Equals, signUrl)
}
