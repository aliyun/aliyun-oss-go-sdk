package oss

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"errors"
	"hash"
	"hash/crc64"
	"io"
	"io/ioutil"
	"os"
	"strconv"
)

// DownloadFile downloads files with multipart download.
//
// objectKey    the object key.
// filePath    the local file to download from objectKey in OSS.
// partSize    the part size in bytes.
// options    object's constraints, check out GetObject for the reference.
//
// error    it's nil when the call succeeds, otherwise it's an error object.
//
func (bucket Bucket) DownloadFile(objectKey, filePath string, partSize int64, options ...Option) error {
	if partSize < 1 {
		return errors.New("oss: part size smaller than 1")
	}

	cpConf, err := getCpConfig(options, filePath)
	if err != nil {
		return err
	}

	uRange, err := getRangeConfig(options)
	if err != nil {
		return err
	}

	routines := getRoutines(options)

	if cpConf.IsEnable {
		return bucket.downloadFileWithCp(objectKey, filePath, partSize, options, cpConf.FilePath, routines, uRange)
	}

	return bucket.downloadFile(objectKey, filePath, partSize, options, routines, uRange)
}

// getRangeConfig gets the download range from the options.
func getRangeConfig(options []Option) (*unpackedRange, error) {
	rangeOpt, err := findOption(options, HTTPHeaderRange, nil)
	if err != nil || rangeOpt == nil {
		return nil, err
	}
	return parseRange(rangeOpt.(string))
}

// ----- concurrent download without checkpoint  -----

// downloadWorkerArg is download worker's parameters
type downloadWorkerArg struct {
	bucket    *Bucket
	key       string
	filePath  string
	options   []Option
	hook      downloadPartHook
	enableCRC bool
}

// downloadPartHook is hook for test
type downloadPartHook func(part downloadPart) error

var downloadPartHooker downloadPartHook = defaultDownloadPartHook

func defaultDownloadPartHook(part downloadPart) error {
	return nil
}

// defaultDownloadProgressListener defines default ProgressListener, shields the ProgressListener in Options of GetObject. 
type defaultDownloadProgressListener struct {
}

// ProgressChanged no-ops
func (listener *defaultDownloadProgressListener) ProgressChanged(event *ProgressEvent) {
}

// downloadWorker
func downloadWorker(id int, arg downloadWorkerArg, jobs <-chan downloadPart, results chan<- downloadPart, failed chan<- error, die <-chan bool) {
	for part := range jobs {
		if err := arg.hook(part); err != nil {
			failed <- err
			break
		}

		// Resolve options
		r := Range(part.Start, part.End)
		p := Progress(&defaultDownloadProgressListener{})
		opts := make([]Option, len(arg.options)+2)
		// Append orderly, can not be reversed!
		opts = append(opts, arg.options...)
		opts = append(opts, r, p)

		rd, err := arg.bucket.GetObject(arg.key, opts...)
		if err != nil {
			failed <- err
			break
		}
		defer rd.Close()

		var crcCalc hash.Hash64
		if arg.enableCRC {
			crcCalc = crc64.New(crcTable())
			contentLen := part.End - part.Start + 1
			rd = ioutil.NopCloser(TeeReader(rd, crcCalc, contentLen, nil, nil))
		}
		defer rd.Close()

		select {
		case <-die:
			return
		default:
		}

		fd, err := os.OpenFile(arg.filePath, os.O_WRONLY, FilePermMode)
		if err != nil {
			failed <- err
			break
		}

		_, err = fd.Seek(part.Start-part.Offset, os.SEEK_SET)
		if err != nil {
			fd.Close()
			failed <- err
			break
		}

		_, err = io.Copy(fd, rd)
		if err != nil {
			fd.Close()
			failed <- err
			break
		}

		if arg.enableCRC {
			part.CRC64 = crcCalc.Sum64()
		}

		fd.Close()
		results <- part
	}
}

// downloadScheduler
func downloadScheduler(jobs chan downloadPart, parts []downloadPart) {
	for _, part := range parts {
		jobs <- part
	}
	close(jobs)
}

