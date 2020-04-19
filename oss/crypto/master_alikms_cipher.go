package osscrypto

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	kms "github.com/aliyun/alibaba-cloud-sdk-go/services/kms"
)

// CreateMasterAliKms Create master key interface implemented by ali kms
// matDesc will be converted to json string
func CreateMasterAliKms(matDesc map[string]string, kmsID string, kmsClient *kms.Client) (MasterCipher, error) {
	var masterCipher MasterAliKmsCipher
	if kmsID == "" || kmsClient == nil {
		return masterCipher, fmt.Errorf("kmsID is empty or kmsClient is nil")
	}

	var jsonDesc string
	if len(matDesc) > 0 {
		b, err := json.Marshal(matDesc)
		if err != nil {
			return masterCipher, err
		}
		jsonDesc = string(b)
	}

	masterCipher.MatDesc = jsonDesc
	masterCipher.KmsID = kmsID
	masterCipher.KmsClient = kmsClient
	return masterCipher, nil
}

// MasterAliKmsCipher ali kms master key interface
type MasterAliKmsCipher struct {
	MatDesc   string
	KmsID     string
	KmsClient *kms.Client
}

// GetWrapAlgorithm get master key wrap algorithm
func (mrc MasterAliKmsCipher) GetWrapAlgorithm() string {
	return KmsAliCryptoWrap
}

// GetMatDesc get master key describe
func (mkms MasterAliKmsCipher) GetMatDesc() string {
	return mkms.MatDesc
}

// Encrypt  encrypt data by ali kms
// Mainly used to encrypt object's symmetric secret key and iv
func (mkms MasterAliKmsCipher) Encrypt(plainData []byte) ([]byte, error) {
	// kms Plaintext must be base64 encoded
	base64Plain := base64.StdEncoding.EncodeToString(plainData)
	request := kms.CreateEncryptRequest()
	request.RpcRequest.Scheme = "https"
	request.RpcRequest.Method = "POST"
	request.RpcRequest.AcceptFormat = "json"

	request.KeyId = mkms.KmsID
	request.Plaintext = base64Plain

	response, err := mkms.KmsClient.Encrypt(request)
	if err != nil {
		return nil, err
	}
	return base64.StdEncoding.DecodeString(response.CiphertextBlob)
}

// Decrypt decrypt data by ali kms
// Mainly used to decrypt object's symmetric secret key and iv
func (mkms MasterAliKmsCipher) Decrypt(cryptoData []byte) ([]byte, error) {
	base64Crypto := base64.StdEncoding.EncodeToString(cryptoData)
	request := kms.CreateDecryptRequest()
	request.RpcRequest.Scheme = "https"
	request.RpcRequest.Method = "POST"
	request.RpcRequest.AcceptFormat = "json"
	request.CiphertextBlob = string(base64Crypto)
	response, err := mkms.KmsClient.Decrypt(request)
	if err != nil {
		return nil, err
	}
	return base64.StdEncoding.DecodeString(response.Plaintext)
}
