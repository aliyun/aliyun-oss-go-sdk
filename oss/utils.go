package oss

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"hash/crc64"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// userAgent gets user agent
// It has the SDK version information, OS information and GO version
func userAgent() string {
	sys := getSysInfo()
	return fmt.Sprintf("aliyun-sdk-go/%s (%s/%s/%s;%s)", Version, sys.name,
		sys.release, sys.machine, runtime.Version())
}

type sysInfo struct {
	name    string // OS name such as windows/Linux
	release string // OS version 2.6.32-220.23.2.ali1089.el5.x86_64 etc
	machine string // CPU type amd64/x86_64
}

// getSysInfo gets system info
// gets the OS information and CPU type
func getSysInfo() sysInfo {
	name := runtime.GOOS
	release := "-"
	machine := runtime.GOARCH
	if out, err := exec.Command("uname", "-s").CombinedOutput(); err == nil {
		name = string(bytes.TrimSpace(out))
	}
	if out, err := exec.Command("uname", "-r").CombinedOutput(); err == nil {
		release = string(bytes.TrimSpace(out))
	}
	if out, err := exec.Command("uname", "-m").CombinedOutput(); err == nil {
		machine = string(bytes.TrimSpace(out))
	}
	return sysInfo{name: name, release: release, machine: machine}
}

// GetRangeConfig gets the download range from the options.
func GetRangeConfig(options []Option) (*UnpackedRange, error) {
	rangeOpt, err := FindOption(options, HTTPHeaderRange, nil)
	if err != nil || rangeOpt == nil {
		return nil, err
	}
	return ParseRange(rangeOpt.(string))
}

// UnpackedRange
type UnpackedRange struct {
	HasStart bool  // Flag indicates if the start point is specified
	HasEnd   bool  // Flag indicates if the end point is specified
	Start    int64 // Start point
	End      int64 // End point
}

// InvalidRangeError returns invalid range error
func InvalidRangeError(r string) error {
	return fmt.Errorf("InvalidRange %s", r)
}

func GetRangeString(unpackRange UnpackedRange) string {
	var strRange string
	if unpackRange.HasStart && unpackRange.HasEnd {
		strRange = fmt.Sprintf("%d-%d", unpackRange.Start, unpackRange.End)
	} else if unpackRange.HasStart {
		strRange = fmt.Sprintf("%d-", unpackRange.Start)
	} else if unpackRange.HasEnd {
		strRange = fmt.Sprintf("-%d", unpackRange.End)
	}
	return strRange
}

// ParseRange parse various styles of range such as bytes=M-N
func ParseRange(normalizedRange string) (*UnpackedRange, error) {
	var err error
	HasStart := false
	HasEnd := false
	var Start int64
	var End int64

	// Bytes==M-N or ranges=M-N
	nrSlice := strings.Split(normalizedRange, "=")
	if len(nrSlice) != 2 || nrSlice[0] != "bytes" {
		return nil, InvalidRangeError(normalizedRange)
	}

	// Bytes=M-N,X-Y
	rSlice := strings.Split(nrSlice[1], ",")
	rStr := rSlice[0]

	if strings.HasSuffix(rStr, "-") { // M-
		startStr := rStr[:len(rStr)-1]
		Start, err = strconv.ParseInt(startStr, 10, 64)
		if err != nil {
			return nil, InvalidRangeError(normalizedRange)
		}
		HasStart = true
	} else if strings.HasPrefix(rStr, "-") { // -N
		len := rStr[1:]
		End, err = strconv.ParseInt(len, 10, 64)
		if err != nil {
			return nil, InvalidRangeError(normalizedRange)
		}
		if End == 0 { // -0
			return nil, InvalidRangeError(normalizedRange)
		}
		HasEnd = true
	} else { // M-N
		valSlice := strings.Split(rStr, "-")
		if len(valSlice) != 2 {
			return nil, InvalidRangeError(normalizedRange)
		}
		Start, err = strconv.ParseInt(valSlice[0], 10, 64)
		if err != nil {
			return nil, InvalidRangeError(normalizedRange)
		}
		HasStart = true
		End, err = strconv.ParseInt(valSlice[1], 10, 64)
		if err != nil {
			return nil, InvalidRangeError(normalizedRange)
		}
		HasEnd = true
	}

	return &UnpackedRange{HasStart, HasEnd, Start, End}, nil
}

// AdjustRange returns adjusted range, adjust the range according to the length of the file
func AdjustRange(ur *UnpackedRange, size int64) (Start, End int64) {
	if ur == nil {
		return 0, size
	}

	if ur.HasStart && ur.HasEnd {
		Start = ur.Start
		End = ur.End + 1
		if ur.Start < 0 || ur.Start >= size || ur.End > size || ur.Start > ur.End {
			Start = 0
			End = size
		}
	} else if ur.HasStart {
		Start = ur.Start
		End = size
		if ur.Start < 0 || ur.Start >= size {
			Start = 0
		}
	} else if ur.HasEnd {
		Start = size - ur.End
		End = size
		if ur.End < 0 || ur.End > size {
			Start = 0
			End = size
		}
	}
	return
}

// GetNowSec returns Unix time, the number of seconds elapsed since January 1, 1970 UTC.
// gets the current time in Unix time, in seconds.
func GetNowSec() int64 {
	return time.Now().Unix()
}

