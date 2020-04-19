package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"

	kms "github.com/aliyun/alibaba-cloud-sdk-go/services/kms"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/aliyun/aliyun-oss-go-sdk/oss/crypto"
)

func SampleRsaNormalObject() {
	// create oss client
	client, err := oss.New("<yourEndpoint>", "<yourAccessKeyId>", "<yourAccessKeySecret>")
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(-1)
	}

	// Create a description of the master key. Once created, it cannot be modified. The master key description and the master key are one-to-one correspondence.
	// If all objects use the same master key, the master key description can also be empty, but subsequent replacement of the master key is not supported.
	// Because if the description is empty, it is impossible to determine which master key is used when decrypting object.
	// It is strongly recommended that: configure the master key description(json string) for each master key, and the client should save the correspondence between them.
	// The server does not save their correspondence

	// Map converted by the master key description information (json string)
	materialDesc := make(map[string]string)
	materialDesc["desc"] = "<your master encrypt key material describe information>"

	// Create a master key object based on the master key description
	masterRsaCipher, err := osscrypto.CreateMasterRsa(materialDesc, "<your rsa public key>", "<your rsa private key>")
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(-1)
	}

	// Create an interface for encryption based on the master key object, encrypt using aec ctr mode
	contentProvider := osscrypto.CreateAesCtrCipher(masterRsaCipher)

	// Get a storage space for client encryption, the bucket has to be created
	// Client-side encrypted buckets have similar usages to ordinary buckets.
	cryptoBucket, err := osscrypto.GetCryptoBucket(client, "<yourBucketName>", contentProvider)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(-1)
	}

	// put object ,will be automatically encrypted
	err = cryptoBucket.PutObject("<yourObjectName>", bytes.NewReader([]byte("yourObjectValueByteArrary")))
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(-1)
	}

	// get object ,will be automatically decrypted
	body, err := cryptoBucket.GetObject("<yourObjectName>")
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(-1)
	}
	defer body.Close()

	data, err := ioutil.ReadAll(body)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(-1)
	}
	fmt.Println("data:", string(data))
}

func SampleRsaMultiPartObject() {
	// create oss client
	client, err := oss.New("<yourEndpoint>", "<yourAccessKeyId>", "<yourAccessKeySecret>")
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(-1)
	}

	// Create a description of the master key. Once created, it cannot be modified. The master key description and the master key are one-to-one correspondence.
	// If all objects use the same master key, the master key description can also be empty, but subsequent replacement of the master key is not supported.
	// Because if the description is empty, it is impossible to determine which master key is used when decrypting object.
	// It is strongly recommended that: configure the master key description(json string) for each master key, and the client should save the correspondence between them.
	// The server does not save their correspondence

	// Map converted by the master key description information (json string)
	materialDesc := make(map[string]string)
	materialDesc["desc"] = "<your master encrypt key material describe information>"

	// Create a master key object based on the master key description
	masterRsaCipher, err := osscrypto.CreateMasterRsa(materialDesc, "<your rsa public key>", "<your rsa private key>")
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(-1)
	}

	// Create an interface for encryption based on the master key object, encrypt using aec ctr mode
	contentProvider := osscrypto.CreateAesCtrCipher(masterRsaCipher)

	// Get a storage space for client encryption, the bucket has to be created
	// Client-side encrypted buckets have similar usages to ordinary buckets.
	cryptoBucket, err := osscrypto.GetCryptoBucket(client, "<yourBucketName>", contentProvider)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(-1)
	}

	fileName := "<yourLocalFilePath>"
	fileInfo, err := os.Stat(fileName)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(-1)
	}
	fileSize := fileInfo.Size()

	// Encryption context information
	var cryptoContext osscrypto.PartCryptoContext
	cryptoContext.DataSize = fileSize

	// The expected number of parts, the actual number of parts is subject to subsequent calculations.
	expectPartCount := int64(10)

	//Currently aes ctr encryption block size requires 16 byte alignment
	cryptoContext.PartSize = (fileSize / expectPartCount / 16) * 16

	imur, err := cryptoBucket.InitiateMultipartUpload("<yourObjectName>", &cryptoContext)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(-1)
	}

	chunks, err := oss.SplitFileByPartSize(fileName, cryptoContext.PartSize)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(-1)
	}

	var partsUpload []oss.UploadPart
	for _, chunk := range chunks {
		part, err := cryptoBucket.UploadPartFromFile(imur, fileName, chunk.Offset, chunk.Size, (int)(chunk.Number), cryptoContext)
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(-1)
		}
		partsUpload = append(partsUpload, part)
	}

	// Complete
	_, err = cryptoBucket.CompleteMultipartUpload(imur, partsUpload)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(-1)
	}
}

