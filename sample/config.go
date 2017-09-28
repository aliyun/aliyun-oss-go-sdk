package sample

const (
	// sample code's env config. You need to specify them with the actual config if you want to run sample code
	endpoint   string = "<endpoint>"
	accessID   string = "<AccessKeyId>"
	accessKey  string = "<AccessKeySecret>"
	bucketName string = "<my-bucket>"

	// the cname endpoint
	// These information are required to run sample/cname_sample
	endpoint4Cname   string = "<endpoint>"
	accessID4Cname   string = "<AccessKeyId>"
	accessKey4Cname  string = "<AccessKeySecret>"
	bucketName4Cname string = "<my-cname-bucket>"

	// the object name in the sample code
	objectKey string = "my-object"

	// the local files to run sample code.
	localFile     string = "src/sample/BingWallpaper-2015-11-07.jpg"
	htmlLocalFile string = "src/sample/The Go Programming Language.html"
	newPicName    string = "src/sample/NewBingWallpaper-2015-11-07.jpg"
)
