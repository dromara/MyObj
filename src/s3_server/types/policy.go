package types

import (
	"fmt"
	"strings"
)

// BucketPolicy Bucket策略JSON结构
// 参考AWS S3 Bucket Policy规范
type BucketPolicy struct {
	Version   string      `json:"Version"`   // 策略版本，通常是 "2012-10-17"
	Statement []Statement `json:"Statement"` // 策略语句列表
}

// Statement 策略语句
type Statement struct {
	Sid       string                 `json:"Sid,omitempty"`       // 语句ID（可选）
	Effect    string                 `json:"Effect"`             // Allow 或 Deny
	Principal Principal              `json:"Principal"`           // 主体（被授权者）
	Action    interface{}            `json:"Action"`              // 操作列表（字符串或字符串数组）
	Resource  interface{}            `json:"Resource"`            // 资源ARN（字符串或字符串数组）
	Condition map[string]interface{} `json:"Condition,omitempty"` // 条件（可选）
}

// Principal 主体（被授权者）
type Principal struct {
	AWS           interface{} `json:"AWS,omitempty"`           // AWS账户、IAM用户或角色（字符串或字符串数组）
	CanonicalUser interface{} `json:"CanonicalUser,omitempty"` // 规范用户ID（字符串或字符串数组）
	Federated     interface{} `json:"Federated,omitempty"`     // 联合身份（字符串或字符串数组）
	Service       interface{} `json:"Service,omitempty"`        // AWS服务（字符串或字符串数组）
}

// ConditionKey 条件键常量
const (
	ConditionKeySourceIP        = "aws:SourceIp"         // 源IP地址
	ConditionKeyCurrentTime     = "aws:CurrentTime"      // 当前时间
	ConditionKeyEpochTime       = "aws:EpochTime"        // 时间戳
	ConditionKeyReferer         = "aws:Referer"          // Referer头
	ConditionKeySecureTransport = "aws:SecureTransport"  // 是否使用HTTPS
	ConditionKeyUserAgent       = "aws:UserAgent"        // User-Agent头
	ConditionKeySourceVPC       = "aws:SourceVpc"        // 源VPC
	ConditionKeySourceVpce      = "aws:SourceVpce"        // 源VPC端点
)

// ValidateBucketPolicy 验证Bucket Policy JSON格式
func ValidateBucketPolicy(policy *BucketPolicy) error {
	// 1. 验证Version
	if policy.Version == "" {
		return fmt.Errorf("Version is required")
	}
	if policy.Version != "2012-10-17" && policy.Version != "2008-10-17" {
		return fmt.Errorf("invalid Version: %s (must be '2012-10-17' or '2008-10-17')", policy.Version)
	}

	// 2. 验证Statement
	if len(policy.Statement) == 0 {
		return fmt.Errorf("Statement is required and cannot be empty")
	}

	// 3. 验证每个Statement
	for i, stmt := range policy.Statement {
		// 验证Effect
		if stmt.Effect != "Allow" && stmt.Effect != "Deny" {
			return fmt.Errorf("Statement[%d]: Effect must be 'Allow' or 'Deny'", i)
		}

		// 验证Principal
		if stmt.Principal.AWS == nil && stmt.Principal.CanonicalUser == nil &&
			stmt.Principal.Federated == nil && stmt.Principal.Service == nil {
			return fmt.Errorf("Statement[%d]: Principal is required", i)
		}

		// 验证Action
		if stmt.Action == nil {
			return fmt.Errorf("Statement[%d]: Action is required", i)
		}

		// 验证Resource
		if stmt.Resource == nil {
			return fmt.Errorf("Statement[%d]: Resource is required", i)
		}

		// 验证Resource格式（应该是ARN格式）
		if err := validateResource(stmt.Resource); err != nil {
			return fmt.Errorf("Statement[%d]: Resource validation failed: %w", i, err)
		}
	}

	return nil
}

// validateResource 验证Resource格式
func validateResource(resource interface{}) error {
	switch v := resource.(type) {
	case string:
		if v == "" {
			return fmt.Errorf("Resource cannot be empty string")
		}
		// 简单验证ARN格式（可以更严格）
		if !strings.HasPrefix(v, "arn:aws:s3:::") && !strings.HasPrefix(v, "arn:aws:s3:::*") {
			return fmt.Errorf("Resource must be a valid S3 ARN (arn:aws:s3:::bucket-name/*)")
		}
	case []interface{}:
		if len(v) == 0 {
			return fmt.Errorf("Resource array cannot be empty")
		}
		for i, r := range v {
			if err := validateResource(r); err != nil {
				return fmt.Errorf("Resource[%d]: %w", i, err)
			}
		}
	case []string:
		if len(v) == 0 {
			return fmt.Errorf("Resource array cannot be empty")
		}
		for i, r := range v {
			if err := validateResource(r); err != nil {
				return fmt.Errorf("Resource[%d]: %w", i, err)
			}
		}
	default:
		return fmt.Errorf("Resource must be a string or array of strings")
	}
	return nil
}