// Query the master key according to the master key description information.
// If you need to decrypt different master key encryption objects, you need to provide this interface.
type MockRsaManager struct {
}

func (mg *MockRsaManager) GetMasterKey(matDesc map[string]string) ([]string, error) {
	// to do
	keyList := []string{"<yourRsaPublicKey>", "<yourRsaPrivatKey>"}
	return keyList, nil
}

// Decrypt the object encrypted by different master keys
func SampleMultipleMasterRsa() {
	// create oss client
	client, err := oss.New("<yourEndpoint>", "<yourAccessKeyId>", "<yourAccessKeySecret>")
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(-1)
	}

	// Create a description of the master key. Once created, it cannot be modified. The master key description and the master key are one-to-one correspondence.
	// If all objects use the same master key, the master key description can also be empty, but subsequent replacement of the master key is not supported.
	// Because if the description is empty, it is impossible to determine which master key is used when decrypting object.
	// It is strongly recommended that: configure the master key description(json string) for each master key, and the client should save the correspondence between them.
	// The server does not save their correspondence

	// Map converted by the master key description information (json string)
	materialDesc := make(map[string]string)
	materialDesc["desc"] = "<your master encrypt key material describe information>"

	// Create a master key object based on the master key description
	masterRsaCipher, err := osscrypto.CreateMasterRsa(materialDesc, "<your rsa public key>", "<your rsa private key>")
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(-1)
	}

	// Create an interface for encryption based on the master key object, encrypt using aec ctr mode
	contentProvider := osscrypto.CreateAesCtrCipher(masterRsaCipher)

	// If you need to decrypt objects encrypted by different ma keys, you need to provide this interface.
	var mockRsaManager MockRsaManager
	var options []osscrypto.CryptoBucketOption
	options = append(options, osscrypto.SetMasterCipherManager(&mockRsaManager))

	// Get a storage space for client encryption, the bucket has to be created
	// Client-side encrypted buckets have similar usages to ordinary buckets.
	cryptoBucket, err := osscrypto.GetCryptoBucket(client, "<yourBucketName>", contentProvider, options...)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(-1)
	}

	// put object ,will be automatically encrypted
	err = cryptoBucket.PutObject("<yourObjectName>", bytes.NewReader([]byte("yourObjectValueByteArrary")))
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(-1)
	}

	// get object ,will be automatically decrypted
	body, err := cryptoBucket.GetObject("<otherObjectNameEncryptedWithOtherRsa>")
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(-1)
	}
	defer body.Close()

	data, err := ioutil.ReadAll(body)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(-1)
	}
	fmt.Println("data:", string(data))
}

func SampleKmsNormalObject() {
	// create oss client
	client, err := oss.New("<yourEndpoint>", "<yourAccessKeyId>", "<yourAccessKeySecret>")
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(-1)
	}

	// create kms client
	kmsClient, err := kms.NewClientWithAccessKey("<yourKmsRegion>", "<yourKmsAccessKeyId>", "<yourKmsAccessKeySecret>")
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(-1)
	}

	// Create a description of the master key. Once created, it cannot be modified. The master key description and the master key are one-to-one correspondence.
	// If all objects use the same master key, the master key description can also be empty, but subsequent replacement of the master key is not supported.
	// Because if the description is empty, it is impossible to determine which master key is used when decrypting object.
	// It is strongly recommended that: configure the master key description(json string) for each master key, and the client should save the correspondence between them.
	// The server does not save their correspondence

	// Map converted by the master key description information (json string)
	materialDesc := make(map[string]string)
	materialDesc["desc"] = "<your kms encrypt key material describe information>"

	// Create a master key object based on the master key description
	masterkmsCipher, err := osscrypto.CreateMasterAliKms(materialDesc, "<YourKmsId>", kmsClient)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(-1)
	}

	// Create an interface for encryption based on the master key object, encrypt using aec ctr mode
	contentProvider := osscrypto.CreateAesCtrCipher(masterkmsCipher)

	// Get a storage space for client encryption, the bucket has to be created
	// Client-side encrypted buckets have similar usages to ordinary buckets.
	cryptoBucket, err := osscrypto.GetCryptoBucket(client, "<yourBucketName>", contentProvider)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(-1)
	}

	// put object ,will be automatically encrypted
	err = cryptoBucket.PutObject("<yourObjectName>", bytes.NewReader([]byte("yourObjectValueByteArrary")))
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(-1)
	}

	// get object ,will be automatically decrypted
	body, err := cryptoBucket.GetObject("<yourObjectName>")
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(-1)
	}
	defer body.Close()

	data, err := ioutil.ReadAll(body)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(-1)
	}
	fmt.Println("data:", string(data))
}

func main() {
	SampleRsaNormalObject()
	SampleRsaMultiPartObject()
	SampleMultipleMasterRsa()
	SampleKmsNormalObject()
}
