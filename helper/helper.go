package helper

import (
	"math/rand"
)

func GenerateRandomString() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	// 随机生成字符串的长度，假设长度范围为 5 到 15
	length := rand.Intn(11) + 5

	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}

	return string(result)
}
