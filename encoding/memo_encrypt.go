package encoding

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha512"

	"github.com/blocktree/go-owcrypt"
)

// SetMemoMessage encrypte the memo
func SetMemoMessage(priv, pub []byte, message string) []byte {
	toPub := owcrypt.PointDecompress(pub, owcrypt.ECC_CURVE_SECP256K1)
	secret := getSharedSecret(priv, toPub)
	encrypted := AesEncryptCBC([]byte(message), secret)

	return encrypted
}

func getSharedSecret(priv, pub []byte) []byte {
	sharedSecret, _ := owcrypt.Point_mul(pub[1:], priv, owcrypt.ECC_CURVE_SECP256K1)

	h := sha512.New()
	h.Write(sharedSecret[:32])
	return h.Sum(nil)
}

func AesEncryptCBC(origData []byte, secret []byte) (encrypted []byte) {
	// 分组秘钥
	// NewCipher该函数限制了输入k的长度必须为16, 24或者32
	key := secret[:32]
	block, _ := aes.NewCipher(key)
	blockSize := block.BlockSize() // 获取秘钥块的长度
	// iv := secret[32:48]
	iv := secret[32 : 32+blockSize]

	origData = pkcs5Padding(origData, blockSize)   // 补全码
	blockMode := cipher.NewCBCEncrypter(block, iv) // 加密模式
	encrypted = make([]byte, len(origData))        // 创建数组
	blockMode.CryptBlocks(encrypted, origData)     // 加密
	return encrypted
}

func pkcs5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func AesDecryptCBC(encrypted []byte, secret []byte) (decrypted []byte) {
	key := secret[:32]
	block, _ := aes.NewCipher(key) // 分组秘钥
	blockSize := block.BlockSize() // 获取秘钥块的长度
	// iv := secret[32:48]
	iv := secret[32 : 32+blockSize]
	blockMode := cipher.NewCBCDecrypter(block, iv) // 加密模式
	decrypted = make([]byte, len(encrypted))       // 创建数组
	blockMode.CryptBlocks(decrypted, encrypted)    // 解密
	decrypted = pkcs5UnPadding(decrypted)          // 去除补全码
	return decrypted
}

func pkcs5UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}
