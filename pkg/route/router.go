// Package route 路由相关
package route

import (
	"net/http"

	"github.com/gorilla/mux"
)

// Router 路由器
var Router *mux.Router

// Name2URL 通过路由名称来获取 URL
func Name2URL(routeName string, pairs ...string) string {
	url, err := Router.Get(routeName).URL(pairs...)
	if err != nil {
		return ""
	}

	return url.String()
}

// GetRouteVariable 获取 URI 路由参数
func GetRouteVariable(parameterName string, r *http.Request) string {
	vars := mux.Vars(r)
	return vars[parameterName]
}
