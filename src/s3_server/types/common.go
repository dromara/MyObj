package types

import (
	"encoding/xml"
	"time"
)

// Owner S3所有者信息
type Owner struct {
	ID          string `xml:"ID"`
	DisplayName string `xml:"DisplayName"`
}

// Bucket S3存储桶信息
type Bucket struct {
	Name         string `xml:"Name"`
	CreationDate string `xml:"CreationDate"`
}

// Buckets 存储桶列表
type Buckets struct {
	Bucket []Bucket `xml:"Bucket"`
}

// ListAllMyBucketsResult ListBuckets响应
type ListAllMyBucketsResult struct {
	XMLName string  `xml:"ListAllMyBucketsResult"`
	Owner   Owner   `xml:"Owner"`
	Buckets Buckets `xml:"Buckets"`
}

// Contents 对象信息
type Contents struct {
	Key          string `xml:"Key"`
	LastModified string `xml:"LastModified"`
	ETag         string `xml:"ETag"`
	Size         int64  `xml:"Size"`
	StorageClass string `xml:"StorageClass"`
	Owner        Owner  `xml:"Owner,omitempty"`
}

// CommonPrefix 公共前缀
type CommonPrefix struct {
	Prefix string `xml:"Prefix"`
}

// ListBucketResult ListObjects响应
type ListBucketResult struct {
	XMLName        string         `xml:"ListBucketResult"`
	Name           string         `xml:"Name"`
	Prefix         string         `xml:"Prefix,omitempty"`
	Marker         string         `xml:"Marker,omitempty"`
	NextMarker     string         `xml:"NextMarker,omitempty"`
	MaxKeys        int            `xml:"MaxKeys"`
	Delimiter      string         `xml:"Delimiter,omitempty"`
	IsTruncated    bool           `xml:"IsTruncated"`
	Contents       []Contents     `xml:"Contents,omitempty"`
	CommonPrefixes []CommonPrefix `xml:"CommonPrefixes,omitempty"`
}

// ListBucketResultV2 ListObjectsV2响应
type ListBucketResultV2 struct {
	XMLName               string         `xml:"ListBucketResult"`
	Name                  string         `xml:"Name"`
	Prefix                string         `xml:"Prefix,omitempty"`
	KeyCount              int            `xml:"KeyCount"`
	MaxKeys               int            `xml:"MaxKeys"`
	Delimiter             string         `xml:"Delimiter,omitempty"`
	IsTruncated           bool           `xml:"IsTruncated"`
	ContinuationToken     string         `xml:"ContinuationToken,omitempty"`
	NextContinuationToken string         `xml:"NextContinuationToken,omitempty"`
	Contents              []Contents     `xml:"Contents,omitempty"`
	CommonPrefixes        []CommonPrefix `xml:"CommonPrefixes,omitempty"`
}

// Version 对象版本
type Version struct {
	Key          string `xml:"Key"`
	VersionID    string `xml:"VersionId"`
	IsLatest     bool   `xml:"IsLatest"`
	LastModified string `xml:"LastModified"`
	ETag         string `xml:"ETag"`
	Size         int64  `xml:"Size"`
	StorageClass string `xml:"StorageClass"`
	Owner        Owner  `xml:"Owner,omitempty"`
}

// DeleteMarkerEntry 删除标记
type DeleteMarkerEntry struct {
	Key          string `xml:"Key"`
	VersionID    string `xml:"VersionId"`
	IsLatest     bool   `xml:"IsLatest"`
	LastModified string `xml:"LastModified"`
	Owner        Owner  `xml:"Owner,omitempty"`
}

// ListVersionsResult ListObjectVersions响应
type ListVersionsResult struct {
	XMLName               xml.Name          `xml:"ListVersionsResult"`
	Name                  string            `xml:"Name"`
	Prefix                string            `xml:"Prefix,omitempty"`
	KeyMarker             string            `xml:"KeyMarker,omitempty"`
	VersionIDMarker       string            `xml:"VersionIdMarker,omitempty"`
	MaxKeys               int               `xml:"MaxKeys"`
	Delimiter             string            `xml:"Delimiter,omitempty"`
	IsTruncated           bool              `xml:"IsTruncated"`
	NextKeyMarker         string            `xml:"NextKeyMarker,omitempty"`
	NextVersionIDMarker   string            `xml:"NextVersionIdMarker,omitempty"`
	Versions              []Version         `xml:"Version,omitempty"`
	DeleteMarkers         []DeleteMarkerEntry `xml:"DeleteMarker,omitempty"`
	CommonPrefixes        []CommonPrefix    `xml:"CommonPrefixes,omitempty"`
}

