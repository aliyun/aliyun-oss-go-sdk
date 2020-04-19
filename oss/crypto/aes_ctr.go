package osscrypto

import (
	"crypto/aes"
	"crypto/cipher"
	"io"
)

type aesCtr struct {
	encrypter cipher.Stream
	decrypter cipher.Stream
}

func newAesCtr(cd CipherData) (Cipher, error) {
	block, err := aes.NewCipher(cd.Key)
	if err != nil {
		return nil, err
	}

	encrypter := cipher.NewCTR(block, cd.IV)
	decrypter := cipher.NewCTR(block, cd.IV)
	return &aesCtr{encrypter, decrypter}, nil
}

func (c *aesCtr) Encrypt(src io.Reader) io.Reader {
	reader := &ctrEncryptReader{
		encrypter: c.encrypter,
		src:       src,
	}
	return reader
}

type ctrEncryptReader struct {
	encrypter cipher.Stream
	src       io.Reader
}

func (reader *ctrEncryptReader) Read(data []byte) (int, error) {
	plainText := make([]byte, len(data), len(data))
	n, err := reader.src.Read(plainText)
	if n > 0 {
		plainText = plainText[0:n]
		reader.encrypter.XORKeyStream(data, plainText)
	}
	return n, err
}

func (c *aesCtr) Decrypt(src io.Reader) io.Reader {
	return &ctrDecryptReader{
		decrypter: c.decrypter,
		src:       src,
	}
}

type ctrDecryptReader struct {
	decrypter cipher.Stream
	src       io.Reader
}

func (reader *ctrDecryptReader) Read(data []byte) (int, error) {
	cryptoText := make([]byte, len(data), len(data))
	n, err := reader.src.Read(cryptoText)
	if n > 0 {
		cryptoText = cryptoText[0:n]
		reader.decrypter.XORKeyStream(data, cryptoText)
	}
	return n, err
}
