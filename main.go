package main

import (
	"database/sql"
	"net/http"
	"strings"

	"goblog/bootstrap"
	"goblog/pkg/database"

	"github.com/gorilla/mux"
)

var db *sql.DB

var router *mux.Router

func forceHTMLMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. 设置标头
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		// 2. 继续处理请求
		next.ServeHTTP(w, r)
	})
}

func removeTrailingSlash(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. 除首页以外，移除所有请求路径后面的斜杆
		if r.URL.Path != "/" {
			r.URL.Path = strings.TrimSuffix(r.URL.Path, "/")
		}

		// 2. 将请求传递下去
		next.ServeHTTP(w, r)
	})
}

func main() {
	database.Initialize()
	db = database.DB

	// route.Initialize()
	// router = route.Router

	bootstrap.SetupDB()
	router = bootstrap.SetupRoute()

	// router.HandleFunc("/articles", articlesIndexHandler).Methods("GET").Name("articles.index")
	//router.HandleFunc("/articles", articlesStoreHandler).Methods("POST").Name("articles.store")
	//router.HandleFunc("/articles/create", articlesCreateHandler).Methods("GET").Name("articles.create")
	// router.HandleFunc("/articles/{id:[0-9]+}/edit", articlesEditHandler).Methods("GET").Name("articles.edit")
	// router.HandleFunc("/articles/{id:[0-9]+}", articlesUpdateHandler).Methods("POST").Name("articles.update")
	// router.HandleFunc("/articles/{id:[0-9]+}/delete", articlesDeleteHandler).Methods("POST").Name("articles.delete")

	// 中间件：强制内容类型为 HTML
	router.Use(forceHTMLMiddleware)

	http.ListenAndServe(":3000", removeTrailingSlash(router))
}