// downloadPart defines download part
type downloadPart struct {
	Index  int    // Part number, starting from 0
	Start  int64  // Start index
	End    int64  // End index
	Offset int64  // Offset
	CRC64  uint64 // CRC check value of part
}

// getDownloadParts gets download parts
func getDownloadParts(bucket *Bucket, objectKey string, partSize int64, uRange *unpackedRange) ([]downloadPart, bool, uint64, error) {
	meta, err := bucket.GetObjectDetailedMeta(objectKey)
	if err != nil {
		return nil, false, 0, err
	}

	parts := []downloadPart{}
	objectSize, err := strconv.ParseInt(meta.Get(HTTPHeaderContentLength), 10, 0)
	if err != nil {
		return nil, false, 0, err
	}

	enableCRC := false
	crcVal := (uint64)(0)
	if bucket.getConfig().IsEnableCRC && meta.Get(HTTPHeaderOssCRC64) != "" {
		if uRange == nil || (!uRange.hasStart && !uRange.hasEnd) {
			enableCRC = true
			crcVal, _ = strconv.ParseUint(meta.Get(HTTPHeaderOssCRC64), 10, 0)
		}
	}

	part := downloadPart{}
	i := 0
	start, end := adjustRange(uRange, objectSize)
	for offset := start; offset < end; offset += partSize {
		part.Index = i
		part.Start = offset
		part.End = GetPartEnd(offset, end, partSize)
		part.Offset = start
		part.CRC64 = 0
		parts = append(parts, part)
		i++
	}
	return parts, enableCRC, crcVal, nil
}

// getObjectBytes gets object bytes length
func getObjectBytes(parts []downloadPart) int64 {
	var ob int64
	for _, part := range parts {
		ob += (part.End - part.Start + 1)
	}
	return ob
}

// combineCRCInParts caculates the total CRC of continuous parts
func combineCRCInParts(dps []downloadPart) uint64 {
	if dps == nil || len(dps) == 0 {
		return 0
	}

	crc := dps[0].CRC64
	for i := 1; i < len(dps); i++ {
		crc = CRC64Combine(crc, dps[i].CRC64, (uint64)(dps[i].End-dps[i].Start+1))
	}

	return crc
}

// downloadFile downloads file concurrently without checkpoint.
func (bucket Bucket) downloadFile(objectKey, filePath string, partSize int64, options []Option, routines int, uRange *unpackedRange) error {
	tempFilePath := filePath + TempFileSuffix
	listener := getProgressListener(options)

	// If the file does not exist, create one. If exists, the download will overwrite it.
	fd, err := os.OpenFile(tempFilePath, os.O_WRONLY|os.O_CREATE, FilePermMode)
	if err != nil {
		return err
	}
	fd.Close()

	// Get the parts of the file
	parts, enableCRC, expectedCRC, err := getDownloadParts(&bucket, objectKey, partSize, uRange)
	if err != nil {
		return err
	}

	jobs := make(chan downloadPart, len(parts))
	results := make(chan downloadPart, len(parts))
	failed := make(chan error)
	die := make(chan bool)

	var completedBytes int64
	totalBytes := getObjectBytes(parts)
	event := newProgressEvent(TransferStartedEvent, 0, totalBytes)
	publishProgress(listener, event)

	// Start the download workers
	arg := downloadWorkerArg{&bucket, objectKey, tempFilePath, options, downloadPartHooker, enableCRC}
	for w := 1; w <= routines; w++ {
		go downloadWorker(w, arg, jobs, results, failed, die)
	}

	// Download parts concurrently
	go downloadScheduler(jobs, parts)

	// Waiting for parts download finished
	completed := 0
	for completed < len(parts) {
		select {
		case part := <-results:
			completed++
			completedBytes += (part.End - part.Start + 1)
			parts[part.Index].CRC64 = part.CRC64
			event = newProgressEvent(TransferDataEvent, completedBytes, totalBytes)
			publishProgress(listener, event)
		case err := <-failed:
			close(die)
			event = newProgressEvent(TransferFailedEvent, completedBytes, totalBytes)
			publishProgress(listener, event)
			return err
		}

		if completed >= len(parts) {
			break
		}
	}

	event = newProgressEvent(TransferCompletedEvent, completedBytes, totalBytes)
	publishProgress(listener, event)

	if enableCRC {
		actualCRC := combineCRCInParts(parts)
		err = checkDownloadCRC(actualCRC, expectedCRC)
		if err != nil {
			return err
		}
	}

	return os.Rename(tempFilePath, filePath)
}

