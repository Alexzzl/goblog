package main

import (
	"context"
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"net"

	"github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"goblog/pkg/route"
	"golang.org/x/crypto/ssh"
)

// var router = mux.NewRouter()
var router *mux.Router
var db *sql.DB
var sshClient *ssh.Client // 添加全局 SSH 客户端变量

func initDB() {
	// 设置 SSH 配置
	sshConfig := &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			ssh.Password("zzl19961120..."), // 也可以使用密钥认证
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 生产环境建议使用具体的主机密钥
	}

	// 建立 SSH 连接
	var err error
	sshClient, err = ssh.Dial("tcp", "119.8.108.55:22", sshConfig)
	checkError(err)

	// 通过 SSH 建立 MySQL 连接
	mysqlAddr := "127.0.0.1:3306"
	tunnel, err := sshClient.Dial("tcp", mysqlAddr)
	checkError(err)

	config := mysql.Config{
		User:                 "root",
		Passwd:               "zzl19961120...",
		Net:                  "tcp",
		Addr:                 "127.0.0.1:3306",
		DBName:               "goblog",
		AllowNativePasswords: true,
	}

	// 使用 RegisterDialContext 注册自定义连接器
	mysql.RegisterDialContext("tcp", func(ctx context.Context, addr string) (net.Conn, error) {
		return tunnel, nil
	})

	// 准备数据库连接池
	db, err = sql.Open("mysql", config.FormatDSN())
	//fmt.Println(config.FormatDSN())
	checkError(err)

	// 设置最大连接数
	db.SetMaxOpenConns(25)
	// 设置最大空闲连接数
	db.SetMaxIdleConns(25)
	// 设置每个链接的过期时间
	db.SetConnMaxLifetime(5 * time.Minute)

	// 尝试连接，失败会报错
	err = db.Ping()
	checkError(err)
}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "<h1>Hello, 欢迎来到 Alex goblog！</h1>")
}

func aboutHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "此博客是用以记录编程笔记，如您有反馈或建议，请联系 "+
		"<a href=\"mailto:805119233@qq.com\">805119233@qq.com</a>")
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprint(w, "<h1>请求页面未找到 :(</h1>"+
		"<p>如有疑惑，请联系我们。</p>")
}

// Artcle 对应一条文章数据
type Article struct {
	Title, Body string
	ID          int64
}

// Link 方法用来生成文章链接
func (a Article) Link() string {
	showURL, err := router.Get("articles.show").URL("id", strconv.FormatInt(a.ID, 10))
	if err != nil {
		checkError(err)
		return "#"
	}
	return showURL.String()
}

func (a Article) Delete() (rowsAffected int64, err error) {
	rs, err := db.Exec("DELETE FROM articles WHERE id = ?", strconv.FormatInt(a.ID, 10))

	if err != nil {
		return 0, err
	}

	// 删除成功
	if n, _ := rs.RowsAffected(); n > 0 {
		return n, nil
	}
	return 0, nil
}

func getRouteVariable(parameterName string, r *http.Request) string {
	vars := mux.Vars(r)
	return vars[parameterName]
}

func getArticleByID(id string) (Article, error) {
	article := Article{}
	query := "SELECT * FROM articles WHERE id = ?"
	err := db.QueryRow(query, id).Scan(&article.ID, &article.Title, &article.Body)
	return article, err
}

func articlesShowHandler(w http.ResponseWriter, r *http.Request) {
	// 1. 获取 URL 参数
	id := getRouteVariable("id", r)

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
			checkError(err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "500 服务器内部错误 文章创建失败，错误信息为：%v", err)
		}
	} else {
		// 4. 读取成功, 显示文章
		tmpl, err := template.New("show.gohtml").
			Funcs(template.FuncMap{
				"RouteName2URL": route.Name2URL,
				"Int64ToString": Int64ToString,
			}).ParseFiles("resources/views/articles/show.gohtml")
		checkError(err)

		err = tmpl.Execute(w, article)
		checkError(err)
		//fmt.Fprint(w, "<h1>"+article.Title+"</h1>")
	}
}

// Int64ToString 将 int64 转换为 string
func Int64ToString(value int64) string {
	return strconv.FormatInt(value, 10)
}

