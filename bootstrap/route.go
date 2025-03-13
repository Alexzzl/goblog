package bootstrap

import (
	"goblog/pkg/route"
	"goblog/routes"

	"github.com/gorilla/mux"
)

// SetupRoute 路由初始化
func SetupRoute() *mux.Router {
	// 初始化路由
	router := mux.NewRouter()
	// 注册路由
	routes.RegisterWebRoutes(router)
	route.SetRoute(router)
	// 返回路由
	return router
}
