package util

import "time"

type TimeUtil struct{}

// GetTimestamp 获取毫秒时间戳
func (TimeUtil) GetTimestamp() int64 {
	return time.Now().UnixMilli()
}

// GetTime 获取当前时间字符串
func (TimeUtil) GetTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

// GetTimeStrByTimestamp 获取时间字符串
func (TimeUtil) GetTimeStrByTimestamp(timestamp int64) string {
	return time.UnixMilli(timestamp).Format("2006-01-02 15:04:05")
}

// GetTimeByTimestamp 从毫秒获取时间
func (TimeUtil) GetTimeByTimestamp(timestamp int64) time.Time {
	return time.UnixMilli(timestamp)
}

// IsAfter 判断时间是否在当前时间之后
func (TimeUtil) IsAfter(t time.Time) bool {
	return time.Now().After(t)
}

// IsBefore 判断时间是否在当前时间之前
func (TimeUtil) IsBefore(t time.Time) bool {
	return time.Now().Before(t)
}

// GetAfterTimestamp 获取指定时间到当前时间的秒数
func (TimeUtil) GetAfterTimestamp(t time.Time) int64 {
	return time.Now().Sub(t).Milliseconds() / 1000
}