// ----- Concurrent download with chcekpoint  -----

const downloadCpMagic = "92611BED-89E2-46B6-89E5-72F273D4B0A3"

type downloadCheckpoint struct {
	Magic     string         // Magic
	MD5       string         // Checkpoint content MD5
	FilePath  string         // Local file
	Object    string         // Key
	ObjStat   objectStat     // Object status
	Parts     []downloadPart // All download parts
	PartStat  []bool         // Parts' download status
	Start     int64          // Start point of the file
	End       int64          // End point of the file
	enableCRC bool           // Whether has CRC check
	CRC       uint64         // CRC check value
}

type objectStat struct {
	Size         int64  // Object size
	LastModified string // Last modified time
	Etag         string // Etag
}

// isValid flags of checkpoint data is valid. It returns true when the data is valid and the checkpoint is valid and the object is not updated.
func (cp downloadCheckpoint) isValid(bucket *Bucket, objectKey string, uRange *unpackedRange) (bool, error) {
	// Compare the CP's Magic and the MD5
	cpb := cp
	cpb.MD5 = ""
	js, _ := json.Marshal(cpb)
	sum := md5.Sum(js)
	b64 := base64.StdEncoding.EncodeToString(sum[:])

	if cp.Magic != downloadCpMagic || b64 != cp.MD5 {
		return false, nil
	}

	// Ensure the object is not updated.
	meta, err := bucket.GetObjectDetailedMeta(objectKey)
	if err != nil {
		return false, err
	}

	objectSize, err := strconv.ParseInt(meta.Get(HTTPHeaderContentLength), 10, 0)
	if err != nil {
		return false, err
	}

	// Compare the object size, last modified time and ETag
	if cp.ObjStat.Size != objectSize ||
		cp.ObjStat.LastModified != meta.Get(HTTPHeaderLastModified) ||
		cp.ObjStat.Etag != meta.Get(HTTPHeaderEtag) {
		return false, nil
	}

	// Check the download range
	if uRange != nil {
		start, end := adjustRange(uRange, objectSize)
		if start != cp.Start || end != cp.End {
			return false, nil
		}
	}

	return true, nil
}

// load checkpoint from local file
func (cp *downloadCheckpoint) load(filePath string) error {
	contents, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	err = json.Unmarshal(contents, cp)
	return err
}

// dump funciton dumps to file
func (cp *downloadCheckpoint) dump(filePath string) error {
	bcp := *cp

	// Calculate MD5
	bcp.MD5 = ""
	js, err := json.Marshal(bcp)
	if err != nil {
		return err
	}
	sum := md5.Sum(js)
	b64 := base64.StdEncoding.EncodeToString(sum[:])
	bcp.MD5 = b64

	// Serialize
	js, err = json.Marshal(bcp)
	if err != nil {
		return err
	}

	// Dump
	return ioutil.WriteFile(filePath, js, FilePermMode)
}

// todoParts gets unfinished parts
func (cp downloadCheckpoint) todoParts() []downloadPart {
	dps := []downloadPart{}
	for i, ps := range cp.PartStat {
		if !ps {
			dps = append(dps, cp.Parts[i])
		}
	}
	return dps
}

// getCompletedBytes gets completed size
func (cp downloadCheckpoint) getCompletedBytes() int64 {
	var completedBytes int64
	for i, part := range cp.Parts {
		if cp.PartStat[i] {
			completedBytes += (part.End - part.Start + 1)
		}
	}
	return completedBytes
}

