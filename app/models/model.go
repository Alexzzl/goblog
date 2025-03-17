// Package models 模型基类
package models

import (
	"goblog/pkg/types"
)

// BaseModel 模型基类
type BaseModel struct {
	ID uint64
}

// GetStringID 获取字符串 ID
func (bm BaseModel) GetStringID() string {
	return types.Uint64ToString(bm.ID)
}
