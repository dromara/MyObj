package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
)

// CONFIG 全局配置实例
var CONFIG *MyObjConfig

// MyObjConfig 应用程序主配置结构
type MyObjConfig struct {
	Server   Server   `toml:"server"`   // 服务配置
	Auth     Auth     `toml:"auth"`     // 认证配置
	Log      Log      `toml:"log"`      // 日志配置
	Database Database `toml:"database"` // 数据库配置
	Storage  Storage  `toml:"storage"`  // 存储配置
	File     File     `toml:"file"`     // 文件配置
	Cors     Cors     `toml:"cors"`     // 跨域配置
	Cache    Cache    `toml:"cache"`    // 缓存配置
	WebDAV   WebDAV   `toml:"webdav"`   // WebDAV配置
	S3       S3       `toml:"s3"`       // S3服务配置
}

// Server 服务器配置
type Server struct {
	// Host 监听地址
	Host string `toml:"host"`
	// Port 监听端口
	Port int `toml:"port"`
	// ApiKey 启用ApiKey
	ApiKey bool `toml:"api_key"`
	// SSL 启用SSL
	SSL bool `toml:"ssl"`
	// SSLKey SSL证书文件路径
	SSLKey string `toml:"ssl_key"`
	// SSLCert SSL证书文件路径
	SSLCert string `toml:"ssl_cert"`
	// Swagger 启用Swagger API文档
	Swagger bool `toml:"swagger"`
}

// Auth 认证配置
type Auth struct {
	// Secret 密钥
	Secret string `toml:"secret"`
	// ApiKey 启用ApiKey
	ApiKey bool `toml:"api_key"`
	// JwtExpire JWT过期时间
	JwtExpire int `toml:"jwt_expire"`
	// AdminGroupID 管理员组ID
	AdminGroupID int `toml:"admin_group_id"`
}

// Log 日志配置
type Log struct {
	// Level 日志等级
	Level string `toml:"level"`
	// LogPath 日志文件路径
	LogPath string `toml:"log_path"`
	// MaxSize 日志文件最大大小
	MaxSize int `toml:"max_size"`
	// MaxAge 日志文件最大保存天数
	MaxAge int `toml:"max_age"`
}

type Database struct {
	Type        string `toml:"type"`          // Type 数据库类型
	Host        string `toml:"host"`          // 数据库主机地址
	Port        int    `toml:"port"`          // 数据库端口号
	User        string `toml:"user"`          // 数据库用户名
	Password    string `toml:"password"`      // 数据库密码
	DBName      string `toml:"db_name"`       // 数据库名称
	MaxOpen     int    `toml:"max_open"`      // 最大连接数
	MaxIdle     int    `toml:"max_idle"`      // 最大空闲连接数
	MaxLife     int    `toml:"max_life"`      // 连接存活时间
	MaxIdleLife int    `toml:"max_idle_life"` // 最大空闲连接存活时间
}

type File struct {
	// thumbnail 是否生产缩略图
	Thumbnail bool `toml:"thumbnail"`
	// BigFileThreshold 大文件分片阈值GB
	BigFileThreshold int `toml:"big_file_threshold"`
	// BigChunkSize 大文件分片大小GB
	BigChunkSize int `toml:"big_chunk_size"`
	// DataDir 文件存储目录
	DatDir string `toml:"data_dir"`
	// TempDir 文件临时存储目录
	TempDir string `toml:"temp_dir"`
}

// Cors 跨域配置
type Cors struct {
	// Enable 是否启用跨域
	Enable bool `toml:"enable"`
	// AllowOrigin 允许的源（多个用逗号分隔）
	AllowOrigin string `toml:"allow_origin"`
	// AllowMethods 允许的HTTP方法
	AllowMethods string `toml:"allow_methods"`
	// AllowHeaders 允许的请求头
	AllowHeaders string `toml:"allow_headers"`
	// AllowCredentials 允许发送凭证(cookies)
	AllowCredentials bool `toml:"allow_credentials"`
	// ExposeHeaders 允许的响应头
	ExposeHeaders string `toml:"expose_headers"`
}

