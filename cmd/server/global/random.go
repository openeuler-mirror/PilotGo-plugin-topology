/*
 * Copyright (c) KylinSoft  Co., Ltd. 2024.All rights reserved.
 * PilotGo-plugin-topology licensed under the Mulan Permissive Software License, Version 2. 
 * See LICENSE file for more details.
 * Author: Wangjunqi123 <wangjunqi@kylinos.cn>
 * Date: Mon Nov 4 14:30:13 2024 +0800
 */
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
