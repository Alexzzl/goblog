// Package controllers 应用控制层
package controllers

import (
	"fmt"
	"net/http"
)

// PagesController 处理静态页面
type PagesController struct {
}

// Home 首页
func (*PagesController) Home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "<h1>Hello, 欢迎来到 Alex goblog！</h1>")
	fmt.Fprint(w, "<a href='/articles'>文章列表</a>")
	fmt.Fprint(w, "<p>"+
		"通过本博客您可以了解到 Alex 在编程路上的故事，"+
		"也希望您能通过博客的形式和我交流。"+
		"</p>")
}

// About 关于我们页面
func (*PagesController) About(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "此博客是用以记录编程笔记，如您有反馈或建议，请联系 "+
		"<a href=\"mailto:805119233@qq.com\">805119233@qq.com</a>")
}

// NotFound 404 页面
func (*PagesController) NotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprint(w, "<h1>请求页面未找到 :(</h1>"+
		"<p>如有疑惑，请联系我们。</p>")
}