// GetNowNanoSec returns t as a Unix time, the number of nanoseconds elapsed
// since January 1, 1970 UTC. The result is undefined if the Unix time
// in nanoseconds cannot be represented by an int64. Note that this
// means the result of calling UnixNano on the zero Time is undefined.
// gets the current time in Unix time, in nanoseconds.
func GetNowNanoSec() int64 {
	return time.Now().UnixNano()
}

// GetNowGMT gets the current time in GMT format.
func GetNowGMT() string {
	return time.Now().UTC().Format(http.TimeFormat)
}

// FileChunk is the file chunk definition
type FileChunk struct {
	Number int   // Chunk number
	Offset int64 // Chunk offset
	Size   int64 // Chunk size.
}

// SplitFileByPartNum splits big file into parts by the num of parts.
// Split the file with specified parts count, returns the split result when error is nil.
func SplitFileByPartNum(fileName string, chunkNum int) ([]FileChunk, error) {
	if chunkNum <= 0 || chunkNum > 10000 {
		return nil, errors.New("chunkNum invalid")
	}

	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}

	if int64(chunkNum) > stat.Size() {
		return nil, errors.New("oss: chunkNum invalid")
	}

	var chunks []FileChunk
	var chunk = FileChunk{}
	var chunkN = (int64)(chunkNum)
	for i := int64(0); i < chunkN; i++ {
		chunk.Number = int(i + 1)
		chunk.Offset = i * (stat.Size() / chunkN)
		if i == chunkN-1 {
			chunk.Size = stat.Size()/chunkN + stat.Size()%chunkN
		} else {
			chunk.Size = stat.Size() / chunkN
		}
		chunks = append(chunks, chunk)
	}

	return chunks, nil
}

// SplitFileByPartSize splits big file into parts by the size of parts.
// Splits the file by the part size. Returns the FileChunk when error is nil.
func SplitFileByPartSize(fileName string, chunkSize int64) ([]FileChunk, error) {
	if chunkSize <= 0 {
		return nil, errors.New("chunkSize invalid")
	}

	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}
	var chunkN = stat.Size() / chunkSize
	if chunkN >= 10000 {
		return nil, errors.New("Too many parts, please increase part size")
	}

	var chunks []FileChunk
	var chunk = FileChunk{}
	for i := int64(0); i < chunkN; i++ {
		chunk.Number = int(i + 1)
		chunk.Offset = i * chunkSize
		chunk.Size = chunkSize
		chunks = append(chunks, chunk)
	}

	if stat.Size()%chunkSize > 0 {
		chunk.Number = len(chunks) + 1
		chunk.Offset = int64(len(chunks)) * chunkSize
		chunk.Size = stat.Size() % chunkSize
		chunks = append(chunks, chunk)
	}

	return chunks, nil
}

// GetPartEnd calculates the end position
func GetPartEnd(begin int64, total int64, per int64) int64 {
	if begin+per > total {
		return total - 1
	}
	return begin + per - 1
}

// CrcTable returns the table constructed from the specified polynomial
var CrcTable = func() *crc64.Table {
	return crc64.MakeTable(crc64.ECMA)
}

func GetReaderLen(reader io.Reader) (int64, error) {
	var contentLength int64
	var err error
	switch v := reader.(type) {
	case *bytes.Buffer:
		contentLength = int64(v.Len())
	case *bytes.Reader:
		contentLength = int64(v.Len())
	case *strings.Reader:
		contentLength = int64(v.Len())
	case *os.File:
		fInfo, fError := v.Stat()
		if fError != nil {
			err = fmt.Errorf("can't get reader content length,%s", fError.Error())
		} else {
			contentLength = fInfo.Size()
		}
	case *io.LimitedReader:
		contentLength = int64(v.N)
	case *LimitedReadCloser:
		contentLength = int64(v.N)
	default:
		err = fmt.Errorf("can't get reader content length,unkown reader type")
	}
	return contentLength, err
}

func LimitReadCloser(r io.Reader, n int64) io.Reader {
	var lc LimitedReadCloser
	lc.R = r
	lc.N = n
	return &lc
}

// LimitedRC support Close()
type LimitedReadCloser struct {
	io.LimitedReader
}

func (lc *LimitedReadCloser) Close() error {
	if closer, ok := lc.R.(io.ReadCloser); ok {
		return closer.Close()
	}
	return nil
}

type DiscardReadCloser struct {
	RC      io.ReadCloser
	Discard int
}

func (drc *DiscardReadCloser) Read(b []byte) (int, error) {
	n, err := drc.RC.Read(b)
	if drc.Discard == 0 || n <= 0 {
		return n, err
	}

	if n <= drc.Discard {
		drc.Discard -= n
		return 0, err
	}

	realLen := n - drc.Discard
	copy(b[0:realLen], b[drc.Discard:n])
	drc.Discard = 0
	return realLen, err
}

func (drc *DiscardReadCloser) Close() error {
	closer, ok := drc.RC.(io.ReadCloser)
	if ok {
		return closer.Close()
	}
	return nil
}

func XmlUnmarshal(body io.Reader, v interface{}) error {
	data, err := ioutil.ReadAll(body)
	if err != nil {
		return err
	}
	return xml.Unmarshal(data, v)
}

// decodeListUploadedPartsResult decodes
func DecodeListUploadedPartsResult(result *ListUploadedPartsResult) error {
	var err error
	result.Key, err = url.QueryUnescape(result.Key)
	if err != nil {
		return err
	}
	return nil
}
