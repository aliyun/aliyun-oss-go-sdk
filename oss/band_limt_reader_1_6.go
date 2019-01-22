// +build !go1.7

// "golang.org/x/time/rate" is depended on golang context package  go1.7 onward
package oss

import (
	"fmt"
)

const (
	perTokenBandSize int = 1024
)

type OssLimiter struct {
}

type BandLimitReader struct {
}

func GetOssLimiter(bandSpeed int) (ossLimiter *OssLimiter, err error) {
	err = fmt.Errorf("rate.Limiter is not supported below version go1.7")
	return nil, err
}
