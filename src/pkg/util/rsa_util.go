package util

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
)

// RSAKeyPair RSA密钥对结构
type RSAKeyPair struct {
	PrivateKey string // PEM格式私钥
	PublicKey  string // PEM格式公钥
}

// GenerateKeyPair 生成RSA密钥对（2048位）
func GenerateKeyPair() (*RSAKeyPair, error) {
	// 生成2048位RSA密钥对
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("生成密钥对失败: %w", err)
	}

	// 将私钥转换为PKCS8 PEM格式
	privateKeyPEM, err := encodePrivateKeyToPEM(privateKey)
	if err != nil {
		return nil, fmt.Errorf("编码私钥失败: %w", err)
	}

	// 将公钥转换为PKIX PEM格式
	publicKeyPEM, err := encodePublicKeyToPEM(&privateKey.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("编码公钥失败: %w", err)
	}

	return &RSAKeyPair{
		PrivateKey: privateKeyPEM,
		PublicKey:  publicKeyPEM,
	}, nil
}

// LoadPrivateKeyFromPEM 从PEM字符串加载私钥
func LoadPrivateKeyFromPEM(pemStr string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, fmt.Errorf("无效的PEM格式")
	}

	// 尝试PKCS8格式
	if block.Type == "PRIVATE KEY" {
		privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("解析PKCS8私钥失败: %w", err)
		}
		rsaKey, ok := privateKey.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("密钥不是RSA私钥")
		}
		return rsaKey, nil
	}

	// 尝试PKCS1格式
	if block.Type == "RSA PRIVATE KEY" {
		privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("解析PKCS1私钥失败: %w", err)
		}
		return privateKey, nil
	}

	return nil, fmt.Errorf("不支持的私钥格式: %s", block.Type)
}

// LoadPublicKeyFromPEM 从PEM字符串加载公钥
func LoadPublicKeyFromPEM(pemStr string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, fmt.Errorf("无效的PEM格式")
	}

	// 尝试PKIX格式（标准公钥格式）
	if block.Type == "PUBLIC KEY" {
		publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("解析PKIX公钥失败: %w", err)
		}
		rsaKey, ok := publicKey.(*rsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("密钥不是RSA公钥")
		}
		return rsaKey, nil
	}

	// 尝试PKCS1格式
	if block.Type == "RSA PUBLIC KEY" {
		publicKey, err := x509.ParsePKCS1PublicKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("解析PKCS1公钥失败: %w", err)
		}
		return publicKey, nil
	}

	return nil, fmt.Errorf("不支持的公钥格式: %s", block.Type)
}

// Encrypt 使用公钥加密数据（返回Base64编码）
func Encrypt(publicKeyPEM string, data []byte) (string, error) {
	publicKey, err := LoadPublicKeyFromPEM(publicKeyPEM)
	if err != nil {
		return "", err
	}

	// 使用PKCS1v15加密（与Rust代码保持一致）
	encryptedData, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey, data)
	if err != nil {
		return "", fmt.Errorf("加密失败: %w", err)
	}

	// 转换为Base64编码
	encryptedB64 := base64.StdEncoding.EncodeToString(encryptedData)
	return encryptedB64, nil
}

// Decrypt 使用私钥解密数据（输入Base64编码）
func Decrypt(privateKeyPEM string, encryptedDataB64 string) ([]byte, error) {
	privateKey, err := LoadPrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		return nil, err
	}

	// 解码Base64
	encryptedData, err := base64.StdEncoding.DecodeString(encryptedDataB64)
	if err != nil {
		return nil, fmt.Errorf("Base64解码失败: %w", err)
	}

	// 使用PKCS1v15解密
	decryptedData, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, encryptedData)
	if err != nil {
		return nil, fmt.Errorf("解密失败: %w", err)
	}

	return decryptedData, nil
}

// DecryptToString 使用私钥解密数据并转换为字符串
func DecryptToString(privateKeyPEM string, encryptedDataB64 string) (string, error) {
	decryptedData, err := Decrypt(privateKeyPEM, encryptedDataB64)
	if err != nil {
		return "", err
	}
	return string(decryptedData), nil
}

// ValidateKeyPair 验证密钥对是否匹配
func ValidateKeyPair(publicKeyPEM, privateKeyPEM string) (bool, error) {
	testData := []byte("test message for validation")

	// 使用公钥加密
	encrypted, err := Encrypt(publicKeyPEM, testData)
	if err != nil {
		return false, err
	}

	// 使用私钥解密
	decrypted, err := Decrypt(privateKeyPEM, encrypted)
	if err != nil {
		return false, err
	}

	// 比较解密结果
	if string(decrypted) != string(testData) {
		return false, nil
	}

	return true, nil
}

// GetKeyInfo 获取密钥信息（返回密钥位数）
func GetKeyInfo(publicKeyPEM string) (int, error) {
	publicKey, err := LoadPublicKeyFromPEM(publicKeyPEM)
	if err != nil {
		return 0, err
	}

	// 返回密钥位数（字节数 * 8）
	return publicKey.Size() * 8, nil
}

// Base64PubKey 获取公钥的Base64编码（DER格式）
func Base64PubKey(publicKeyPEM string) (string, error) {
	publicKey, err := LoadPublicKeyFromPEM(publicKeyPEM)
	if err != nil {
		return "", err
	}

	// 转换为DER格式
	publicKeyDER, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return "", fmt.Errorf("转换公钥为DER失败: %w", err)
	}

	// Base64编码
	publicKeyB64 := base64.StdEncoding.EncodeToString(publicKeyDER)
	return publicKeyB64, nil
}

// RestorePubKeyFromBase64 从Base64编码还原公钥PEM
func RestorePubKeyFromBase64(publicKeyB64 string) (string, error) {
	// 解码Base64
	publicKeyDER, err := base64.StdEncoding.DecodeString(publicKeyB64)
	if err != nil {
		return "", fmt.Errorf("Base64解码失败: %w", err)
	}

	// 解析DER格式
	publicKeyInterface, err := x509.ParsePKIXPublicKey(publicKeyDER)
	if err != nil {
		return "", fmt.Errorf("解析DER公钥失败: %w", err)
	}

	publicKey, ok := publicKeyInterface.(*rsa.PublicKey)
	if !ok {
		return "", fmt.Errorf("密钥不是RSA公钥")
	}

	// 转换为PEM格式
	publicKeyPEM, err := encodePublicKeyToPEM(publicKey)
	if err != nil {
		return "", err
	}

	return publicKeyPEM, nil
}

// encodePrivateKeyToPEM 将私钥编码为PKCS8 PEM格式
func encodePrivateKeyToPEM(privateKey *rsa.PrivateKey) (string, error) {
	// 转换为PKCS8格式
	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return "", err
	}

	// 创建PEM块
	privateKeyBlock := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyBytes,
	}

	// 编码为PEM格式
	privateKeyPEM := pem.EncodeToMemory(privateKeyBlock)
	return string(privateKeyPEM), nil
}

// encodePublicKeyToPEM 将公钥编码为PKIX PEM格式
func encodePublicKeyToPEM(publicKey *rsa.PublicKey) (string, error) {
	// 转换为PKIX格式
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return "", err
	}

	// 创建PEM块
	publicKeyBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	}

	// 编码为PEM格式
	publicKeyPEM := pem.EncodeToMemory(publicKeyBlock)
	return string(publicKeyPEM), nil
}
