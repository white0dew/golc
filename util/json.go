package util

import "encoding/json"

func ToString(data interface{}) string {
	bytes, err := json.Marshal(data)
	if err != nil {
		return ""
	}
	return string(bytes)
}
