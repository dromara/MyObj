package types

import "time"

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
