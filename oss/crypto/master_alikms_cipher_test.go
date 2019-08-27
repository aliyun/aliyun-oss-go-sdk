package osscrypto

import (
	crypto_rand "crypto/rand"
	"encoding/base64"
	"io"
	"math/rand"
	"time"

	kms "github.com/aliyun/alibaba-cloud-sdk-go/services/kms"
	. "gopkg.in/check.v1"
)

func (s *OssCryptoBucketSuite) TestKmsClient(c *C) {
	rand.Seed(time.Now().UnixNano())
	kmsClient, err := kms.NewClientWithAccessKey(kmsRegion, kmsAccessID, kmsAccessKey)
	c.Assert(err, IsNil)

	// encrypte
	enReq := kms.CreateEncryptRequest()
	enReq.RpcRequest.Scheme = "https"
	enReq.RpcRequest.Method = "POST"
	enReq.RpcRequest.AcceptFormat = "json"

	enReq.KeyId = kmsID

	buff := make([]byte, 10)
	_, err = io.ReadFull(crypto_rand.Reader, buff)
	c.Assert(err, IsNil)
	enReq.Plaintext = base64.StdEncoding.EncodeToString(buff)

	enResponse, err := kmsClient.Encrypt(enReq)
	c.Assert(err, IsNil)

	// decrypte
	deReq := kms.CreateDecryptRequest()
	deReq.RpcRequest.Scheme = "https"
	deReq.RpcRequest.Method = "POST"
	deReq.RpcRequest.AcceptFormat = "json"
	deReq.CiphertextBlob = enResponse.CiphertextBlob
	deResponse, err := kmsClient.Decrypt(deReq)
	c.Assert(err, IsNil)
	c.Assert(deResponse.Plaintext, Equals, enReq.Plaintext)
}

func (s *OssCryptoBucketSuite) TestMasterAliKmsCipherSuccess(c *C) {

	kmsClient, err := kms.NewClientWithAccessKey(kmsRegion, kmsAccessID, kmsAccessKey)
	c.Assert(err, IsNil)

	masterCipher, _ := CreateMasterAliKms(matDesc, kmsID, kmsClient)

	var cd CipherData
	err = cd.RandomKeyIv(aesKeySize, ivSize)
	c.Assert(err, IsNil)

	cd.WrapAlgorithm = masterCipher.GetWrapAlgorithm()
	cd.CEKAlgorithm = KmsAliCryptoWrap
	cd.MatDesc = masterCipher.GetMatDesc()

	// EncryptedKey
	cd.EncryptedKey, err = masterCipher.Encrypt(cd.Key)

	// EncryptedIV
	cd.EncryptedIV, err = masterCipher.Encrypt(cd.IV)

	cloneData := cd.Clone()

	cloneData.Key, _ = masterCipher.Decrypt(cloneData.EncryptedKey)
	cloneData.IV, _ = masterCipher.Decrypt(cloneData.EncryptedIV)

	c.Assert(string(cd.Key), Equals, string(cloneData.Key))
	c.Assert(string(cd.IV), Equals, string(cloneData.IV))

}

func (s *OssCryptoBucketSuite) TestMasterAliKmsCipherError(c *C) {
	kmsClient, err := kms.NewClientWithAccessKey(kmsRegion, kmsAccessID, kmsAccessKey)
	c.Assert(err, IsNil)

	masterCipher, _ := CreateMasterAliKms(matDesc, kmsID, kmsClient)
	v := masterCipher.(MasterAliKmsCipher)
	v.KmsID = ""
	_, err = v.Encrypt([]byte("hellow"))
	c.Assert(err, NotNil)

	_, err = v.Decrypt([]byte("hellow"))
	c.Assert(err, NotNil)
}
