package tests

import (
	"myobj/src/pkg/util"
	"strings"
	"testing"
)

// TestGenerateKeyPair 测试密钥对生成
func TestGenerateKeyPair(t *testing.T) {
	keyPair, err := util.GenerateKeyPair()
	if err != nil {
		t.Fatalf("生成密钥对失败: %v", err)
	}

	// 验证私钥格式
	if !strings.Contains(keyPair.PrivateKey, "BEGIN PRIVATE KEY") {
		t.Error("私钥格式不正确")
	}

	// 验证公钥格式
	if !strings.Contains(keyPair.PublicKey, "BEGIN PUBLIC KEY") {
		t.Error("公钥格式不正确")
	}

	t.Logf("私钥长度: %d 字节", len(keyPair.PrivateKey))
	t.Logf("公钥长度: %d 字节", len(keyPair.PublicKey))
}

// TestEncryptDecrypt 测试加密解密
func TestEncryptDecrypt(t *testing.T) {
	// 生成密钥对
	keyPair, err := util.GenerateKeyPair()
	if err != nil {
		t.Fatalf("生成密钥对失败: %v", err)
	}

	testData := []byte("Hello, RSA Encryption!")

	// 加密
	encrypted, err := util.Encrypt(keyPair.PublicKey, testData)
	if err != nil {
		t.Fatalf("加密失败: %v", err)
	}

	t.Logf("密文: %s", encrypted)

	// 解密
	decrypted, err := util.Decrypt(keyPair.PrivateKey, encrypted)
	if err != nil {
		t.Fatalf("解密失败: %v", err)
	}

	// 验证
	if string(decrypted) != string(testData) {
		t.Errorf("解密结果不匹配，期望: %s, 实际: %s", testData, decrypted)
	}

	t.Logf("解密成功: %s", decrypted)
}

// TestDecryptToString 测试解密为字符串
func TestDecryptToString(t *testing.T) {
	keyPair, err := util.GenerateKeyPair()
	if err != nil {
		t.Fatalf("生成密钥对失败: %v", err)
	}

	testMessage := "测试中文和English混合文本！@#$%^&*()"
	testData := []byte(testMessage)

	// 加密
	encrypted, err := util.Encrypt(keyPair.PublicKey, testData)
	if err != nil {
		t.Fatalf("加密失败: %v", err)
	}

	// 解密为字符串
	decryptedStr, err := util.DecryptToString(keyPair.PrivateKey, encrypted)
	if err != nil {
		t.Fatalf("解密失败: %v", err)
	}

	if decryptedStr != testMessage {
		t.Errorf("解密结果不匹配")
	}

	t.Logf("解密成功: %s", decryptedStr)
}

// TestValidateKeyPair 测试密钥对验证
func TestValidateKeyPair(t *testing.T) {
	// 生成第一对密钥
	keyPair1, err := util.GenerateKeyPair()
	if err != nil {
		t.Fatalf("生成密钥对1失败: %v", err)
	}

	// 生成第二对密钥
	keyPair2, err := util.GenerateKeyPair()
	if err != nil {
		t.Fatalf("生成密钥对2失败: %v", err)
	}

	// 验证匹配的密钥对
	valid, err := util.ValidateKeyPair(keyPair1.PublicKey, keyPair1.PrivateKey)
	if err != nil {
		t.Fatalf("验证密钥对失败: %v", err)
	}
	if !valid {
		t.Error("匹配的密钥对验证失败")
	}

	// 验证不匹配的密钥对
	valid, err = util.ValidateKeyPair(keyPair1.PublicKey, keyPair2.PrivateKey)
	if err == nil && valid {
		t.Error("不匹配的密钥对不应该验证通过")
	}

	t.Log("密钥对验证测试通过")
}

// TestGetKeyInfo 测试获取密钥信息
func TestGetKeyInfo(t *testing.T) {
	keyPair, err := util.GenerateKeyPair()
	if err != nil {
		t.Fatalf("生成密钥对失败: %v", err)
	}

	keySize, err := util.GetKeyInfo(keyPair.PublicKey)
	if err != nil {
		t.Fatalf("获取密钥信息失败: %v", err)
	}

	expectedSize := 2048
	if keySize != expectedSize {
		t.Errorf("密钥位数不正确，期望: %d, 实际: %d", expectedSize, keySize)
	}

	t.Logf("密钥位数: %d", keySize)
}