// prepare initiates download tasks
func (cp *downloadCheckpoint) prepare(bucket *Bucket, objectKey, filePath string, partSize int64, uRange *unpackedRange) error {
	// CP
	cp.Magic = downloadCpMagic
	cp.FilePath = filePath
	cp.Object = objectKey

	// Object
	meta, err := bucket.GetObjectDetailedMeta(objectKey)
	if err != nil {
		return err
	}

	objectSize, err := strconv.ParseInt(meta.Get(HTTPHeaderContentLength), 10, 0)
	if err != nil {
		return err
	}

	cp.ObjStat.Size = objectSize
	cp.ObjStat.LastModified = meta.Get(HTTPHeaderLastModified)
	cp.ObjStat.Etag = meta.Get(HTTPHeaderEtag)

	// Parts
	cp.Parts, cp.enableCRC, cp.CRC, err = getDownloadParts(bucket, objectKey, partSize, uRange)
	if err != nil {
		return err
	}
	cp.PartStat = make([]bool, len(cp.Parts))
	for i := range cp.PartStat {
		cp.PartStat[i] = false
	}

	return nil
}

func (cp *downloadCheckpoint) complete(cpFilePath, downFilepath string) error {
	os.Remove(cpFilePath)
	return os.Rename(downFilepath, cp.FilePath)
}

// downloadFileWithCp downloads files with checkpoint.
func (bucket Bucket) downloadFileWithCp(objectKey, filePath string, partSize int64, options []Option, cpFilePath string, routines int, uRange *unpackedRange) error {
	tempFilePath := filePath + TempFileSuffix
	listener := getProgressListener(options)

	// Load checkpoint data.
	dcp := downloadCheckpoint{}
	err := dcp.load(cpFilePath)
	if err != nil {
		os.Remove(cpFilePath)
	}

	// Load error or data invalid. Re-initialize the download.
	valid, err := dcp.isValid(&bucket, objectKey, uRange)
	if err != nil || !valid {
		if err = dcp.prepare(&bucket, objectKey, filePath, partSize, uRange); err != nil {
			return err
		}
		os.Remove(cpFilePath)
	}

	// Creates the file if not exists. Otherwise the parts download will overwrite it.
	fd, err := os.OpenFile(tempFilePath, os.O_WRONLY|os.O_CREATE, FilePermMode)
	if err != nil {
		return err
	}
	fd.Close()

	// Unfinished parts
	parts := dcp.todoParts()
	jobs := make(chan downloadPart, len(parts))
	results := make(chan downloadPart, len(parts))
	failed := make(chan error)
	die := make(chan bool)

	completedBytes := dcp.getCompletedBytes()
	event := newProgressEvent(TransferStartedEvent, completedBytes, dcp.ObjStat.Size)
	publishProgress(listener, event)

	// Starts the download workers routine
	arg := downloadWorkerArg{&bucket, objectKey, tempFilePath, options, downloadPartHooker, dcp.enableCRC}
	for w := 1; w <= routines; w++ {
		go downloadWorker(w, arg, jobs, results, failed, die)
	}

	// Concurrently downloads parts
	go downloadScheduler(jobs, parts)

	// Waits for the parts download finished
	completed := 0
	for completed < len(parts) {
		select {
		case part := <-results:
			completed++
			dcp.PartStat[part.Index] = true
			dcp.Parts[part.Index].CRC64 = part.CRC64
			dcp.dump(cpFilePath)
			completedBytes += (part.End - part.Start + 1)
			event = newProgressEvent(TransferDataEvent, completedBytes, dcp.ObjStat.Size)
			publishProgress(listener, event)
		case err := <-failed:
			close(die)
			event = newProgressEvent(TransferFailedEvent, completedBytes, dcp.ObjStat.Size)
			publishProgress(listener, event)
			return err
		}

		if completed >= len(parts) {
			break
		}
	}

	event = newProgressEvent(TransferCompletedEvent, completedBytes, dcp.ObjStat.Size)
	publishProgress(listener, event)

	if dcp.enableCRC {
		actualCRC := combineCRCInParts(dcp.Parts)
		err = checkDownloadCRC(actualCRC, dcp.CRC)
		if err != nil {
			return err
		}
	}

	return dcp.complete(cpFilePath, tempFilePath)
}
