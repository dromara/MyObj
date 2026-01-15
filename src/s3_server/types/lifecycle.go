package types

import (
	"encoding/xml"
	"fmt"
	"time"
)

// LifecycleConfiguration 生命周期配置JSON结构
// 参考AWS S3 Lifecycle规范
type LifecycleConfiguration struct {
	Rules []LifecycleRule `json:"Rule"` // XML中是小写rule，JSON中通常也是Rule
}

// LifecycleRule 生命周期规则
type LifecycleRule struct {
	ID                             string                  `json:"ID,omitempty"`                             // 规则ID（可选）
	Status                         string                  `json:"Status"`                                   // Enabled 或 Disabled
	Prefix                         string                  `json:"Prefix,omitempty"`                         // 对象前缀过滤（可选）
	Filter                         *LifecycleFilter        `json:"Filter,omitempty"`                        // 过滤器（可选，与Prefix互斥）
	Expiration                     *LifecycleExpiration    `json:"Expiration,omitempty"`                    // 过期删除规则（可选）
	NoncurrentVersionExpiration     *NoncurrentVersionExpiration `json:"NoncurrentVersionExpiration,omitempty"` // 非当前版本过期（可选）
	AbortIncompleteMultipartUpload *AbortIncompleteMultipartUpload `json:"AbortIncompleteMultipartUpload,omitempty"` // 取消未完成的分片上传（可选）
	Transitions                    []LifecycleTransition    `json:"Transition,omitempty"`                     // 存储类别转换规则（可选）
	NoncurrentVersionTransitions   []NoncurrentVersionTransition `json:"NoncurrentVersionTransition,omitempty"` // 非当前版本转换（可选）
}

// LifecycleFilter 生命周期过滤器
type LifecycleFilter struct {
	XMLName xml.Name          `xml:"Filter"`
	Prefix  string            `xml:"Prefix,omitempty" json:"Prefix,omitempty"` // 前缀过滤
	Tag     *LifecycleTag     `xml:"Tag,omitempty" json:"Tag,omitempty"`      // 标签过滤
	And     *LifecycleFilterAnd `xml:"And,omitempty" json:"And,omitempty"`    // AND条件组合
}

// LifecycleFilterAnd AND条件组合
type LifecycleFilterAnd struct {
	XMLName xml.Name       `xml:"And"`
	Prefix  string         `xml:"Prefix,omitempty" json:"Prefix,omitempty"`
	Tags    []LifecycleTag `xml:"Tag,omitempty" json:"Tag,omitempty"`
}

// LifecycleTag 标签过滤
type LifecycleTag struct {
	XMLName xml.Name `xml:"Tag"`
	Key     string   `xml:"Key" json:"Key"`
	Value   string   `xml:"Value" json:"Value"`
}

// LifecycleExpiration 过期删除规则
type LifecycleExpiration struct {
	XMLName                   xml.Name `xml:"Expiration"`
	Date                      string   `xml:"Date,omitempty" json:"Date,omitempty"`                      // 过期日期（ISO 8601格式）
	Days                      int      `xml:"Days,omitempty" json:"Days,omitempty"`                      // 过期天数（从对象创建时间算起）
	ExpiredObjectDeleteMarker bool     `xml:"ExpiredObjectDeleteMarker,omitempty" json:"ExpiredObjectDeleteMarker,omitempty"` // 是否删除过期对象的DeleteMarker
}

// NoncurrentVersionExpiration 非当前版本过期规则
type NoncurrentVersionExpiration struct {
	XMLName        xml.Name `xml:"NoncurrentVersionExpiration"`
	NoncurrentDays int      `xml:"NoncurrentDays" json:"NoncurrentDays"` // 非当前版本保留天数
}

// AbortIncompleteMultipartUpload 取消未完成的分片上传规则
type AbortIncompleteMultipartUpload struct {
	XMLName             xml.Name `xml:"AbortIncompleteMultipartUpload"`
	DaysAfterInitiation int      `xml:"DaysAfterInitiation" json:"DaysAfterInitiation"` // 初始化后多少天取消
}

// LifecycleTransition 存储类别转换规则
type LifecycleTransition struct {
	XMLName     xml.Name `xml:"Transition"`
	Date        string   `xml:"Date,omitempty" json:"Date,omitempty"`         // 转换日期（ISO 8601格式）
	Days        int      `xml:"Days,omitempty" json:"Days,omitempty"`         // 转换天数（从对象创建时间算起）
	StorageClass string  `xml:"StorageClass" json:"StorageClass"`            // 目标存储类别（如：STANDARD_IA, GLACIER等）
}

