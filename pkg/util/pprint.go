package util

import (
	"encoding/json"
	"fmt"
)

func PrettyPrintMap(v map[string]interface{}) string {
	result, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Println("error:", err)
	}
	return string(result)
}