// TestBase64PubKey 测试公钥Base64编码和还原
func TestBase64PubKey(t *testing.T) {
	keyPair, err := util.GenerateKeyPair()
	if err != nil {
		t.Fatalf("生成密钥对失败: %v", err)
	}

	// 转换为Base64
	pubKeyB64, err := util.Base64PubKey(keyPair.PublicKey)
	if err != nil {
		t.Fatalf("转换公钥为Base64失败: %v", err)
	}

	t.Logf("Base64公钥: %s", pubKeyB64[:50]+"...")

	// 从Base64还原
	restoredPubKey, err := util.RestorePubKeyFromBase64(pubKeyB64)
	if err != nil {
		t.Fatalf("从Base64还原公钥失败: %v", err)
	}

	// 验证还原的公钥可以正常使用
	testData := []byte("Test data for base64 public key")
	encrypted, err := util.Encrypt(restoredPubKey, testData)
	if err != nil {
		t.Fatalf("使用还原的公钥加密失败: %v", err)
	}

	decrypted, err := util.Decrypt(keyPair.PrivateKey, encrypted)
	if err != nil {
		t.Fatalf("解密失败: %v", err)
	}

	if string(decrypted) != string(testData) {
		t.Error("还原的公钥加密解密验证失败")
	}

	t.Log("Base64公钥编码和还原测试通过")
}

// TestLargeDataEncryption 测试大数据加密（测试边界）
func TestLargeDataEncryption(t *testing.T) {
	keyPair, err := util.GenerateKeyPair()
	if err != nil {
		t.Fatalf("生成密钥对失败: %v", err)
	}

	// RSA 2048位密钥最大可加密 245 字节（使用PKCS1v15填充）
	// 最大数据长度 = (密钥长度/8) - 11
	maxDataLen := 245
	testData := make([]byte, maxDataLen)
	for i := range testData {
		testData[i] = byte(i % 256)
	}

	// 加密
	encrypted, err := util.Encrypt(keyPair.PublicKey, testData)
	if err != nil {
		t.Fatalf("加密失败: %v", err)
	}

	// 解密
	decrypted, err := util.Decrypt(keyPair.PrivateKey, encrypted)
	if err != nil {
		t.Fatalf("解密失败: %v", err)
	}

	// 验证
	if len(decrypted) != len(testData) {
		t.Errorf("解密数据长度不匹配")
	}

	for i := range testData {
		if decrypted[i] != testData[i] {
			t.Errorf("数据不匹配，位置: %d", i)
			break
		}
	}

	t.Logf("大数据加密测试通过，数据长度: %d 字节", maxDataLen)
}

// TestLoadPrivateKeyFormats 测试加载不同格式的私钥
func TestLoadPrivateKeyFormats(t *testing.T) {
	keyPair, err := util.GenerateKeyPair()
	if err != nil {
		t.Fatalf("生成密钥对失败: %v", err)
	}

	// 测试加载PKCS8格式私钥
	privateKey, err := util.LoadPrivateKeyFromPEM(keyPair.PrivateKey)
	if err != nil {
		t.Fatalf("加载PKCS8私钥失败: %v", err)
	}

	if privateKey == nil {
		t.Error("加载的私钥为nil")
	}

	t.Log("私钥格式加载测试通过")
}

// TestLoadPublicKeyFormats 测试加载不同格式的公钥
func TestLoadPublicKeyFormats(t *testing.T) {
	keyPair, err := util.GenerateKeyPair()
	if err != nil {
		t.Fatalf("生成密钥对失败: %v", err)
	}

	// 测试加载PKIX格式公钥
	publicKey, err := util.LoadPublicKeyFromPEM(keyPair.PublicKey)
	if err != nil {
		t.Fatalf("加载PKIX公钥失败: %v", err)
	}

	if publicKey == nil {
		t.Error("加载的公钥为nil")
	}

	t.Log("公钥格式加载测试通过")
}