// NoncurrentVersionTransition 非当前版本转换规则
type NoncurrentVersionTransition struct {
	XMLName        xml.Name `xml:"NoncurrentVersionTransition"`
	NoncurrentDays int      `xml:"NoncurrentDays" json:"NoncurrentDays"` // 非当前版本保留天数
	StorageClass   string   `xml:"StorageClass" json:"StorageClass"`     // 目标存储类别
}

// ValidateLifecycleConfiguration 验证生命周期配置JSON格式
func ValidateLifecycleConfiguration(config *LifecycleConfiguration) error {
	// 1. 验证Rules
	if len(config.Rules) == 0 {
		return fmt.Errorf("Rule is required and cannot be empty")
	}

	// 2. 验证每个Rule
	for i, rule := range config.Rules {
		// 验证Status
		if rule.Status != "Enabled" && rule.Status != "Disabled" {
			return fmt.Errorf("Rule[%d]: Status must be 'Enabled' or 'Disabled'", i)
		}

		// 验证Prefix和Filter不能同时存在
		if rule.Prefix != "" && rule.Filter != nil {
			return fmt.Errorf("Rule[%d]: Prefix and Filter cannot be specified together", i)
		}

		// 验证至少有一个操作（Expiration、Transition、NoncurrentVersionExpiration等）
		hasAction := rule.Expiration != nil ||
			rule.NoncurrentVersionExpiration != nil ||
			rule.AbortIncompleteMultipartUpload != nil ||
			len(rule.Transitions) > 0 ||
			len(rule.NoncurrentVersionTransitions) > 0

		if !hasAction {
			return fmt.Errorf("Rule[%d]: At least one action (Expiration, Transition, etc.) is required", i)
		}

		// 验证Expiration
		if rule.Expiration != nil {
			if rule.Expiration.Date == "" && rule.Expiration.Days == 0 {
				return fmt.Errorf("Rule[%d]: Expiration must specify either Date or Days", i)
			}
			if rule.Expiration.Date != "" && rule.Expiration.Days > 0 {
				return fmt.Errorf("Rule[%d]: Expiration cannot specify both Date and Days", i)
			}
			if rule.Expiration.Days < 0 {
				return fmt.Errorf("Rule[%d]: Expiration Days must be non-negative", i)
			}
			if rule.Expiration.Date != "" {
				if _, err := time.Parse("2006-01-02T15:04:05Z", rule.Expiration.Date); err != nil {
					return fmt.Errorf("Rule[%d]: Expiration Date must be in ISO 8601 format", i)
				}
			}
		}

		// 验证NoncurrentVersionExpiration
		if rule.NoncurrentVersionExpiration != nil {
			if rule.NoncurrentVersionExpiration.NoncurrentDays < 1 {
				return fmt.Errorf("Rule[%d]: NoncurrentVersionExpiration NoncurrentDays must be at least 1", i)
			}
		}

		// 验证AbortIncompleteMultipartUpload
		if rule.AbortIncompleteMultipartUpload != nil {
			if rule.AbortIncompleteMultipartUpload.DaysAfterInitiation < 1 {
				return fmt.Errorf("Rule[%d]: AbortIncompleteMultipartUpload DaysAfterInitiation must be at least 1", i)
			}
		}

		// 验证Transitions
		for j, transition := range rule.Transitions {
			if transition.StorageClass == "" {
				return fmt.Errorf("Rule[%d].Transition[%d]: StorageClass is required", i, j)
			}
			if transition.Date == "" && transition.Days == 0 {
				return fmt.Errorf("Rule[%d].Transition[%d]: must specify either Date or Days", i, j)
			}
			if transition.Date != "" && transition.Days > 0 {
				return fmt.Errorf("Rule[%d].Transition[%d]: cannot specify both Date and Days", i, j)
			}
			if transition.Days < 0 {
				return fmt.Errorf("Rule[%d].Transition[%d]: Days must be non-negative", i, j)
			}
		}

		// 验证NoncurrentVersionTransitions
		for j, transition := range rule.NoncurrentVersionTransitions {
			if transition.StorageClass == "" {
				return fmt.Errorf("Rule[%d].NoncurrentVersionTransition[%d]: StorageClass is required", i, j)
			}
			if transition.NoncurrentDays < 1 {
				return fmt.Errorf("Rule[%d].NoncurrentVersionTransition[%d]: NoncurrentDays must be at least 1", i, j)
			}
		}
	}

	return nil
}