// Cache 缓存配置
type Cache struct {
	// Type 缓存类型（redis/local）
	Type     string `toml:"type"`
	Host     string `toml:"host"`      // Redis 主机地址
	Port     int    `toml:"port"`      // Redis 端口号
	Password string `toml:"password"`  // Redis 密码
	DB       int    `toml:"db"`        // Redis 数据库索引
	PoolSize int    `toml:"pool_size"` // Redis 连接池大小
}

// Storage 存储配置
type Storage struct {
	// Driver 存储驱动（local/aliyun/baidu等）
	Driver string `toml:"driver"`
}

// WebDAV WebDAV服务配置
type WebDAV struct {
	// Enable 是否启用WebDAV服务
	Enable bool `toml:"enable"`
	// Host 监听地址
	Host string `toml:"host"`
	// Port 监听端口
	Port int `toml:"port"`
	// Prefix 路径前缀
	Prefix string `toml:"prefix"`
}

// S3 S3服务配置
type S3 struct {
	// OperationTimeout 操作超时时间（秒），默认30秒
	OperationTimeout int `toml:"operation_timeout"`
	// Enable 是否启用S3服务
	Enable bool `toml:"enable"`
	// Region 区域名称
	Region string `toml:"region"`
	// SharePort 是否与主服务共用端口
	SharePort bool `toml:"share_port"`
	// Port 独立端口（如果SharePort=false）
	Port int `toml:"port"`
	// PathPrefix 路径前缀
	PathPrefix string `toml:"path_prefix"`
	// EncryptionMasterKey 加密主密钥（用于服务端加密，支持环境变量 S3_ENCRYPTION_MASTER_KEY）
	EncryptionMasterKey string `toml:"encryption_master_key"`
}

// InitConfig 初始化配置
// 自动搜索并加载 config.toml 文件，然后使用环境变量覆盖
// 搜索顺序:
// 1. 当前工作目录
// 2. 可执行文件所在目录
// 3. 项目根目录（通过向上查找）
// 环境变量命名规则: MYOBJ_<SECTION>_<FIELD> (例如: MYOBJ_SERVER_PORT, MYOBJ_DATABASE_HOST)
func InitConfig() error {
	config := new(MyObjConfig)

	// 尝试加载 TOML 配置文件（可选）
	configPath := findConfigFile()
	if configPath != "" {
		if _, err := toml.DecodeFile(configPath, config); err != nil {
			return fmt.Errorf("配置文件解析失败: %v", err)
		}
	}

	// 应用环境变量覆盖
	applyEnvOverrides(config)

	// 验证必要的配置
	if err := validateConfig(config); err != nil {
		return fmt.Errorf("配置验证失败: %v", err)
	}

	CONFIG = config
	return nil
}

// findConfigFile 查找配置文件路径
// 返回找到的配置文件绝对路径，如果未找到则返回空字符串
func findConfigFile() string {
	configName := "config.toml"

	// 1. 尝试当前工作目录
	if wd, err := os.Getwd(); err == nil {
		path := filepath.Join(wd, configName)
		if fileExists(path) {
			return path
		}
	}

	// 2. 尝试可执行文件所在目录
	if execPath, err := os.Executable(); err == nil {
		execDir := filepath.Dir(execPath)
		path := filepath.Join(execDir, configName)
		if fileExists(path) {
			return path
		}
	}

	// 3. 尝试项目根目录（向上查找）
	if wd, err := os.Getwd(); err == nil {
		path := searchUpwards(wd, configName)
		if path != "" {
			return path
		}
	}

	return ""
}

// searchUpwards 从给定目录向上查找指定文件
func searchUpwards(startDir, filename string) string {
	currentDir := startDir

	// 最多向上查找5层
	for i := 0; i < 5; i++ {
		path := filepath.Join(currentDir, filename)
		if fileExists(path) {
			return path
		}

		// 移动到父目录
		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			// 已到达根目录
			break
		}
		currentDir = parentDir
	}

	return ""
}

// fileExists 检查文件是否存在
func fileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// weakSecrets 弱密钥检测列表
var weakSecrets = map[string]bool{
	"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa": true,
	"00000000000000000000000000000000":  true,
	"11111111111111111111111111111111":  true,
	"passwordpasswordpasswordpassword": true,
	"secretsecretsecretsecretsecretsec": true,
	"abcdefghijklmnopqrstuvwxyzabcdef": true,
	"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa": true,
	"000000000000000000000000000000000": true,
	"12345678901234567890123456789012": true,
	"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa":  true,
	"password":                          true,
	"secret":                            true,
	"12345678":                          true,
	"qwerty":                            true,
	"admin":                             true,
	"default":                           true,
	"test":                              true,
	"your-secret-key-here-change-me":    true,
	"change-me-to-a-real-secret":        true,
}

