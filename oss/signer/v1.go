package signer

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"hash"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"
)

var requiredSignedParametersMap = map[string]struct{}{
	"acl":                          {},
	"bucketInfo":                   {},
	"location":                     {},
	"stat":                         {},
	"delete":                       {},
	"append":                       {},
	"tagging":                      {},
	"objectMeta":                   {},
	"uploads":                      {},
	"uploadId":                     {},
	"partNumber":                   {},
	"security-token":               {},
	"position":                     {},
	"response-content-type":        {},
	"response-content-language":    {},
	"response-expires":             {},
	"response-cache-control":       {},
	"response-content-disposition": {},
	"response-content-encoding":    {},
	"restore":                      {},
	"callback":                     {},
	"callback-var":                 {},
	"versions":                     {},
	"versioning":                   {},
	"versionId":                    {},
	"sequential":                   {},
	"continuation-token":           {},
	"regionList":                   {},
	"cloudboxes":                   {},
}

const (
	authorizationHeader = "Authorization"
	securityTokenHeader = "x-oss-security-token"

	dateHeader        = "Date"
	contentTypeHeader = "Content-Type"
	contentMd5Header  = "Content-MD5"
	ossHeaderPreifx   = "x-oss-"
)

type SignerV1 struct {
}

func isSubResource(list []string, key string) bool {
	for _, k := range list {
		if key == k {
			return true
		}
	}
	return false
}

func (SignerV1) Sign(ctx context.Context, signingCtx *SigningContext) error {
	/*
		SignToString =
			VERB + "\n"
			+ Content-MD5 + "\n"
			+ Content-Type + "\n"
			+ Date + "\n"
			+ CanonicalizedOSSHeaders
			+ CanonicalizedResource
		Signature = base64(hmac-sha1(AccessKeySecret, SignToString))
	*/
	request := signingCtx.Request
	cred := signingCtx.Credentials

	signingCtx.Time = time.Now().UTC()
	request.Header.Set(dateHeader, signingCtx.Time.Format(http.TimeFormat))

	if cred.SessionToken != "" {
		request.Header.Set(securityTokenHeader, cred.SessionToken)
	}

	contentMd5 := request.Header.Get(contentMd5Header)
	contentType := request.Header.Get(contentTypeHeader)
	date := request.Header.Get(dateHeader)

	//CanonicalizedOSSHeaders
	var headers []string
	for k, _ := range request.Header {
		lowerCaseKey := strings.ToLower(k)
		if strings.HasPrefix(lowerCaseKey, ossHeaderPreifx) {
			headers = append(headers, lowerCaseKey)
		}
	}
	sort.Strings(headers)
	headerItems := make([]string, len(headers))
	for i, k := range headers {
		headerValues := make([]string, len(request.Header.Values(k)))
		for i, v := range request.Header.Values(k) {
			headerValues[i] = strings.TrimSpace(v)
		}
		headerItems[i] = k + ":" + strings.Join(headerValues, ",") + "\n"
	}
	canonicalizedOSSHeaders := strings.Join(headerItems, "")

	//CanonicalizedResource
	query := request.URL.Query()
	var params []string
	for k, _ := range query {
		if _, ok := requiredSignedParametersMap[k]; ok {
			params = append(params, k)
		} else if strings.HasPrefix(k, ossHeaderPreifx) {
			params = append(params, k)
		} else if isSubResource(signingCtx.SubResource, k) {
			params = append(params, k)
		}
	}
	sort.Strings(params)
	paramItems := make([]string, len(params))
	for i, k := range params {
		v := query.Get(k)
		if len(v) > 0 {
			paramItems[i] = k + "=" + v
		} else {
			paramItems[i] = k
		}
	}
	subResource := strings.Join(paramItems, "&")
	if subResource != "" {
		subResource = "?" + subResource
	}
	var canonicalizedResource string
	if signingCtx.Bucket == "" {
		canonicalizedResource = fmt.Sprintf("/%s%s", signingCtx.Bucket, subResource)
	} else {
		canonicalizedResource = fmt.Sprintf("/%s/%s%s", signingCtx.Bucket, signingCtx.Key, subResource)
	}

	// string to Sign
	stringToSign :=
		request.Method + "\n" +
			contentMd5 + "\n" +
			contentType + "\n" +
			date + "\n" +
			canonicalizedOSSHeaders +
			canonicalizedResource

	// sign
	h := hmac.New(func() hash.Hash { return sha1.New() }, []byte(cred.AccessKeySecret))
	_, err := io.WriteString(h, stringToSign)

	if err != nil {
		return err
	}

	authorizationStr := "OSS " + cred.AccessKeyID + ":" + base64.StdEncoding.EncodeToString(h.Sum(nil))
	request.Header.Set(authorizationHeader, authorizationStr)

	//save sign info
	signingCtx.StringToSign = stringToSign

	//fmt.Printf("StringToSign=%s", stringToSign)
	return nil
}

func (SignerV1) Presign(ctx context.Context, signingCtx *SigningContext) error {
	return nil
}