func articlesIndexHandler(w http.ResponseWriter, r *http.Request) {
	// 1. 执行查询语句，返回一个结果集
	rows, err := db.Query("SELECT * FROM articles")
	checkError(err)
	defer rows.Close()

	var articles []Article
	// 2. 循环读取结果
	for rows.Next() {
		var article Article
		// 2.1 扫描每一行的结果并赋值到一个 article 对象中
		err = rows.Scan(&article.ID, &article.Title, &article.Body)
		checkError(err)
		// 2.2 将 article 追加到 articles 的这个数组中
		articles = append(articles, article)
	}

	// 2.3 检测遍历时是否发生错误
	err = rows.Err()
	checkError(err)

	/// 3.加载模块
	tmpl, err := template.ParseFiles("resources/views/articles/index.gohtml")
	checkError(err)

	// 4. 渲染模板，将所有文章的数据传输进去
	err = tmpl.Execute(w, articles)
	checkError(err)
}

func articlesStoreHandler(w http.ResponseWriter, r *http.Request) {
	title := r.PostFormValue("title")
	body := r.PostFormValue("body")

	errors := validateArticleFormData(title, body)

	// 检查是否有错误
	if len(errors) == 0 {
		lastInsertID, err := saveArticleToDB(title, body)
		if lastInsertID > 0 {
			fmt.Fprint(w, "插入成功，ID 为"+strconv.FormatInt(lastInsertID, 10))
		} else {
			checkError(err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "500 服务器内部错误 文章创建失败，错误信息为：%v", err)
		}
	} else {

		storeURL, _ := router.Get("articles.store").URL()

		data := ArticlesFormData{
			Title:  title,
			Body:   body,
			URL:    storeURL,
			Errors: errors,
		}
		tmpl, err := template.ParseFiles("resources/views/articles/create.gohtml")
		if err != nil {
			panic(err)
		}

		err = tmpl.Execute(w, data)
		if err != nil {
			panic(err)
		}
	}
}

func saveArticleToDB(title, body string) (int64, error) {
	// 变量初始化
	var (
		id   int64
		err  error
		rs   sql.Result
		stmt *sql.Stmt
	)

	// 1. 获取一个 prepare 声明语句
	stmt, err = db.Prepare("INSERT INTO articles (title, body) VALUES(?,?)")
	// 2. 检查错误
	if err != nil {
		return 0, err
	}

	// 3. 在此函数运行结束后关闭此语句，防止占用 SQL 连接
	defer stmt.Close()

	// 4. 执行 SQL 语句
	rs, err = stmt.Exec(title, body)
	if err != nil {
		return 0, err
	}

	// 5. 获取插入记录的 ID
	if id, err = rs.LastInsertId(); id > 0 {
		return id, err
	}

	return 0, err
}

func articlesUpdateHandler(w http.ResponseWriter, r *http.Request) {
	// 1. 获取 URL 参数
	id := getRouteVariable("id", r)

	//  2. 读取对应的文章数据
	_, err := getArticleByID(id)

	// 3. 如果出现错误
	if err != nil {
		if err == sql.ErrNoRows {
			// 3.1 数据没找到 返回 404 页面
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, "<h1>404 文章未找到 :(</h1>"+
				"<p>如有疑惑，请联系我们。</p>")
		} else {
			// 3.2 数据库错误
			checkError(err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "500 服务器内部错误 文章更新失败，错误信息为：%v", err)
		}
	} else {
		// 4. 未出现错误

		// 4.1 表单验证
		title := r.PostFormValue("title")
		body := r.PostFormValue("body")

		errors := validateArticleFormData(title, body)

		if len(errors) == 0 {
			// 4.2 表单验证通过，更新数据

			query := "UPDATE articles SET title = ?, body = ? WHERE id = ?"
			rs, err := db.Exec(query, title, body, id)
			if err != nil {
				checkError(err)
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "500 服务器内部错误 文章更新失败，错误信息为：%v", err)
			}

			// 更新成功，跳转到文章详情页
			if n, _ := rs.RowsAffected(); n > 0 {
				showURL, _ := router.Get("articles.show").URL("id", id)
				http.Redirect(w, r, showURL.String(), http.StatusFound)
			} else {
				fmt.Fprint(w, "您没有做任何更改！")
			}

		} else {
			// 4.3 表单验证不通过，显示理由
			updateURL, _ := router.Get("articles.update").URL("id", id)
			data := ArticlesFormData{
				Title:  title,
				Body:   body,
				URL:    updateURL,
				Errors: errors,
			}
			tmpl, err := template.ParseFiles("resources/views/articles/edit.gohtml")
			checkError(err)

			err = tmpl.Execute(w, data)
			checkError(err)
		}
	}
}

