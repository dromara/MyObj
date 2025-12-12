package models

type JsonResponse struct {
	// 状态码
	Code int `json:"code"`
	// 状态
	Message string `json:"message"`
	// 数据
	Data any `json:"data,omitempty"`
}

func NewJsonResponse(code int, message string, data any) *JsonResponse {
	return &JsonResponse{
		Code:    code,
		Message: message,
		Data:    data,
	}
}
