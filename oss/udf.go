package oss

import (
	"bytes"
	"encoding/xml"
	"hash"
	"hash/crc64"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

// UDF implements the operations of udf.
type UDF struct {
	Client Client
}

//
// UDF 取存储空间（UDF）的对象实例。
//
// error 操作无错误时返回nil，非nil为错误信息。
//
func (client Client) UDF() (*UDF, error) {
	return &UDF{
		client,
	}, nil
}

//
// CreateUDF 新建UDF。
//
// UDF名称OSS全局唯一，如果UDF已被其他用户占用，报UdfAlreadyExist错误。
// 若指定UDF ID，则表示将该UDF绑定到一个已存在的UDF上，即别名。
//
// error  操作无错误为nil，非nil为错误信息。
//
func (udf UDF) CreateUDF(udfConfig UDFConfiguration) error {
	udfConfig.UDFDescription = url.QueryEscape(udfConfig.UDFDescription)
	bs, err := xml.Marshal(udfConfig)
	if err != nil {
		return err
	}
	buffer := new(bytes.Buffer)
	buffer.Write(bs)

	contentType := http.DetectContentType(buffer.Bytes())
	headers := map[string]string{}
	headers[HTTPHeaderContentType] = contentType

	params := map[string]interface{}{}
	params["udf"] = nil
	resp, err := udf.doHeader("POST", params, headers, buffer)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return checkRespCode(resp.StatusCode, []int{http.StatusOK, http.StatusCreated})
}

//
// GetUDF 获得一个UDF的相关信息。包括该UDF的ID、描述信息、权限、Owner、创建时间等。
//
// udfName  需要访问的UDF名称。
// GetUDFResult 操作成功的返回值，error为nil时该返回值有效。
//
// error 操作无错误为nil，非nil为错误信息。
//
func (udf UDF) GetUDF(udfName string) (GetUDFResult, error) {
	var out GetUDFResult
	params := map[string]interface{}{}
	params["udf"] = nil
	params["udfName"] = udfName
	resp, err := udf.doHeader("GET", params, nil, nil)
	if err != nil {
		return out, err
	}
	defer resp.Body.Close()

	err = xmlUnmarshal(resp.Body, &out)
	if err != nil {
		return out, err
	}

	err = decodeGetUDFResult(&out)
	return out, err
}

//
// ListUDFs 获取当前用户下的所有UDF。
//
// options 指定ListUDF的筛选行为，可选，包含UDFID选项。
// 如果指定UDFID选项，表示只列举该UDFID对应的UDF信息。
// 常用使用场景的实现，参数示例程序list_udf.go。
// ListUDFsResult 操作成功后的返回值，error为nil时该返回值有效。
//
// error 操作无错误时返回nil，非nil为错误信息。
//
func (udf UDF) ListUDFs(options ...Option) (ListUDFsResult, error) {
	var out ListUDFsResult

	params, err := getRawParams(options)
	if err != nil {
		return out, err
	}

	params["udf"] = nil
	resp, err := udf.doHeader("GET", params, nil, nil)
	if err != nil {
		return out, err
	}
	defer resp.Body.Close()

	err = xmlUnmarshal(resp.Body, &out)
	if err != nil {
		return out, err
	}

	err = decodeListUDFsResult(&out)
	return out, err
}

//
// DeleteUDF 删除UDF。
//
// udfName 需要删除的UDF名称。
//
// error 操作无错误为nil，非nil为错误信息。
//
func (udf UDF) DeleteUDF(udfName string) error {
	params := map[string]interface{}{}
	params["udf"] = nil
	params["udfName"] = udfName
	resp, err := udf.doHeader("DELETE", params, nil, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return checkRespCode(resp.StatusCode, []int{http.StatusNoContent})
}

//
// UploadUDFImage 上传UDF镜像。OSS会根据用户上传的包构建一个镜像。
//
// 用户上传的包必须为tar.gz格式，并包含如下文件：
//      udf.yaml（udf镜像构建脚本），application（用户自己开发的应用程序）。
// 由于构建镜像需要消耗较多时间，所以该接口对用户来说为一个异步接口。用户在调用该接口成功后，
// 镜像会处于building状态，OSS会在后台进行构建操作。
// 初次上传镜像，镜像版本为1。以后每上传一次镜像，对应的镜像版本都会增加1。
// 如果该UDF的镜像为deleting状态（用户删除镜像，且OSS后台删除镜像任务还未执行完成），则再上
// 传镜像时OSS返回400 Bad Request，错误码为：BadUDFImageStatus。
// 上传的镜像允许的最大大小为5GB，如果超过该大小OSS返回400 Bad Request错误，错误码：InvalidArgument。
//
// udfName  上传镜像时指定的UDF名称。
// reader   io.Reader 需要上传的镜像的reader。
// options  上传镜像时可以指定对象的属性，可用选项有UDFImageDesc，表示UDF镜像的描述信息（
// UploadUDFImage接口自动对描述信息进行url编码，描述信息字符内容长度最大为128字节，允许的字符为除去ASCII）。
// 码中十进制值小于32或等于127之外的所有UTF-8字符。
//
// error 操作成功error为nil，非nil为错误信息。
//
func (udf UDF) UploadUDFImage(udfName string, reader io.Reader, options ...Option) error {
	opts := addContentType(options)
	resp, err := udf.DoUploadUDFImage(udfName, reader, opts)
	defer resp.Body.Close()
	return err
}

//
// UploadUDFImageFromFile 上传UDF镜像。
//
// udfName        上传镜像时指定的UDF名称。
// filePath       需要上传的镜像包文件。
// options        上传镜像时可以指定对象的属性，可用选项有UDFImageDesc，详见UploadUDFImage的options。
//
// error 操作成功error为nil，非nil为错误信息。
//
func (udf UDF) UploadUDFImageFromFile(udfName string, filePath string, options ...Option) error {
	fd, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer fd.Close()

	opts := addContentType(options, filePath, "")

	resp, err := udf.DoUploadUDFImage(udfName, fd, opts)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return err
}

//
// DoUploadUDFImage 上传镜像。
//
// udfName  上传镜像时指定的UDF名称。
// reader   io.Reader 需要上传的镜像的reader。
// options  上传镜像时可以指定对象的属性，可用选项有UDFImageDesc，表示UDF镜像的描述信息
//
// Response 上传请求返回值。
// error  操作无错误为nil，非nil为错误信息。
//
func (udf UDF) DoUploadUDFImage(udfName string, reader io.Reader, options []Option) (*Response, error) {
	isOptSet, _, _ := isOptionSet(options, HTTPHeaderContentType)
	if !isOptSet {
		options = addContentType(options, "")
	}

	listener := getProgressListener(options)

	params, err := getRawParams(options)
	if err != nil {
		return nil, err
	}

	params["udfImage"] = nil
	params["udfName"] = udfName
	resp, err := udf.doOption("POST", params, options, reader, listener)
	if err != nil {
		return nil, err
	}

	if udf.getConfig().IsEnableCRC {
		err = checkCRC(resp, "DoUploadUDFImage")
		if err != nil {
			return resp, err
		}
	}

	err = checkRespCode(resp.StatusCode, []int{http.StatusOK})

	return resp, err
}

//
// GetUDFImageInfo 获得一个UDF镜像的相关信息。所有的镜像版本、镜像状态、所属区域、描述信息以及创建时间等。
//
// udfName  需要访问的镜像的UDF名称。
// GetUDFImageInfoResult 操作成功的返回值，error为nil时该返回值有效。
//
// error 操作无错误为nil，非nil为错误信息。
//
func (udf UDF) GetUDFImageInfo(udfName string) (GetUDFImageInfoResult, error) {
	var out GetUDFImageInfoResult
	params := map[string]interface{}{}
	params["udfImage"] = nil
	params["udfName"] = udfName
	resp, err := udf.doHeader("GET", params, nil, nil)
	if err != nil {
		return out, err
	}
	defer resp.Body.Close()

	err = xmlUnmarshal(resp.Body, &out)
	if err != nil {
		return out, err
	}

	err = decodeGetUDFImageInfoResult(&out)
	return out, err
}

//
// DeleteUDFImage 删除UDF已上传的镜像。调用该接口后，该UDF对应的所有版本的镜像都会被删除。
//
// 如果需要删除的UDF不存在，调用该接口时，OSS返回404 Not Found错误。
// 由于删除镜像需要消耗较多时间，该OSS接口是一个异步接口。用户在调用该接口成功后，该UDF对应
// 的所有版本的镜像都会处于deleting状态，OSS会在后台进行删除操作，最终镜像会被彻底删除。
//
// 删除UDF镜像前，该UDF所有版本的镜像必须处于build_success或者build_failed状态，如果还有building
// 状态的镜像，或者处于deleting状态的镜像（重复删除），调用删除镜像接口会失败，OSS返回400，错误码：BadUdfImageStatus。
//
// udfName 需要删除镜像的UDF名称
//
// error 操作无错误为nil，非nil为错误信息。
//
func (udf UDF) DeleteUDFImage(udfName string) error {
	params := map[string]interface{}{}
	params["udfImage"] = nil
	params["udfName"] = udfName
	resp, err := udf.doHeader("DELETE", params, nil, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return checkRespCode(resp.StatusCode, []int{http.StatusAccepted, http.StatusNoContent})
}

//
// CreateUDFApplication 新建UDF应用。
//
// 创建成功后，该UDF应用会处于creating状态，OSS后台会根据用户的配置分配资源、启动实例。
// 当UDF应用启动成功并成功运行后，用户就可以使用该UDF自定义进行数据处理工作。
//
// 由于创建并启动UDF应用需要消耗较多时间，所以该OSS接口为一个异步接口。用户在调用该接口成功后，
// 相应UDF应用处于creating状态，OSS后台任务会创建并启动UDF应用。
// 创建UDF应用时，指定的镜像版本必须已上传并处于build_success状态。
// 创建UDF应用时，指定的实例个数不能超过该UDF的实例上限
//
// udfName  需要创建应用的UDF名称。
// UDFAppConfiguration  创建UDF应用的配置。
//
// error  操作无错误为nil，非nil为错误信息。
//
func (udf UDF) CreateUDFApplication(udfName string, udfAppConfig UDFAppConfiguration) error {
	bs, err := xml.Marshal(udfAppConfig)
	if err != nil {
		return err
	}
	buffer := new(bytes.Buffer)
	buffer.Write(bs)

	contentType := http.DetectContentType(buffer.Bytes())
	headers := map[string]string{}
	headers[HTTPHeaderContentType] = contentType

	params := map[string]interface{}{}
	params["udfApplication"] = nil
	params["udfName"] = udfName
	params["comp"] = "create"
	resp, err := udf.doHeader("POST", params, headers, buffer)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return checkRespCode(resp.StatusCode, []int{http.StatusOK, http.StatusCreated})
}

//
// GetUDFApplicationInfo 获得一个UDF应用的相关信息。包括该UDF的应用的UDF ID、
// 所处数据中心、镜像版本、实例个数、应用状态、应用资源、创建时间等。
//
// udfName  需要查询的UDF应用的名称。
// GetUDFApplicationInfoResult 操作成功的返回值，error为nil时该返回值有效。
//
// error 操作无错误为nil，非nil为错误信息。
//
func (udf UDF) GetUDFApplicationInfo(udfName string) (GetUDFApplicationInfoResult, error) {
	var out GetUDFApplicationInfoResult

	params := map[string]interface{}{}
	params["udfApplication"] = nil
	params["udfName"] = udfName
	resp, err := udf.doHeader("GET", params, nil, nil)
	if err != nil {
		return out, err
	}
	defer resp.Body.Close()

	err = xmlUnmarshal(resp.Body, &out)
	return out, err
}

//
// ListUDFApplications 获取某个数据中心中用户所有的UDF应用信息。
// 包括该各个UDF的应用的UDF ID、镜像版本、实例个数、应用状态、应用资源、创建时间等。
//
// ListUDFApplicationsResult 操作成功后的返回值，error为nil时该返回值有效。
//
// error 操作无错误时返回nil，非nil为错误信息。
//
func (udf UDF) ListUDFApplications() (ListUDFApplicationsResult, error) {
	var out ListUDFApplicationsResult

	params := map[string]interface{}{}
	params["udfApplication"] = nil
	resp, err := udf.doHeader("GET", params, nil, nil)
	if err != nil {
		return out, err
	}
	defer resp.Body.Close()

	err = xmlUnmarshal(resp.Body, &out)
	return out, err
}

//
// DeleteUDFApplication 删除某个数据中心中对应的UDF应用，并释放UDF应用实例对应的计算资源。
//
// 如果需要删除的UDF不存在，调用该接口时，OSS返回404 Not Found错误。
// 由于删除UDF应用需要消耗较多时间，该OSS接口是一个异步接口。用户在调用该接口成功后，
// 相应UDF应用处于deleting状态，OSS后台任务会删除UDF应用，并释放实例对应的计算资源。
//
// 删除的UDF应用为调用该接口时使用的Endpoint对应的数据中心中的应用。
// 只有处于running或者failed或者bad_image状态下的UDF应用才能删除，否则OSS返回409，InvalidUdfAppOriginalStatus。
//
// udfName 需要删除的应用的UDF名称
//
// error 操作无错误为nil，非nil为错误信息。
//
func (udf UDF) DeleteUDFApplication(udfName string) error {
	params := map[string]interface{}{}
	params["udfApplication"] = nil
	params["udfName"] = udfName
	resp, err := udf.doHeader("DELETE", params, nil, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return checkRespCode(resp.StatusCode, []int{http.StatusAccepted, http.StatusNoContent})
}

//
// UpgradeUDFApplication 更新UDF应用的镜像到指定的版本。
//
// 由于升级UDF应用需要消耗较多时间，所以该OSS接口为一个异步接口。用户在调用该接口成功后，
// 相应UDF应用处于upgrading状态，OSS后台任务会完成UDF应用的升级。升级成功后，应用状态会变成running。
// 升级UDF应用时，指定的镜像版本必须已上传并处于build_success状态。
// 只有处于running或者bad_image状态下的UDF应用才能升级，否则OSS返回409，InvalidUdfAppOriginalStatus。
//
// udfName  需要升级应用的UDF名称。
// imageVersion   使用的UDF镜像版本。
//
// error  操作无错误为nil，非nil为错误信息。
//
func (udf UDF) UpgradeUDFApplication(udfName string, imageVersion int64) error {
	uxml := UpgradeUDFApplicationXML{}
	uxml.ImageVersion = imageVersion
	bs, err := xml.Marshal(uxml)
	if err != nil {
		return err
	}
	buffer := new(bytes.Buffer)
	buffer.Write(bs)

	contentType := http.DetectContentType(buffer.Bytes())
	headers := map[string]string{}
	headers[HTTPHeaderContentType] = contentType

	params := map[string]interface{}{}
	params["udfApplication"] = nil
	params["udfName"] = udfName
	params["comp"] = "upgrade"
	resp, err := udf.doHeader("POST", params, headers, buffer)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return checkRespCode(resp.StatusCode, []int{http.StatusOK, http.StatusAccepted})
}

//
// ResizeUDFApplication 扩容UDF应用。目前还不支持应用缩容。
//
// 由于扩容UDF应用需要消耗较多时间，所以该OSS接口为一个异步接口。用户在调用该接口成功后，
// 相应UDF应用处于resizing状态，OSS后台任务会完成UDF应用的扩容工作。扩容成功后，应用状态会变成running。
// 只有处于running状态下的UDF应用才能扩容，否则OSS返回409，InvalidUdfAppOriginalStatus。
// 扩容UDF应用时，指定的实例个数不能超过该UDF的实例上限
//
// udfName  需要扩容的UDF应用名称。
// instanceNum  应用扩容后的实例个数。
//
// error  操作无错误为nil，非nil为错误信息。
//
func (udf UDF) ResizeUDFApplication(udfName string, instanceNum int64) error {
	rxml := ResizeUDFApplicationXML{}
	rxml.InstanceNum = instanceNum
	bs, err := xml.Marshal(rxml)
	if err != nil {
		return err
	}
	buffer := new(bytes.Buffer)
	buffer.Write(bs)

	contentType := http.DetectContentType(buffer.Bytes())
	headers := map[string]string{}
	headers[HTTPHeaderContentType] = contentType

	params := map[string]interface{}{}
	params["udfApplication"] = nil
	params["udfName"] = udfName
	params["comp"] = "resize"
	resp, err := udf.doHeader("POST", params, headers, buffer)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return checkRespCode(resp.StatusCode, []int{http.StatusOK, http.StatusAccepted})
}

//
// GetUDFApplicationLog 获取一个UDF应用的日志信息，日志为应用的标准输出和标准错误输出（stdout和stderr）。
//
// udfName  需要获取日志的UDF名称。
// options  对象的属性限制项，可选值UDFSince、UDFTail。
//
// io.ReadCloser  reader，读取数据后需要close。error为nil时有效。
// error  操作无错误为nil，非nil为错误信息。
//
func (udf UDF) GetUDFApplicationLog(udfName string, options ...Option) (io.ReadCloser, error) {
	result, err := udf.DoGetUDFApplicationLog(udfName, options)
	if err != nil {
		return nil, err
	}
	return result.Response.Body, nil
}

//
// GetUDFApplicationLogToFile 下载UDF应用的日志文件。
//
// udfName  需要获取日志的UDF名称。
// options  对象的属性限制项，可选值UDFSince、UDFTail。
//
// error  操作无错误时返回error为nil，非nil为错误说明。
//
func (udf UDF) GetUDFApplicationLogToFile(udfName, filePath string, options ...Option) error {
	tempFilePath := filePath + TempFileSuffix

	// 读取Object内容
	result, err := udf.DoGetUDFApplicationLog(udfName, options)
	if err != nil {
		return err
	}
	defer result.Response.Body.Close()

	// 如果文件不存在则创建，存在则清空
	fd, err := os.OpenFile(tempFilePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, FilePermMode)
	if err != nil {
		return err
	}

	// 存储数据到文件
	_, err = io.Copy(fd, result.Response.Body)
	fd.Close()
	if err != nil {
		return err
	}

	// 比较CRC值
	hasRange, _, _ := isOptionSet(options, HTTPHeaderRange)
	if udf.getConfig().IsEnableCRC && !hasRange {
		result.Response.ClientCRC = result.ClientCRC.Sum64()
		err = checkCRC(result.Response, "GetUDFApplicationLogToFile")
		if err != nil {
			os.Remove(tempFilePath)
			return err
		}
	}

	return os.Rename(tempFilePath, filePath)
}

//
// DoGetUDFApplicationLog 下载UDF应用的日志文件。
//
// udfName  需要获取日志的UDF名称。
// options  对象的属性限制项，可选值UDFSince、UDFTail。
//
// error  操作无错误时返回error为nil，非nil为错误说明。
//
func (udf UDF) DoGetUDFApplicationLog(udfName string, options []Option) (*GetObjectResult, error) {
	params, err := getRawParams(options)
	if err != nil {
		return nil, err
	}

	params["udfApplicationLog"] = nil
	params["udfName"] = udfName
	resp, err := udf.doOption("GET", params, options, nil, nil)
	if err != nil {
		return nil, err
	}

	result := &GetObjectResult{
		Response: resp,
	}

	// crc
	var crcCalc hash.Hash64
	hasRange, _, _ := isOptionSet(options, HTTPHeaderRange)
	if udf.getConfig().IsEnableCRC && !hasRange {
		crcCalc = crc64.New(crcTable())
		result.ServerCRC = resp.ServerCRC
		result.ClientCRC = crcCalc
	}

	// progress
	listener := getProgressListener(options)

	contentLen, _ := strconv.ParseInt(resp.Headers.Get(HTTPHeaderContentLength), 10, 64)
	resp.Body = ioutil.NopCloser(TeeReader(resp.Body, crcCalc, contentLen, listener, nil))

	return result, nil
}

// Private
func (udf UDF) doHeader(method string, params map[string]interface{},
	headers map[string]string, data io.Reader) (*Response, error) {
	return udf.Client.Conn.Do(method, "", "", params, headers, data, 0, nil)
}

func (udf UDF) doOption(method string, params map[string]interface{}, options []Option,
	data io.Reader, listener ProgressListener) (*Response, error) {
	headers := make(map[string]string)
	err := handleOptions(headers, options)
	if err != nil {
		return nil, err
	}
	return udf.Client.Conn.Do(method, "", "", params, headers, data, 0, listener)
}

func (udf UDF) getConfig() *Config {
	return udf.Client.Config
}
