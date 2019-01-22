// +build go1.7

package oss

import (
	"fmt"
	"io"
	"math"
	"time"

	"golang.org/x/time/rate"
)

const (
	perTokenBandSize int = 1024
)

// OssLimiter: wrapper rate.Limiter
type OssLimiter struct {
	limiter *rate.Limiter
}

// GetOssLimiter:create OssLimiter
// bandSpeed:KB/s
func GetOssLimiter(bandSpeed int) (ossLimiter *OssLimiter, err error) {
	limiter := rate.NewLimiter(rate.Limit(bandSpeed), bandSpeed)

	// first consume the initial full token,the limiter will behave more accurately
	limiter.AllowN(time.Now(), bandSpeed)

	return &OssLimiter{
		limiter: limiter,
	}, nil
}

// BandLimitReader: for limit band upload
type BandLimitReader struct {
	io.ReadCloser
	reader     io.Reader
	ossLimiter *OssLimiter
}

// Read
func (r *BandLimitReader) Read(p []byte) (n int, err error) {
	n = 0
	err = nil
	start := 0
	burst := r.ossLimiter.limiter.Burst()
	var end int
	var tmpN int
	var tc int
	for start < len(p) {
		if start+burst*perTokenBandSize < len(p) {
			end = start + burst*perTokenBandSize
		} else {
			end = len(p)
		}

		tmpN, err = r.reader.Read(p[start:end])
		if tmpN > 0 {
			n += tmpN
			start = n
		}

		if err != nil {
			return
		}

		tc = int(math.Ceil(float64(tmpN) / float64(perTokenBandSize)))
		now := time.Now()
		re := r.ossLimiter.limiter.ReserveN(now, tc)
		if !re.OK() {
			err = fmt.Errorf("ReserveN error,start:%d,end:%d,burst:%d,perTokenBandSize:%d",
				start, end, burst, perTokenBandSize)
			return
		} else {
			timeDelay := re.Delay()
			time.Sleep(timeDelay)
		}
	}
	return
}

// Close ...
func (r *BandLimitReader) Close() error {
	rc, ok := r.reader.(io.ReadCloser)
	if ok {
		return rc.Close()
	}
	return nil
}