func articlesEditHandler(w http.ResponseWriter, r *http.Request) {
	// 1. 获取 URL 参数
	id := getRouteVariable("id", r)

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
			checkError(err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "500 服务器内部错误 文章创建失败，错误信息为：%v", err)
		}
	} else {
		// 4. 读取成功, 显示表单
		updateURL, _ := router.Get("articles.update").URL("id", id)
		data := ArticlesFormData{
			Title:  article.Title,
			Body:   article.Body,
			URL:    updateURL,
			Errors: nil,
		}
		tmpl, err := template.ParseFiles("resources/views/articles/edit.gohtml")
		checkError(err)

		err = tmpl.Execute(w, data)
		checkError(err)
	}
}

func validateArticleFormData(title, body string) map[string]string {
	errors := make(map[string]string)

	// 验证标题
	if title == "" {
		errors["title"] = "标题不能为空"
	} else if utf8.RuneCountInString(title) < 3 || len(title) > 40 {
		errors["title"] = "标题长度需介于 3-40"
	}

	// 验证内容
	if body == "" {
		errors["body"] = "内容不能为空"
	} else if utf8.RuneCountInString(body) < 10 {
		errors["body"] = "内容长度需大于或等于 10 个字节"
	}

	return errors
}

// ArticlesFormData 创建博文表单数据
type ArticlesFormData struct {
	Title, Body string
	URL         *url.URL
	Errors      map[string]string
}

func articlesCreateHandler(w http.ResponseWriter, r *http.Request) {

	storeURL, _ := router.Get("articles.store").URL()
	data := ArticlesFormData{
		Title:  "",
		Body:   "",
		URL:    storeURL,
		Errors: nil,
	}
	tmpl, err := template.ParseFiles("resources/views/articles/create.gohtml")
	if err != nil {
		panic(err)
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		panic(err)
	}
}

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

// func createTables() {
// 	createArticlesSQL := `CREATE TABLE IF NOT EXISTS articles(
// 	id bigint(20) PRIMARY KEY AUTO_INCREMENT NOT NULL,
// 	title varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
// 	body longtext COLLATE utf8mb4_unicode_ci);`

// 	_, err := db.Exec(createArticlesSQL)
// 	checkError(err)
// }

func articlesDeleteHandler(w http.ResponseWriter, r *http.Request) {
	// 1. 获取 URL 参数
	id := getRouteVariable("id", r)

	// 2. 读取对应的文章数据
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
			checkError(err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "500 服务器内部错误 文章创建失败，错误信息为：%v", err)
		}
	} else {
		// 4. 未出现错误，执行删除操作
		rowsAffected, err := article.Delete()

		// 4.1 如果出现错误
		if err != nil {
			checkError(err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "500 服务器内部错误 文章删除失败，错误信息为：%v", err)
		} else {
			// 4.2 删除成功，跳转到文章列表页
			if rowsAffected > 0 {
				indexURL, _ := router.Get("articles.index").URL()
				http.Redirect(w, r, indexURL.String(), http.StatusFound)
			} else {
				// Edge case
				w.WriteHeader(http.StatusNotFound)
				fmt.Fprint(w, "<h1>404 文章未找到 :(</h1>"+
					"<p>如有疑惑，请联系我们。</p>")
			}
		}
	}
}

func main() {
	initDB()
	defer sshClient.Close() // 移到 main 函数中
	//createTables()

	route.Initialize()
	router = route.Router

	router.HandleFunc("/", homeHandler).Methods("GET").Name("home")
	router.HandleFunc("/about", aboutHandler).Methods("GET").Name("about")
	router.HandleFunc("/articles/{id:[0-9]+}", articlesShowHandler).Methods("GET").Name("articles.show")
	router.HandleFunc("/articles", articlesIndexHandler).Methods("GET").Name("articles.index")
	router.HandleFunc("/articles", articlesStoreHandler).Methods("POST").Name("articles.store")
	router.HandleFunc("/articles/create", articlesCreateHandler).Methods("GET").Name("articles.create")
	router.HandleFunc("/articles/{id:[0-9]+}/edit", articlesEditHandler).Methods("GET").Name("articles.edit")
	router.HandleFunc("/articles/{id:[0-9]+}", articlesUpdateHandler).Methods("POST").Name("articles.update")
	router.HandleFunc("/articles/{id:[0-9]+}/delete", articlesDeleteHandler).Methods("POST").Name("articles.delete")

	// 自定义 404 页面
	router.NotFoundHandler = http.HandlerFunc(notFoundHandler)

	// 中间件：强制内容类型为 HTML
	router.Use(forceHTMLMiddleware)

	http.ListenAndServe(":3000", removeTrailingSlash(router))
}
