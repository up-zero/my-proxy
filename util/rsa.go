package util

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"sync"
)

// rsaKeyBits RSA 密钥长度
const rsaKeyBits = 2048

var (
	rsaKeyOnce    sync.Once
	rsaPrivateKey *rsa.PrivateKey
	rsaKeyErr     error
)

// getRSAPrivateKey 获取 RSA 私钥（首次调用时生成，仅存于内存，重启后重新生成）
func getRSAPrivateKey() (*rsa.PrivateKey, error) {
	rsaKeyOnce.Do(func() {
		rsaPrivateKey, rsaKeyErr = rsa.GenerateKey(rand.Reader, rsaKeyBits)
	})
	return rsaPrivateKey, rsaKeyErr
}

// RSAPublicKeyPEM 获取 PEM 格式的 RSA 公钥（供前端加密登录密码使用）
func RSAPublicKeyPEM() (string, error) {
	privateKey, err := getRSAPrivateKey()
	if err != nil {
		return "", err
	}
	der, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return "", err
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der})), nil
}

// RSADecrypt 解密前端使用 RSA 公钥加密的密文（base64 编码，PKCS#1 v1.5 填充）
func RSADecrypt(cipherText string) (string, error) {
	privateKey, err := getRSAPrivateKey()
	if err != nil {
		return "", err
	}
	cipherData, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil {
		return "", err
	}
	plainData, err := rsa.DecryptPKCS1v15(nil, privateKey, cipherData)
	if err != nil {
		return "", err
	}
	return string(plainData), nil
}
