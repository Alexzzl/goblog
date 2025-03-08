// Package controllers 应用控制层
package controllers

import (
	"database/sql"
	"fmt"
	"goblog/pkg/logger"
	"goblog/pkg/route"
	"goblog/pkg/types"
	"html/template"
	"net/http"
)

// ArticlesController 文章相关页面
type ArticlesController struct {
}

// type Article struct {
// 	Title, Body string
// 	ID          int64
// }

// func getArticleByID(id string) (Article, error) {
// 	article := Article{}
// 	query := "SELECT * FROM articles WHERE id = ?"
// 	err := DB.QueryRow(query, id).Scan(&article.ID, &article.Title, &article.Body)
// 	return article, err
// }

// Show 文章详情页面
func (*ArticlesController) Home(w http.ResponseWriter, r *http.Request) {
	// 1. 获取 URL 参数
	id := route.GetRouteVariable("id", r)

	//  2. 读取对应的文章数据
	article, err := getArticleByID(id)

	// 3. 如果出现错误
	if err != nil {
		if err == sql.ErrNoRows {
			// 3.1 数据没找到 返回 404 页面
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, "<h1>404 文章未找到 :(</h1>"+
				"<p>如有疑惑，请联系我们。</p>")
		} else {
			// 3.2 数据库错误
			logger.LogError(err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "500 服务器内部错误 文章创建失败，错误信息为：%v", err)
		}
	} else {
		// 4. 读取成功, 显示文章
		tmpl, err := template.New("show.gohtml").
			Funcs(template.FuncMap{
				"RouteName2URL": route.Name2URL,
				"Int64ToString": types.Int64ToString,
			}).ParseFiles("resources/views/articles/show.gohtml")
		logger.LogError(err)

		err = tmpl.Execute(w, article)
		logger.LogError(err)
		//fmt.Fprint(w, "<h1>"+article.Title+"</h1>")
	}
}

// Index 文章列表页面
func (*ArticlesController) About(w http.ResponseWriter, r *http.Request) {
}
