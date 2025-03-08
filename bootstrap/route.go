package bootstrap

import (
	"goblog/routes"

	"github.com/gorilla/mux"
)

// SetupRoute 路由初始化
func SetupRoute() *mux.Router {
	// 初始化路由
	Router := mux.NewRouter()
	// 注册路由
	routes.RegisterWebRoutes(Router)
	// 返回路由
	return Router
}
