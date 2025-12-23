package custom_type

import (
	"database/sql"
	"database/sql/driver"
	"time"
)

type JsonTime time.Time

// MarshalJSON 实现 MarshalJSON 接口
func (t JsonTime) MarshalJSON() ([]byte, error) {
	if time.Time(t).IsZero() {
		return []byte(`""`), nil
	}
	formatted := time.Time(t).Format("2006-01-02 15:04:05")
	return []byte(`"` + formatted + `"`), nil
}

// UnmarshalJSON 实现 UnmarshalJSON 接口
func (t *JsonTime) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}
	parsed, err := time.Parse(`"2006-01-02 15:04:05"`, string(data))
	if err != nil {
		return err
	}
	*t = JsonTime(parsed)
	return nil
}

func (t *JsonTime) Scan(value interface{}) (err error) {
	nullTime := &sql.NullTime{}
	err = nullTime.Scan(value)
	*t = JsonTime(nullTime.Time)
	return
}

func (t JsonTime) Value() (driver.Value, error) {
	if t.IsZero() {
		// 返回 nil，让数据库保存 NULL（适用于可选字段）
		return nil, nil
	}
	return time.Time(t), nil
}
func (t JsonTime) GormDataType() string {
	return "datetime"
}

// Before 判断时间是否在另一个时间之前
func (t *JsonTime) Before(u JsonTime) bool {
	return time.Time(*t).Before(time.Time(u))
}

// After 判断时间是否在另一个时间之后
func (t *JsonTime) After(u JsonTime) bool {
	return time.Time(*t).After(time.Time(u))
}

// Sub 计算两个时间的差值
func (t *JsonTime) Sub(u JsonTime) time.Duration {
	return time.Time(*t).Sub(time.Time(u))
}

// Format 格式化时间
func (t JsonTime) Format(layout string) string {
	return time.Time(t).Format(layout)
}

// Add 增加时间
func (t *JsonTime) Add(d time.Duration) JsonTime {
	return JsonTime(time.Time(*t).Add(d))
}

// Equal 判断两个时间是否相等
func (t *JsonTime) Equal(u JsonTime) bool {
	return time.Time(*t).Equal(time.Time(u))
}

// Compare 比较两个时间
// 如果 t 在 u 之前，则返回 -1;如果 t 在 u 之后，则返回 +1;如果它们相同，则返回 0。
func (t *JsonTime) Compare(u JsonTime) int {
	return time.Time(*t).Compare(time.Time(u))
}

func (t *JsonTime) Unix() int64 {
	return time.Time(*t).Unix()
}

func (t *JsonTime) UnixMilli() int64 {
	return time.Time(*t).UnixMilli()
}

func (t *JsonTime) UnixMicro() int64 {
	return time.Time(*t).UnixMicro()
}

func (t *JsonTime) UnixNano() int64 {
	return time.Time(*t).UnixNano()
}

func (t *JsonTime) IsZero() bool {
	return time.Time(*t).IsZero()
}

// Now 返回当前时间
func Now() JsonTime {
	return JsonTime(time.Now())
}

// TimePtr 将 JsonTime 转换为指针
func TimePtr(t JsonTime) *JsonTime {
	return &t
}

func (t JsonTime) ToTime() time.Time {
	return time.Time(t)
}
