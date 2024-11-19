package utils

import "encoding/json"

func PrintJSON(data interface{}) string {
	value, err := json.Marshal(data)
	if err != nil {
		return ""
	}

	return string(value)
}
