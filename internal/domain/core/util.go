package core

import (
	"encoding/json"
	"fmt"
	"strings"
)

func GenerateCacheKey(prefix string, params any) string {
	return fmt.Sprintf("%s:%v", prefix, params)
}

func GenerateCacheKeyParams(params ...any) string {
	var builder strings.Builder
	for i, param := range params {
		fmt.Fprintf(&builder, "%v", param)
		last := len(params) - 1
		if i != last {
			fmt.Fprintf(&builder, "-")
		}
	}
	return builder.String()
}

func Serialize(data any) ([]byte, error) {
	return json.Marshal(data)
}

func Deserialize(data []byte, output any) error {
	return json.Unmarshal(data, output)
}