// CORSRule CORS规则
type CORSRule struct {
	AllowedOrigins []string `xml:"AllowedOrigin"`           // 允许的来源
	AllowedMethods []string `xml:"AllowedMethod"`           // 允许的HTTP方法
	AllowedHeaders []string `xml:"AllowedHeader,omitempty"` // 允许的请求头
	ExposeHeaders  []string `xml:"ExposeHeader,omitempty"`  // 暴露的响应头
	MaxAgeSeconds  int      `xml:"MaxAgeSeconds,omitempty"` // 预检请求缓存时间（秒）
}

// CORSConfiguration CORS配置
type CORSConfiguration struct {
	XMLName xml.Name   `xml:"CORSConfiguration"`
	CORSRules []CORSRule `xml:"CORSRule"`
}

// Tag 对象标签
type Tag struct {
	Key   string `xml:"Key"`
	Value string `xml:"Value"`
}

// Tagging 对象标签配置
type Tagging struct {
	XMLName xml.Name `xml:"Tagging"`
	TagSet  TagSet   `xml:"TagSet"`
}

// TagSet 标签集合
type TagSet struct {
	Tags []Tag `xml:"Tag"`
}

// ==================== ACL 相关类型 ====================

// Grantee ACL被授权者
type Grantee struct {
	XMLName      xml.Name `xml:"Grantee"`
	Type         string   `xml:"xsi:type,attr,omitempty"` // CanonicalUser, Group, AmazonCustomerByEmail
	ID           string   `xml:"ID,omitempty"`             // CanonicalUser ID
	DisplayName  string   `xml:"DisplayName,omitempty"`    // CanonicalUser DisplayName
	URI          string   `xml:"URI,omitempty"`            // Group URI (如: http://acs.amazonaws.com/groups/global/AllUsers)
	EmailAddress string   `xml:"EmailAddress,omitempty"`   // EmailAddress
}

// Grant ACL授权
type Grant struct {
	XMLName xml.Name `xml:"Grant"`
	Grantee Grantee  `xml:"Grantee"`
	Permission string `xml:"Permission"` // READ, WRITE, READ_ACP, WRITE_ACP, FULL_CONTROL
}

// AccessControlList ACL配置
type AccessControlList struct {
	XMLName xml.Name `xml:"AccessControlList"`
	Owner   Owner    `xml:"Owner"`
	Grants  []Grant  `xml:"Grant"`
}

// AccessControlPolicy ACL策略（完整ACL响应）
type AccessControlPolicy struct {
	XMLName            xml.Name          `xml:"AccessControlPolicy"`
	Owner              Owner             `xml:"Owner"`
	AccessControlList  AccessControlList `xml:"AccessControlList"`
}

// InitiateMultipartUploadResult 初始化分片上传响应
type InitiateMultipartUploadResult struct {
	XMLName  string `xml:"InitiateMultipartUploadResult"`
	Bucket   string `xml:"Bucket"`
	Key      string `xml:"Key"`
	UploadID string `xml:"UploadId"`
}

// Part 分片信息
type Part struct {
	PartNumber int    `xml:"PartNumber"`
	ETag       string `xml:"ETag"`
}

// CompleteMultipartUpload 完成分片上传请求
type CompleteMultipartUpload struct {
	XMLName string `xml:"CompleteMultipartUpload"`
	Part    []Part `xml:"Part"`
}

// CompleteMultipartUploadResult 完成分片上传响应
type CompleteMultipartUploadResult struct {
	XMLName  string `xml:"CompleteMultipartUploadResult"`
	Location string `xml:"Location"`
	Bucket   string `xml:"Bucket"`
	Key      string `xml:"Key"`
	ETag     string `xml:"ETag"`
}

// CopyObjectResult 复制对象响应
type CopyObjectResult struct {
	XMLName      string `xml:"CopyObjectResult"`
	LastModified string `xml:"LastModified"`
	ETag         string `xml:"ETag"`
}

