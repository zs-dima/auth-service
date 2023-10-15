package tool

import (
	"os"
	"strings"
)

func GetFileValue(key string) string {
	value := os.Getenv(key)

	if strings.TrimSpace(value) == "" {
		fileKey := key + "_file"
		filePath := os.Getenv(fileKey)
		if strings.TrimSpace(filePath) == "" {
			return ""
		}
		fileBytes, err := os.ReadFile(filePath)
		if err != nil {
			return ""
		}
		value = string(fileBytes)
	}
	return value
}
