package types

import (
	"encoding/xml"
	"errors"
	"net/http"
)

// 定义业务错误类型，用于错误判断
// 注意：这些是error类型，与下面的S3Error类型不同
var (
	ErrBucketAlreadyExistsError      = errors.New("bucket already exists")
	ErrBucketNotFoundError           = errors.New("bucket not found")
	ErrBucketNotEmptyError           = errors.New("bucket is not empty")
	ErrObjectNotFoundError           = errors.New("object not found")
	ErrUploadNotFoundError           = errors.New("upload not found")
	ErrAccessDeniedError             = errors.New("access denied")
	ErrInvalidBucketNameError        = errors.New("invalid bucket name")
	ErrInvalidRangeError             = errors.New("invalid range")
	ErrContentMD5MismatchError       = errors.New("content MD5 mismatch")
	ErrInsufficientUserSpaceError    = errors.New("insufficient user space")
	ErrNoAvailableDiskError          = errors.New("no available disk")
	ErrUploadNotInProgressError      = errors.New("upload is not in progress")
	ErrInvalidPartNumberError        = errors.New("invalid part number")
	ErrPartsNotInAscendingOrderError = errors.New("parts must be in ascending order")
	ErrPartNotFoundError             = errors.New("part not found")
	ErrPartETagMismatchError         = errors.New("part etag mismatch")
	ErrNoPartsProvidedError          = errors.New("no parts provided")
	ErrBucketOrKeyMismatchError      = errors.New("bucket or key mismatch")
	ErrPolicyNotFoundError           = errors.New("policy not found")
	ErrLifecycleNotFoundError        = errors.New("lifecycle configuration not found")
	ErrCORSNotFoundError             = errors.New("CORS configuration not found")
	ErrACLRequiredError              = errors.New("ACL configuration is required")
	ErrACLOwnerMismatchError         = errors.New("ACL owner must be the object owner")
	ErrInvalidPermissionError        = errors.New("invalid permission")
	ErrTooManyTagsError              = errors.New("too many tags")
	ErrTagKeyTooLongError            = errors.New("tag key too long")
	ErrTagValueTooLongError          = errors.New("tag value too long")
)

// S3Error S3错误码定义
type S3Error struct {
	Code           string
	Description    string
	HTTPStatusCode int
}

// S3错误码常量
var (
	ErrNone = S3Error{
		Code:           "",
		Description:    "",
		HTTPStatusCode: http.StatusOK,
	}
	ErrAccessDenied = S3Error{
		Code:           "AccessDenied",
		Description:    "Access Denied",
		HTTPStatusCode: http.StatusForbidden,
	}
	ErrBucketAlreadyExists = S3Error{
		Code:           "BucketAlreadyExists",
		Description:    "The requested bucket name is not available",
		HTTPStatusCode: http.StatusConflict,
	}
	ErrBucketAlreadyOwnedByYou = S3Error{
		Code:           "BucketAlreadyOwnedByYou",
		Description:    "Your previous request to create the named bucket succeeded",
		HTTPStatusCode: http.StatusConflict,
	}
	ErrBucketNotEmpty = S3Error{
		Code:           "BucketNotEmpty",
		Description:    "The bucket you tried to delete is not empty",
		HTTPStatusCode: http.StatusConflict,
	}
	ErrNoSuchBucket = S3Error{
		Code:           "NoSuchBucket",
		Description:    "The specified bucket does not exist",
		HTTPStatusCode: http.StatusNotFound,
	}
	ErrNoSuchKey = S3Error{
		Code:           "NoSuchKey",
		Description:    "The specified key does not exist",
		HTTPStatusCode: http.StatusNotFound,
	}
	ErrInvalidAccessKeyId = S3Error{
		Code:           "InvalidAccessKeyId",
		Description:    "The AWS access key ID you provided does not exist in our records",
		HTTPStatusCode: http.StatusForbidden,
	}
	ErrInvalidBucketName = S3Error{
		Code:           "InvalidBucketName",
		Description:    "The specified bucket is not valid",
		HTTPStatusCode: http.StatusBadRequest,
	}
	ErrInvalidRange = S3Error{
		Code:           "InvalidRange",
		Description:    "The requested range is not satisfiable",
		HTTPStatusCode: http.StatusRequestedRangeNotSatisfiable,
	}
	ErrSignatureDoesNotMatch = S3Error{
		Code:           "SignatureDoesNotMatch",
		Description:    "The request signature we calculated does not match the signature you provided",
		HTTPStatusCode: http.StatusForbidden,
	}
	ErrInternalError = S3Error{
		Code:           "InternalError",
		Description:    "We encountered an internal error. Please try again",
		HTTPStatusCode: http.StatusInternalServerError,
	}
	ErrInvalidArgument = S3Error{
		Code:           "InvalidArgument",
		Description:    "Invalid Argument",
		HTTPStatusCode: http.StatusBadRequest,
	}
	ErrMethodNotAllowed = S3Error{
		Code:           "MethodNotAllowed",
		Description:    "The specified method is not allowed against this resource",
		HTTPStatusCode: http.StatusMethodNotAllowed,
	}
	ErrNoSuchUpload = S3Error{
		Code:           "NoSuchUpload",
		Description:    "The specified multipart upload does not exist",
		HTTPStatusCode: http.StatusNotFound,
	}
	ErrEntityTooLarge = S3Error{
		Code:           "EntityTooLarge",
		Description:    "Your proposed upload exceeds the maximum allowed object size",
		HTTPStatusCode: http.StatusBadRequest,
	}
	ErrInvalidPart = S3Error{
		Code:           "InvalidPart",
		Description:    "One or more of the specified parts could not be found",
		HTTPStatusCode: http.StatusBadRequest,
	}
	ErrInvalidPartOrder = S3Error{
		Code:           "InvalidPartOrder",
		Description:    "The list of parts was not in ascending order",
		HTTPStatusCode: http.StatusBadRequest,
	}
	ErrNoSuchCORSConfiguration = S3Error{
		Code:           "NoSuchCORSConfiguration",
		Description:    "The CORS configuration does not exist",
		HTTPStatusCode: http.StatusNotFound,
	}
	ErrNoSuchPolicy = S3Error{
		Code:           "NoSuchPolicy",
		Description:    "The specified policy does not exist",
		HTTPStatusCode: http.StatusNotFound,
	}
	ErrNoSuchLifecycleConfiguration = S3Error{
		Code:           "NoSuchLifecycleConfiguration",
		Description:    "The lifecycle configuration does not exist",
		HTTPStatusCode: http.StatusNotFound,
	}
)