// DeleteRequest 批量删除请求
type DeleteRequest struct {
	XMLName string           `xml:"Delete"`
	Quiet   bool             `xml:"Quiet,omitempty"`
	Object  []ObjectToDelete `xml:"Object"`
}

// ObjectToDelete 待删除对象
type ObjectToDelete struct {
	Key       string `xml:"Key"`
	VersionID string `xml:"VersionId,omitempty"`
}

// DeleteResult 批量删除响应
type DeleteResult struct {
	XMLName string          `xml:"DeleteResult"`
	Deleted []DeletedObject `xml:"Deleted,omitempty"`
	Error   []DeleteError   `xml:"Error,omitempty"`
}

// DeletedObject 已删除对象
type DeletedObject struct {
	Key                   string `xml:"Key"`
	VersionID             string `xml:"VersionId,omitempty"`
	DeleteMarker          bool   `xml:"DeleteMarker,omitempty"`
	DeleteMarkerVersionID string `xml:"DeleteMarkerVersionId,omitempty"`
}

// DeleteError 删除错误
type DeleteError struct {
	Key       string `xml:"Key"`
	Code      string `xml:"Code"`
	Message   string `xml:"Message"`
	VersionID string `xml:"VersionId,omitempty"`
}

// ObjectMetadata 对象元数据
type ObjectMetadata struct {
	ContentType   string
	ContentLength int64
	ETag          string
	LastModified  time.Time
	UserMetadata  map[string]string
	StorageClass  string
	VersionID     string
}

// PartDetail 分片详细信息（用于 ListParts）
type PartDetail struct {
	PartNumber   int    `xml:"PartNumber"`
	ETag         string `xml:"ETag"`
	Size         int64  `xml:"Size"`
	LastModified string `xml:"LastModified"`
}

// VersioningConfiguration 版本控制配置
type VersioningConfiguration struct {
	XMLName xml.Name `xml:"VersioningConfiguration"`
	Status  string   `xml:"Status,omitempty"` // Enabled/Suspended
}

// ListPartsResult ListParts响应
type ListPartsResult struct {
	XMLName               xml.Name     `xml:"ListPartsResult"`
	Bucket                string       `xml:"Bucket"`
	Key                   string       `xml:"Key"`
	UploadID              string       `xml:"UploadId"`
	PartNumberMarker      int          `xml:"PartNumberMarker"`
	NextPartNumberMarker  int          `xml:"NextPartNumberMarker,omitempty"`
	MaxParts              int          `xml:"MaxParts"`
	IsTruncated           bool         `xml:"IsTruncated"`
	Initiator             Owner        `xml:"Initiator,omitempty"`
	Owner                 Owner        `xml:"Owner,omitempty"`
	StorageClass          string       `xml:"StorageClass,omitempty"`
	Parts                 []PartDetail `xml:"Part,omitempty"`
}

// Upload 分片上传会话信息（用于 ListMultipartUploads）
type Upload struct {
	Key          string `xml:"Key"`
	UploadID     string `xml:"UploadId"`
	Initiator    Owner  `xml:"Initiator,omitempty"`
	Owner        Owner  `xml:"Owner,omitempty"`
	StorageClass string `xml:"StorageClass,omitempty"`
	Initiated    string `xml:"Initiated"`
}

// ListMultipartUploadsResult ListMultipartUploads响应
type ListMultipartUploadsResult struct {
	XMLName              xml.Name `xml:"ListMultipartUploadsResult"`
	Bucket               string   `xml:"Bucket"`
	KeyMarker            string   `xml:"KeyMarker,omitempty"`
	UploadIDMarker       string   `xml:"UploadIdMarker,omitempty"`
	NextKeyMarker        string   `xml:"NextKeyMarker,omitempty"`
	NextUploadIDMarker   string   `xml:"NextUploadIdMarker,omitempty"`
	MaxUploads           int      `xml:"MaxUploads"`
	IsTruncated          bool     `xml:"IsTruncated"`
	Prefix               string   `xml:"Prefix,omitempty"`
	Delimiter            string   `xml:"Delimiter,omitempty"`
	Uploads              []Upload `xml:"Upload,omitempty"`
	CommonPrefixes       []CommonPrefix `xml:"CommonPrefixes,omitempty"`
}