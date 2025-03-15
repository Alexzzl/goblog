// Package middlewares 存放应用中间件
package middlewares

import (
	"net/http"
)

// ForceHTML 强制请求为 HTML
func ForceHTML(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 设置标头
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		// 分发处理
		next.ServeHTTP(w, r)
	})
}