// validateConfig 验证配置的必要字段
func validateConfig(cfg *MyObjConfig) error {
	// 验证服务器配置
	if cfg.Server.Port <= 0 || cfg.Server.Port > 65535 {
		return fmt.Errorf("无效的端口号: %d", cfg.Server.Port)
	}

	// 验证 Server.Host 是否为合法 IP 或 0.0.0.0
	if cfg.Server.Host != "" {
		if net.ParseIP(cfg.Server.Host) == nil {
			return fmt.Errorf("无效的服务器监听地址: %s，必须是合法 IP 或 0.0.0.0", cfg.Server.Host)
		}
	}

	// 验证数据库配置
	if cfg.Database.Type == "" {
		return fmt.Errorf("数据库类型不能为空")
	}
	if cfg.Database.Type != "mysql" && cfg.Database.Type != "sqlite" {
		return fmt.Errorf("不支持的数据库类型: %s", cfg.Database.Type)
	}

	// MySQL 时校验 Host/Port/User/Password 非空
	if cfg.Database.Type == "mysql" {
		if cfg.Database.Host == "" {
			return fmt.Errorf("MySQL 数据库主机地址不能为空")
		}
		if cfg.Database.Port <= 0 || cfg.Database.Port > 65535 {
			return fmt.Errorf("MySQL 数据库端口号无效: %d", cfg.Database.Port)
		}
		if cfg.Database.User == "" {
			return fmt.Errorf("MySQL 数据库用户名不能为空")
		}
		if cfg.Database.Password == "" {
			return fmt.Errorf("MySQL 数据库密码不能为空")
		}
	}

	// 验证认证配置
	if cfg.Auth.Secret == "" {
		// 自动生成安全随机密钥
		key := make([]byte, 32)
		if _, err := rand.Read(key); err != nil {
			return fmt.Errorf("无法生成随机 JWT 密钥: %v", err)
		}
		cfg.Auth.Secret = hex.EncodeToString(key)
		fmt.Println("[INFO] JWT 密钥已自动生成（本次会话有效）")
	}
	if len(cfg.Auth.Secret) < 32 {
		return fmt.Errorf("JWT密钥长度至少32字符")
	}
	// 弱密钥检测
	if weakSecrets[strings.ToLower(cfg.Auth.Secret)] {
		return fmt.Errorf("JWT密钥为常见弱密钥，请更换为安全的随机密钥")
	}

	// 验证 JWT 过期时间为正数且不超过 720 小时（30天）
	if cfg.Auth.JwtExpire <= 0 {
		return fmt.Errorf("JWT 过期时间必须为正数")
	}
	if cfg.Auth.JwtExpire > 720 {
		return fmt.Errorf("JWT 过期时间不能超过 720 小时（30天）")
	}

	// 验证管理员组ID
	if cfg.Auth.AdminGroupID <= 0 {
		cfg.Auth.AdminGroupID = 1 // 默认管理员组ID为1
	}

	// 验证日志配置
	if cfg.Log.LogPath == "" {
		cfg.Log.LogPath = "./logs/" // 使用默认路径
	}
	// 验证 Log.Level
	if cfg.Log.Level != "" {
		validLogLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
		if !validLogLevels[strings.ToLower(cfg.Log.Level)] {
			return fmt.Errorf("无效的日志级别: %s，必须为 debug/info/warn/error 之一", cfg.Log.Level)
		}
	}

	// 验证 Cache.Type
	if cfg.Cache.Type != "" {
		validCacheTypes := map[string]bool{"redis": true, "local": true}
		if !validCacheTypes[strings.ToLower(cfg.Cache.Type)] {
			return fmt.Errorf("无效的缓存类型: %s，必须为 redis/local 之一", cfg.Cache.Type)
		}
	}

	// 验证 Storage.Driver
	if cfg.Storage.Driver != "" {
		validStorageDrivers := map[string]bool{"local": true}
		if !validStorageDrivers[strings.ToLower(cfg.Storage.Driver)] {
			return fmt.Errorf("无效的存储驱动: %s，目前仅支持 local", cfg.Storage.Driver)
		}
	}

	// S3 启用时 encryption_master_key 不能为空
	if cfg.S3.Enable {
		if cfg.S3.EncryptionMasterKey == "" {
			fmt.Println("[WARN] S3 服务已启用但 encryption_master_key 为空，已自动禁用 S3 服务以保障数据安全")
			cfg.S3.Enable = false
		}
	}

	return nil
}

