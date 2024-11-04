package global

import (
	"math/rand"
	"time"
)

// abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789
func GenerateRandomID(length int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// 可选的字符集合
	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	// 生成随机ID
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		randomIndex := r.Intn(len(charset))
		result[i] = charset[randomIndex]
	}

	return string(result)
}
