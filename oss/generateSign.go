package oss

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"hash"
	"io"
	"net/url"
	"strings"
	"time"
)

type Compositions struct {
	Xossmap         map[string]string
	Date            string
	ContentType     string
	ContentMd5      string
	AccessKeySecret string
	Method          string //GET POST PUT DELETE...
}

//GenerateSign 生成签名方法（返回签名url或header方式的签名字符串）.
func GenerateSign(flag int, canonicalizedResource string, compositions *Compositions) (string, error) {
	// 找出“x-oss——”在此请求'header的地址
	temp := make(map[string]string)

	for k, v := range compositions.Xossmap {
		if strings.HasPrefix(strings.ToLower(k), "x-oss-") {
			//检查前缀转换小写key 置入map temp
			temp[strings.ToLower(k)] = v
		}
	}
	//head store
	hs := newHeaderSorter(temp)

	// 升序排序
	hs.Sort()

	// Get the CanonicalizedOSSHeaders
	//所有以 x-oss- 为前缀的HTTP Header被称为CanonicalizedOSSHeaders
	canonicalizedOSSHeaders := ""
	for i := range hs.Keys {
		canonicalizedOSSHeaders += hs.Keys[i] + ":" + hs.Vals[i] + "\n"
	}

	// Give other parameters values
	date := compositions.Date               //"Date"
	contentType := compositions.ContentType //"Content-Type"
	contentMd5 := compositions.ContentMd5   //"Content-MD5"
	time.Now()
	signStr := compositions.Method + "\n" + contentMd5 + "\n" + contentType + "\n" + date + "\n" + canonicalizedOSSHeaders + canonicalizedResource
	h := hmac.New(func() hash.Hash { return sha1.New() }, []byte(compositions.AccessKeySecret))
	io.WriteString(h, signStr)
	signedStr := base64.StdEncoding.EncodeToString(h.Sum(nil))
	var result string
	if flag == 1 { //url sign
		URL, _ := url.Parse(signedStr)
		result = url.QueryEscape(URL.String())
	} else { //header sign
		result = "OSS " + "LTAINwY5Hri5wwQL" + ":" + signedStr
	}
	return result, nil
}