// GetConfig 获取全局配置实例
// 返回当前的配置对象
func GetConfig() *MyObjConfig {
	return CONFIG
}

// applyEnvOverrides 应用环境变量覆盖配置
// 使用反射自动遍历配置结构，根据 TOML tag 构建环境变量名
// 环境变量命名规则: MYOBJ_<SECTION>_<FIELD> (例如: MYOBJ_SERVER_PORT, MYOBJ_DATABASE_HOST)
func applyEnvOverrides(cfg *MyObjConfig) {
	applyEnvOverridesRecursive(reflect.ValueOf(cfg).Elem(), "MYOBJ", "")
	
	// S3 加密主密钥特殊处理（保持向后兼容，支持 S3_ENCRYPTION_MASTER_KEY）
	if val := getEnv("S3_ENCRYPTION_MASTER_KEY"); val != "" {
		cfg.S3.EncryptionMasterKey = val
	}
}

// applyEnvOverridesRecursive 递归应用环境变量覆盖
// v: 结构体的反射值
// prefix: 环境变量前缀（如 "MYOBJ"）
// sectionPrefix: 当前节的名称（如 "SERVER", "DATABASE"）
func applyEnvOverridesRecursive(v reflect.Value, prefix, sectionPrefix string) {
	t := v.Type()
	
	// 遍历结构体的所有字段
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)
		
		// 跳过未导出的字段
		if !field.CanSet() {
			continue
		}
		
		// 获取 TOML tag
		tomlTag := fieldType.Tag.Get("toml")
		if tomlTag == "" || tomlTag == "-" {
			continue
		}
		
		// 构建环境变量名
		// 将 toml tag 转换为大写，例如: "api_key" -> "API_KEY"
		envKey := strings.ToUpper(tomlTag)
		
		// 如果是嵌套结构体，递归处理
		if field.Kind() == reflect.Struct {
			// 构建新的 section 前缀
			newSectionPrefix := tomlTag
			if sectionPrefix != "" {
				newSectionPrefix = sectionPrefix + "_" + strings.ToUpper(tomlTag)
			} else {
				newSectionPrefix = strings.ToUpper(tomlTag)
			}
			applyEnvOverridesRecursive(field, prefix, newSectionPrefix)
			continue
		}
		
		// 构建完整的环境变量名: MYOBJ_<SECTION>_<FIELD>
		var envVarName string
		if sectionPrefix != "" {
			envVarName = prefix + "_" + sectionPrefix + "_" + envKey
		} else {
			envVarName = prefix + "_" + envKey
		}
		
		// 根据字段类型设置值
		setFieldFromEnv(field, envVarName)
	}
}

// setFieldFromEnv 根据环境变量设置字段值
func setFieldFromEnv(field reflect.Value, envVarName string) {
	envValue := os.Getenv(envVarName)
	if envValue == "" {
		return
	}
	
	switch field.Kind() {
	case reflect.String:
		field.SetString(envValue)
		
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if intVal, err := strconv.Atoi(envValue); err == nil {
			field.SetInt(int64(intVal))
		}
		
	case reflect.Bool:
		if boolVal := parseBool(envValue); boolVal != nil {
			field.SetBool(*boolVal)
		}
		
	default:
		// 其他类型暂不支持
	}
}

// getEnv 获取环境变量（字符串）
func getEnv(key string) string {
	return os.Getenv(key)
}

// parseBool 解析布尔值字符串
// 支持: "true"/"false", "1"/"0", "yes"/"no", "on"/"off"
func parseBool(val string) *bool {
	val = strings.ToLower(strings.TrimSpace(val))
	var result bool
	switch val {
	case "true", "1", "yes", "on":
		result = true
		return &result
	case "false", "0", "no", "off":
		result = false
		return &result
	default:
		return nil
	}
}