// MapErrorToS3Error 将业务错误映射到S3错误响应
func MapErrorToS3Error(err error) S3Error {
	// 使用业务错误类型（error类型）进行判断
	if errors.Is(err, ErrBucketAlreadyExistsError) {
		// 返回HTTP响应错误（S3Error类型）
		return ErrBucketAlreadyExists
	}
	if errors.Is(err, ErrBucketNotFoundError) {
		return ErrNoSuchBucket
	}
	if errors.Is(err, ErrBucketNotEmptyError) {
		return ErrBucketNotEmpty
	}
	if errors.Is(err, ErrObjectNotFoundError) {
		return ErrNoSuchKey
	}
	if errors.Is(err, ErrUploadNotFoundError) {
		return ErrNoSuchUpload
	}
	if errors.Is(err, ErrAccessDeniedError) {
		return ErrAccessDenied
	}
	if errors.Is(err, ErrInvalidBucketNameError) {
		return ErrInvalidBucketName
	}
	if errors.Is(err, ErrInvalidRangeError) {
		return ErrInvalidRange
	}
	if errors.Is(err, ErrContentMD5MismatchError) {
		return ErrInvalidArgument
	}
	if errors.Is(err, ErrInsufficientUserSpaceError) {
		return ErrInvalidArgument
	}
	if errors.Is(err, ErrNoAvailableDiskError) {
		return ErrInternalError
	}
	if errors.Is(err, ErrUploadNotInProgressError) {
		return ErrInvalidArgument
	}
	if errors.Is(err, ErrInvalidPartNumberError) {
		return ErrInvalidPart
	}
	if errors.Is(err, ErrPartsNotInAscendingOrderError) {
		return ErrInvalidPartOrder
	}
	if errors.Is(err, ErrPartNotFoundError) {
		return ErrInvalidPart
	}
	if errors.Is(err, ErrPartETagMismatchError) {
		return ErrInvalidPart
	}
	if errors.Is(err, ErrNoPartsProvidedError) {
		return ErrInvalidArgument
	}
	if errors.Is(err, ErrBucketOrKeyMismatchError) {
		return ErrInvalidArgument
	}
	if errors.Is(err, ErrPolicyNotFoundError) {
		return ErrNoSuchPolicy
	}
	if errors.Is(err, ErrLifecycleNotFoundError) {
		return ErrNoSuchLifecycleConfiguration
	}
	if errors.Is(err, ErrCORSNotFoundError) {
		return ErrNoSuchCORSConfiguration
	}
	// 默认返回内部错误
	return ErrInternalError
}

// ErrorResponse S3错误响应XML结构
type ErrorResponse struct {
	XMLName   xml.Name `xml:"Error"`
	Code      string   `xml:"Code"`
	Message   string   `xml:"Message"`
	Resource  string   `xml:"Resource,omitempty"`
	RequestID string   `xml:"RequestId,omitempty"`
}

// WriteErrorResponse 写入S3错误响应
func WriteErrorResponse(w http.ResponseWriter, r *http.Request, err S3Error, resource string) {
	errorResp := ErrorResponse{
		Code:      err.Code,
		Message:   err.Description,
		Resource:  resource,
		RequestID: r.Header.Get("X-Amz-Request-Id"),
	}

	w.Header().Set("Content-Type", "application/xml")
	w.Header().Set("X-Amz-Request-Id", errorResp.RequestID)
	w.WriteHeader(err.HTTPStatusCode)

	xmlData, _ := xml.MarshalIndent(errorResp, "", "  ")
	w.Write([]byte(xml.Header))
	w.Write(xmlData)
}
