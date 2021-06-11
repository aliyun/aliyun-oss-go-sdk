package osscrypto

import (
	"strings"

	. "gopkg.in/check.v1"
)

func (s *OssCryptoBucketSuite) TestMasterRsaError(c *C) {

	masterRsaCipher, _ := CreateMasterRsa(matDesc, RandLowStr(100), rsaPrivateKey)
	_, err := masterRsaCipher.Encrypt([]byte("123"))
	c.Assert(err, NotNil)

	masterRsaCipher, _ = CreateMasterRsa(matDesc, rsaPublicKey, RandLowStr(100))
	_, err = masterRsaCipher.Decrypt([]byte("123"))
	c.Assert(err, NotNil)

	testPrivateKey := rsaPrivateKey
	[]byte(testPrivateKey)[100] = testPrivateKey[90]
	masterRsaCipher, _ = CreateMasterRsa(matDesc, rsaPublicKey, testPrivateKey)
	_, err = masterRsaCipher.Decrypt([]byte("123"))
	c.Assert(err, NotNil)

	masterRsaCipher, _ = CreateMasterRsa(matDesc, rsaPublicKey, rsaPrivateKey)

	var cipherData CipherData
	err = cipherData.RandomKeyIv(aesKeySize/2, ivSize/4)
	c.Assert(err, NotNil)

	masterRsaCipher, _ = CreateMasterRsa(matDesc, rsaPublicKey, rsaPrivateKey)
	v := masterRsaCipher.(MasterRsaCipher)

	v.PublicKey = strings.Replace(rsaPublicKey, "PUBLIC KEY", "CERTIFICATE", -1)
	_, err = v.Encrypt([]byte("HELLOW"))
	c.Assert(err, NotNil)

	v.PrivateKey = strings.Replace(rsaPrivateKey, "PRIVATE KEY", "CERTIFICATE", -1)
	_, err = v.Decrypt([]byte("HELLOW"))
	c.Assert(err, NotNil)
}
