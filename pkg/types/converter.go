// Package types 提供了一些类型转换的方法
package types

import (
	"goblog/pkg/logger"
	"strconv"
)

// Int64ToString 将 int64 转换为 string
func Int64ToString(value int64) string {
	return strconv.FormatInt(value, 10)
}

// StringToUint64 将字符串转换为 uint64
func StringToUint64(value string) uint64 {
	num, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		logger.LogError(err)
	}
	return num
}
