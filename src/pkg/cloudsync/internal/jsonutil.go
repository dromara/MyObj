package internal

import "encoding/json"

func JSONIntField(data []byte, key string) int {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		return 0
	}
	v, ok := m[key]
	if !ok {
		return 0
	}
	var n int
	_ = json.Unmarshal(v, &n)
	return n
}

func JSONStringField(data []byte, key string) string {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		return ""
	}
	v, ok := m[key]
	if !ok {
		return ""
	}
	var s string
	_ = json.Unmarshal(v, &s)
	return s
}

func JSONErrnoField(data []byte) int {
	return JSONIntField(data, "errno")
}
