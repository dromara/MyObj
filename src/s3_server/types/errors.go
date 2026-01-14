package types

import (
	"encoding/xml"
	"net/http"
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
)

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
