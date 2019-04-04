package sample

import "os"

var (
	// Sample code's env configuration. You need to specify them with the actual configuration if you want to run sample code
	endpoint   = os.Getenv("OSS_TEST_ENDPOINT")
	accessID   = os.Getenv("OSS_TEST_ACCESS_KEY_ID")
	accessKey  = os.Getenv("OSS_TEST_ACCESS_KEY_SECRET")
	bucketName = os.Getenv("OSS_TEST_BUCKET")
	kmsID      = os.Getenv("OSS_TEST_KMS_ID")

	// The cname endpoint
	// These information are required to run sample/cname_sample
	endpoint4Cname   = os.Getenv("OSS_TEST_ENDPOINT")
	accessID4Cname   = os.Getenv("OSS_TEST_ACCESS_KEY_ID")
	accessKey4Cname  = os.Getenv("OSS_TEST_ACCESS_KEY_SECRET")
	bucketName4Cname = os.Getenv("OSS_TEST_CNAME_BUCKET")
)

const (

	// The object name in the sample code
	objectKey       string = "my-object"
	objectAppendKey string = "my-object-append"

	// The local files to run sample code.
	localFile     string = "src/sample/BingWallpaper-2015-11-07.jpg"
	htmlLocalFile string = "src/sample/The Go Programming Language.html"
	newPicName    string = "src/sample/NewBingWallpaper-2015-11-07.jpg"
)