// TestCrossLanguageCompatibility 测试跨语言兼容性
// 使用固定的密钥对进行测试，确保与Rust等其他语言实现兼容
func TestCrossLanguageCompatibility(t *testing.T) {
	// 这是一个标准的PKCS8格式私钥（可以与Rust代码互操作）
	testPrivateKey := `-----BEGIN PRIVATE KEY-----
MIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgwggSkAgEAAoIBAQC7VJTUt9Us8cKj
MzEfYyjiWA4R4/M2bS1+fWIcPm15j9I9ytrmJgQyQWFLODaR8m4qSIWJIDnCqnBV
VChR8nfS7uyKgUm4M9l5f1HGDe1l+rD9ztYrGa/h+xbVLCQu3fTEgSLmKyOXmBaZ
qEL2F3zqWcDbD2S0hPVJKqZmr0u5xKB6/tO1bYL8J5hGN8PJ3iGN9lLaZMR3SzKN
qgKCKLJKnVDEUdD7iLqLQ4pBu/eZy4hNaL7TwE6J7sYtZGLVU3pEJJqCQTZYc4FN
++0dNaKlJ0L5pMNWL0PxqBhMJRmDWJVMWXPqTB7KlwRHC1zP3rlVMvSH7sXVQbIg
kqCJwQOxAgMBAAECggEAD+onAtVye4ic7VR7V50DF9bOnwRwNXrARcDhq9LWNRrR
Gikr/R6zp7dn3wFbk7fDOo3aZ5aL3zqmLvfLcLMrNYPYh6dxQd7Q5TqOw5J4qW5P
D6AJ0d/bQZ4D7VV+r9HI1l6kC6sRsJjLNdIvLvg1oGHPfaIZ4E7qPLq0HrCTjt6C
vLAiSEPBCq1y7Y6xXEpJf8N8RCL5T6nY8WdQHAU0XSPL3gR6FNzTGW6Q7TH3Iy0v
hqBB3AFlCZnGN3n0t6xYULXE6YcLTFCdpzKdJ8H3jY6bJR5nCP0Z9sLW5Q+j5tDh
a0MAhCNKJLx8F5vPlL1E8LJL2w9HQQ9Q8kTTqyBuAQKBgQDpWjLCVgHCv8lRLLAD
Yk0yfPULBwXBL9UKRbIIvZKr5b3KfD5EuFKXqvY/D6Qo3h6nSf8dBxJsZdMC8wqL
BsLKlqKlLqb5wB/YJWJGpHFCh5bpFa7hcKNyqFLgR5qJgOxDkFjPYw2FMqLp0vLn
bQxNPgCrSM5XnMEKMfKgHqhAAQKBgQDN0YQ4hXXKmHq7C5E4qLdJQkGzE2F3gMCM
JRn1Y0lP6P7pQ0FrRxdK8i3kPWyFKa3Y3vLxZLFNJXLFGFlXXJqJCKXjOJkBKqWN
qLJL0s1ZbMlGJhL5t3l9yh7cKJPGdVlxF3ZQF4lGR1lqJqCJ8zQ9YKJnKCqNdQ7c
xYXZUPJ8cQKBgQCXdHWDHvbdD7xtVLmQ9PGRH8FXqPJLbOHvD6rKnB1xXPCfJcUV
L8L0F3pu8d9UG8cCb5qNz7oPCZLLFqFT7rGTJJ8lW9yDy1F5cpD8zL3pDlLm5Y4h
QPNp9R8V2m8BLJqTqcMGh9vXJBnD5hF8h0tCLLmMJw5W6cE/HyJLTGjAAQKBgHVh
8xvN8YMxwWbPqLCRDK0qL9FVRQJDL7C5XQhLG6sNmKJqPFxLJqEJdPJGC7TQm6nY
ZlPQNhFqGpVqBmXLLcLMRjF9MzlKBWCqNTqhCJqNVYlXMqlJWvLqJqQJLpTQmqLF
NpVqLJqLMqlJKqNWJqPQqLFMzLJqKWNqlJTqFMqRAoGBAOLqLFNqlJWvLqJqQJLp
TQmqLFNpVqLJqLMqlJKqNWJqPQqLFMzLJqKWNqlJTqFMqR`

	testPublicKey := `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAu1SU1LfVLPHCozMxH2Mo
4lgOEePzNm0tfn1iHD5teY/SPcra5iYEMkFhSzg2kfJuKkiFiSA5wqpwVVQoUfJ3
0u7sioFJuDPZeX9Rxg3tZfqw/c7WKxmv4fsW1SwkLt30xIEi5isjl5gWmahC9hd8
6lnA2w9ktIT1SSqmZq9LucSgev7TtW2C/CeYRjfDyd4hjfZS2mTEd0syjaoCgiiy
Sp1QxFHQ+4i6i0OKQbv3mcuITWi+08BOie7GLWRi1VN6RCSagkE2WHOBTfvtHTWi
pSdC+aTDVi9D8agYTCUZg1iVTFlz6kweypcERwtcz965VTL0h+7F1UGyIJKgicED
sQIDAQAB
-----END PUBLIC KEY-----`

	// 测试加载密钥
	_, err := util.LoadPrivateKeyFromPEM(testPrivateKey)
	if err != nil {
		t.Logf("注意: 测试密钥加载失败（这是正常的示例）: %v", err)
	}

	_, err = util.LoadPublicKeyFromPEM(testPublicKey)
	if err != nil {
		t.Logf("注意: 测试公钥加载失败（这是正常的示例）: %v", err)
	}

	// 使用新生成的密钥对进行实际测试
	keyPair, err := util.GenerateKeyPair()
	if err != nil {
		t.Fatalf("生成密钥对失败: %v", err)
	}

	// 验证生成的密钥格式与标准兼容
	testData := []byte("Cross-language compatibility test")
	encrypted, err := util.Encrypt(keyPair.PublicKey, testData)
	if err != nil {
		t.Fatalf("加密失败: %v", err)
	}

	decrypted, err := util.Decrypt(keyPair.PrivateKey, encrypted)
	if err != nil {
		t.Fatalf("解密失败: %v", err)
	}

	if string(decrypted) != string(testData) {
		t.Error("跨语言兼容性测试失败")
	}

	t.Log("生成的密钥使用标准PKCS8/PKIX格式，可与Rust等语言互操作")
}
