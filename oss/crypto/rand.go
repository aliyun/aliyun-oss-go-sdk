//go:build !go1.20
// +build !go1.20

package osscrypto

import (
	math_rand "math/rand"
	"time"
)

func init() {
	math_rand.Seed(time.Now().UnixNano())
}
