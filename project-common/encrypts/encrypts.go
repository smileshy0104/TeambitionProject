package encrypts

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/hex"
	"io"
	"strconv"
)

// Md5 计算给定字符串的MD5哈希值。
// 该函数使用MD5算法来生成输入字符串的哈希值，主要用于数据完整性校验。
func Md5(str string) string {
	hash := md5.New()
	_, _ = io.WriteString(hash, str)
	return hex.EncodeToString(hash.Sum(nil))
}

// commonIV 是用于AES加密器的通用初始化向量。
// 初始化向量(IV)是一个固定长度的输入，对于使用相同密钥的每个加密操作，它使得加密输出唯一。
var commonIV = []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f}

// EncryptInt64 加密int64类型的ID。
// 该函数首先将ID转换为字符串，然后调用Encrypt函数进行加密。
func EncryptInt64(id int64, keyText string) (cipherStr string, err error) {
	idStr := strconv.FormatInt(id, 10)
	return Encrypt(idStr, keyText)
}

// Encrypt 使用AES算法加密给定的文本。
// 该函数接受明文和密钥文本作为输入，返回加密后的密文和可能的错误。
// 它使用通用的初始化向量和AES加密算法来确保数据的安全性。
func Encrypt(plainText string, keyText string) (cipherStr string, err error) {
	plainByte := []byte(plainText)
	keyByte := []byte(keyText)
	c, err := aes.NewCipher(keyByte)
	if err != nil {
		return "", err
	}
	cfb := cipher.NewCFBEncrypter(c, commonIV)
	cipherByte := make([]byte, len(plainByte))
	cfb.XORKeyStream(cipherByte, plainByte)
	cipherStr = hex.EncodeToString(cipherByte)
	return
}

// Decrypt 使用AES算法解密给定的密文。
// 该函数接受密文和密钥文本作为输入，返回解密后的明文和可能的错误。
// 它使用通用的初始化向量和AES解密算法来恢复原始数据。
func Decrypt(cipherStr string, keyText string) (plainText string, err error) {
	keyByte := []byte(keyText)
	c, err := aes.NewCipher(keyByte)
	if err != nil {
		return "", err
	}
	cfbdec := cipher.NewCFBDecrypter(c, commonIV)
	cipherByte, _ := hex.DecodeString(cipherStr)
	plainByte := make([]byte, len(cipherByte))
	cfbdec.XORKeyStream(plainByte, cipherByte)
	plainText = string(plainByte)
	return
}
