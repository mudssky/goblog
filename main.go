package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	"./models/category"
	"./models/post"
	"./models/session"
	"./models/user"
	"github.com/globalsign/mgo/bson"
)

// 日志对象，分别用于输出不同级别的日志
var (
	LogDebug   *log.Logger
	LogWarning *log.Logger
	LogPanic   *log.Logger
	LogFatal   *log.Logger
	// 初始化过程中，一次性解析所有模板
	// 由于每次请求都是重新解析的，是动态页面，所以一次性解析不适用
	// AllTemplate *template.Template
	// SessionMap map[string]interface{}
)

// MessagePage 存储跳转页渲染信息的结构体
type MessagePage struct {
	Message string
	URL     string
}

func init() {
	// 创建输出日志文件
	/*logFile, err := os.Create("./" + time.Now().Format("20060102") + ".log")
	if err != nil {
		fmt.Println(err)
	}*/
	logFile := os.Stdout
	errFile := os.Stderr
	LogDebug = log.New(logFile, "[Debug]:", log.Ldate|log.Ltime|log.Llongfile)
	LogWarning = log.New(logFile, "[Warning]:", log.Ldate|log.Ltime|log.Llongfile)
	LogPanic = log.New(errFile, "[Panic]:", log.Ldate|log.Ltime|log.Llongfile)
	LogFatal = log.New(errFile, "[Fatal]:", log.Ldate|log.Ltime|log.Llongfile)
	/*
		// init函数里面用:=赋值 相当于给局部变量复制，所以我们要用=复制
		var alltemperr error
		AllTemplate, alltemperr = template.ParseFiles("views/index.html", "views/components/navbar.html", "views/components/footer.html","views/components/header.html", "views/signin.html", "views/signup.html")
		if alltemperr != nil {
			LogPanic.Panicln("parse template views/index.html failed")
		}
		LogDebug.Println("parse template succeed:", AllTemplate.DefinedTemplates())
	*/
}

// IndexHandle 处理首页的逻辑
func IndexHandle(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		posts := post.Post{}
		postindex, err := posts.GetPostsIndex()
		LogDebug.Println(postindex)
		LogDebug.Println("path", "/")
		// 解析需要的模板
		temp, err := template.ParseFiles("views/index.html", "views/components/navbar.html", "views/components/footer.html",
			"views/components/header.html")
		// temp, err := template.ParseFiles("views/index.html")
		if err != nil {
			LogPanic.Panicln("parse template views/index.html failed", err)
		}
		// 获取session中的数据
		sess := session.New()
		sess = sess.SessionStart(w, r)
		// SessionMap = sess.Data
		// LogDebug.Println(sess.Data["signin"])
		sess.Data["postindex"] = postindex
		err = temp.ExecuteTemplate(w, "index", sess.Data)
		if err != nil {
			LogPanic.Panicln("parse template views/index.html failed", err)
		}
		// http包默认的路由规则，按照最长前缀匹配
		// 所有路径都可以匹配到/,那样404页面就失去作用了，所以当/匹配失败的时候展示404页面
	} else {
		temp, err := template.ParseFiles("views/404.html")
		// temp, err := template.ParseFiles("views/index.html")
		if err != nil {
			LogPanic.Panicln("parse template views/404.html failed", err)
		}
		err = temp.ExecuteTemplate(w, "404.html", nil)
		if err != nil {
			LogPanic.Panicln("parse template views/404.html failed", err)
		}
	}

}
func validateSignin(form map[string][]string) (errorMessage string) {
	username := form["Username"][0]
	password := form["Password"][0]
	if matched, _ := regexp.MatchString(`\w{4,20}`, username); !matched {
		errorMessage = "请输入正确的用户名"
		return
	}
	if !user.CheckLogin(username, password) {
		errorMessage = "用户名或者密码错误"
	}
	return
}

// SigninHandle 处理登录页的逻辑
func SigninHandle(w http.ResponseWriter, r *http.Request) {
	LogDebug.Println("path", "/signin")
	// GET请求返回登录页
	if r.Method == "GET" {
		// 解析需要的模板
		temp, err := template.ParseFiles("views/signin.html", "views/components/navbar.html", "views/components/footer.html", "views/components/header.html")
		if err != nil {
			LogPanic.Panicln("parse template views/index.html failed", err)
		}
		err = temp.ExecuteTemplate(w, "signin.html", nil)
		if err != nil {
			LogPanic.Panicln("parse template views/signin.html failed", err)
		}
		return
	}
	// POST请求，处理登录逻辑
	if r.Method == "POST" {
		r.ParseForm()
		var errorMessage string
		errorMessage = validateSignin(r.Form)
		// 如果登陆过程出错，仍显示登录页，并显示错误信息
		if errorMessage != "" {
			temp, err := template.ParseFiles("views/signin.html", "views/components/navbar.html", "views/components/footer.html", "views/components/header.html")
			if err != nil {
				LogPanic.Panicln("parse template views/index.html failed", err)
			}
			err = temp.ExecuteTemplate(w, "signin.html", errorMessage)
			if err != nil {
				LogPanic.Panicln("parse template views/signin.html failed", err)
			}
			// 如果登陆成功，跳转到首页，并把登陆状态加入到session
		} else {
			sess := session.New()
			// sess.Set("dsa", "dsa")
			// sess.Save()
			// fmt.Println(sess)
			username := r.Form["Username"][0]
			sess = sess.SessionStart(w, r)
			// 登陆成功，自动跳转到首页,设置session记录登录状态
			sess.Set("signin", true)
			sess.Set("username", username)
			sess.Save()
			succeedMessage := fmt.Sprintf("登录成功,%s", username)
			message := MessagePage{Message: succeedMessage, URL: "/"}
			temp, err := template.ParseFiles("views/messagepage.html", "views/components/footer.html")
			if err != nil {
				LogPanic.Panicln("parse all templates failed", err)
			}
			err = temp.ExecuteTemplate(w, "messagepage.html", message)
			if err != nil {
				LogPanic.Panicln("parse template views/messagepage.html failed", err)
			}
			// http.Redirect(w,r,"/",http.StatusFound)
		}
	}
}
func validateSignup(form map[string][]string) (errorMessage string) {
	// LogDebug.Println(form)
	// LogDebug.Println(len(form["Username"]))
	// 用户名为4-20个字符，并且只能由数字字母下划线组成
	if matched, _ := regexp.MatchString(`\w{4,20}`, form["Username"][0]); !matched {
		errorMessage = "用户名为4-20个字符，并且只能由数字字母下划线组成"
		return
	}
	if matched, _ := regexp.MatchString(`^[a-z0-9]+([._\\-]*[a-z0-9])*@([a-z0-9]+[-a-z0-9]*[a-z0-9]+.){1,63}[a-z0-9]+$`, form["Email"][0]); !matched {
		errorMessage = "请输入正确的邮箱地址"
		return
	}
	if matched, _ := regexp.MatchString(`[0-9a-zA-z!@#$%^&*_]{6,20}`, form["Password"][0]); !matched || form["Password"][0] != form["RepeatPassword"][0] {
		errorMessage = "密码格式错误或者重输密码不正确，密码长度需要6-20位，并且只能包含数字、字母，或者!@#$%^&*_这几种特殊字符"
		return
	}
	return
}

// SignupHandle 处理注册页面的逻辑
func SignupHandle(w http.ResponseWriter, r *http.Request) {
	LogDebug.Println("path", "/signup")
	if r.Method == "GET" {
		temp, err := template.ParseFiles("views/signup.html", "views/components/navbar.html", "views/components/footer.html", "views/components/header.html")
		if err != nil {
			LogPanic.Panicln("parse template views/index.html failed", err)
		}
		err = temp.ExecuteTemplate(w, "signup.html", nil)
		if err != nil {
			LogPanic.Panicln("parse template views/signin.html failed", err)
		}
		return
	}
	if r.Method == "POST" {
		// 解析表单数据
		r.ParseForm()
		// var errorMessage string
		// 对表单数据进行验证
		errorMessage := validateSignup(r.Form)
		LogDebug.Println(errorMessage)
		// 如果errorMessage=""，说明验证都能正常通过，我们还要验证用户名是否重复，查询mongodb中的用户表
		if errorMessage == "" && user.CheckUserName(r.Form["Username"][0]) {
			errorMessage = "输入的用户名已经存在"
		}
		if errorMessage != "" {
			temp, err := template.ParseFiles("views/signup.html", "views/components/navbar.html", "views/components/footer.html", "views/components/header.html")
			if err != nil {
				LogPanic.Panicln("parse all templates failed", err)
			}
			err = temp.ExecuteTemplate(w, "signup.html", errorMessage)
			if err != nil {
				LogPanic.Panicln("parse template views/signup.html failed", err)
			}
		} else {
			message := MessagePage{Message: "注册成功", URL: "/signin"}
			err := user.InsertUser(&user.User{Username: r.Form["Username"][0], Password: r.Form["Password"][0], Email: r.Form["Email"][0], ID: bson.NewObjectId()})
			if err != nil {
				log.Panicln("insert user failed", err)
			}
			// 跳转页，自动跳转到登录页
			temp, err := template.ParseFiles("views/messagepage.html", "views/components/footer.html")
			if err != nil {
				LogPanic.Panicln("parse all templates failed", err)
			}
			err = temp.ExecuteTemplate(w, "messagepage.html", message)
			if err != nil {
				LogPanic.Panicln("parse template views/messagepage.html failed", err)
			}
			// http.Redirect(w, r, "/signin", http.StatusFound)
		}
		// r.Form["Username"]
	}
}

// SignoutHandle 退出登录的路由
func SignoutHandle(w http.ResponseWriter, r *http.Request) {
	// 退出登录，那么我们只需要删除session，再进行重定向
	LogDebug.Println("path", "/signout")
	// 先获取session
	sess := session.New()
	sess = sess.SessionStart(w, r)
	// 执行session.Destroy 对应的session数据会在数据库中被删除
	sess.Destroy()
	http.Redirect(w, r, "/", http.StatusFound)
}

// PostHandle 文章页路由
// func PostHandle(w http.ResponseWriter, r *http.Request) {
// 	temp, err := template.ParseFiles("views/post.html", "views/components/navbar.html", "views/components/footer.html",
// 		"views/components/header.html")
// 	if err != nil {
// 		LogPanic.Panicln("parse template views/index.html failed", err)
// 	}
// 	err = temp.ExecuteTemplate(w, "post.html", nil)
// 	if err != nil {
// 		LogPanic.Panicln("parse template views/index.html failed", err)
// 	}
// }

// NewPostHandle 新建一篇文章
func NewPostHandle(w http.ResponseWriter, r *http.Request) {
	// 首先开启session,检查登录状态
	sess := session.New()
	sess = sess.SessionStart(w, r)
	if checkSignin(sess.Data) != true {
		return
	}
	if r.Method == "GET" {
		categoryCotainer := category.Category{}
		namelist, err := categoryCotainer.FindAllCategoryName()
		if err != nil {
			LogPanic.Panicln("get categoriesName failed", err)
		}
		sess.Data["CategoryNames"] = namelist
		temp, err := template.ParseFiles("views/newpost.html", "views/components/navbar.html", "views/components/footer.html",
			"views/components/header.html")
		if err != nil {
			LogPanic.Panicln("parse template views/newpost.html failed", err)
		}
		err = temp.ExecuteTemplate(w, "newpost.html", sess.Data)
		if err != nil {
			LogPanic.Panicln("parse template views/newpost.html failed", err)
		}
		return
	}
	if r.Method == "POST" {
		r.ParseForm()
		Author := r.Form.Get("Author")
		Title := r.Form.Get("Title")
		Content := r.Form.Get("Content")
		Category := r.Form.Get("Category")
		CategoryList := strings.Split(Category, "&")
		newpost := post.New(Author, Title, Content, CategoryList)
		err := newpost.Add()
		var message MessagePage
		if err != nil {
			message = MessagePage{Message: "保存文章失败，请重试", URL: "/post/new"}

		}
		message = MessagePage{Message: "保存文章成功", URL: "/post/new"}
		temp, err := template.ParseFiles("views/messagepage.html", "views/components/footer.html")
		if err != nil {
			LogPanic.Panicln("parse all templates failed", err)
		}
		err = temp.ExecuteTemplate(w, "messagepage.html", message)
		if err != nil {
			LogPanic.Panicln("parse template views/messagepage.html failed", err)
		}

	}
}

// PostIDHandle 根据monggodb的ID字符串来查找文章，返回对应hexid的文章页
func PostIDHandle(w http.ResponseWriter, r *http.Request) {
	urlValue := r.URL.Query()
	idhex := urlValue["id"][0]
	if idhex == "" {
		return
	}
	sess := session.New()
	sess = sess.SessionStart(w, r)
	// 判断是否登录，若没有登录则没有权限，登录后sigin标志会变成true
	// if checkSignin(sess.Data) != true {
	// 	return
	// }
	curpost := &post.Post{}
	err := curpost.FindPostByIDhex(idhex)
	// 如果获取文章出错，跳转到404页面
	if err != nil {
		LogPanic.Panicln("获取文章失败", err)
		http.Redirect(w, r, "/404", http.StatusNotFound)
		return
	}
	sess.Data["post"] = curpost
	temp, err := template.ParseFiles("views/post.html", "views/components/footer.html", "views/components/header.html", "views/components/navbar.html")
	if err != nil {
		LogPanic.Panicln("parse all templates failed", err)
	}
	err = temp.ExecuteTemplate(w, "post.html", sess.Data)
	if err != nil {
		LogPanic.Panicln("parse template views/post.html failed", err)
	}
}
func checkSignin(sessionData map[string]interface{}) bool {
	if sessionData["signin"] == true {
		return true
	}
	return false
}

// PostIDEditHandle 浏览文章的时候可以对指定文章进行编辑操作
func PostIDEditHandle(w http.ResponseWriter, r *http.Request) {
	//1.检查url参数和登录状况
	urlValue := r.URL.Query()
	idhex := urlValue.Get("id")
	if !checkIDhexLen(idhex) {
		fmt.Fprintln(w, "不合法id")
		return
	}
	sess := session.New()
	sess = sess.SessionStart(w, r)
	// 判断是否登录，若没有登录则没有权限，登录后sigin标志会变成true
	if checkSignin(sess.Data) != true {
		return
	}
	if r.Method == "GET" {
		curpost := &post.Post{}
		// 查询对应的文章信息
		err := curpost.FindPostByIDhex(idhex)
		// 如果出错说明没有找到，重定向到404
		if err != nil {
			http.Redirect(w, r, "/404", http.StatusNotFound)
			return
		}
		// 找到后渲染到文章编辑页，编辑页基本上和新建页是一样的
		sess.Data["post"] = curpost
		categoryCotainer := category.Category{}
		namelist, err := categoryCotainer.FindAllCategoryName()
		if err != nil {
			LogPanic.Panicln("get categoriesName failed", err)
		}
		sess.Data["CategoryNames"] = namelist
		temp, err := template.ParseFiles("views/edit.html", "views/components/footer.html", "views/components/header.html", "views/components/navbar.html")
		if err != nil {
			LogPanic.Panicln("parse all templates failed", err)
		}
		err = temp.ExecuteTemplate(w, "edit.html", sess.Data)
		if err != nil {
			LogPanic.Panicln("parse template views/edit.html failed", err)
		}
		return
	}
	if r.Method == "POST" {
		r.ParseForm()
		Author := r.Form.Get("Author")
		Title := r.Form.Get("Title")
		Content := r.Form.Get("Content")
		Category := r.Form.Get("Category")
		CategoryList := strings.Split(Category, "&")
		newpost := post.New(Author, Title, Content, CategoryList)
		newpost.IDhex = idhex
		err := newpost.Update()
		// 如果更新过程出错，说明给的文章id出错，同样重定向到404
		if err != nil {
			LogPanic.Println("更新文章失败", err)
			http.Redirect(w, r, "/404", http.StatusNotFound)
			return
		}
		message := MessagePage{Message: "更新文章成功", URL: "/postid?id=" + idhex}
		showJumpMessage(message, w)
	}
}
func showJumpMessage(message MessagePage, w http.ResponseWriter) {
	temp, err := template.ParseFiles("views/messagepage.html", "views/components/footer.html")
	if err != nil {
		LogPanic.Panicln("parse all templates failed", err)
	}
	err = temp.ExecuteTemplate(w, "messagepage.html", message)
	if err != nil {
		LogPanic.Panicln("parse template views/messagepage.html failed", err)
	}
}

// PostIDDeleteHandle 根据指定id删除对应文章的处理器
func PostIDDeleteHandle(w http.ResponseWriter, r *http.Request) {
	//1.检查url参数和登录状况
	urlValue := r.URL.Query()
	idhex := urlValue.Get("id")
	if !checkIDhexLen(idhex) {
		return
	}
	sess := session.New()
	sess = sess.SessionStart(w, r)
	// 判断是否登录，若没有登录则没有权限，登录后sigin标志会变成true
	if checkSignin(sess.Data) != true {
		return
	}
	if r.Method == "GET" {
		curpost := &post.Post{}
		// 查询对应的文章信息
		err := curpost.DeleteByIDhex(idhex)
		// 如果出错，说明删除失败，
		if err != nil {
			LogPanic.Println("删除文章失败", err)
			http.Redirect(w, r, "/404", http.StatusNotFound)
			return
		}
		message := MessagePage{"删除文章成功", "/"}
		showJumpMessage(message, w)
	}
}

// NewCategoryHandle 处理新建分类页面的路由,首先要检查登录状况
func NewCategoryHandle(w http.ResponseWriter, r *http.Request) {
	sess := session.New()
	sess = sess.SessionStart(w, r)
	if checkSignin(sess.Data) != true {
		return
	}
	if r.Method == "GET" {
		temp, err := template.ParseFiles("views/newcategory.html", "views/components/footer.html", "views/components/header.html", "views/components/navbar.html")
		if err != nil {
			LogPanic.Panicln("parse all templates failed", err)
		}
		err = temp.ExecuteTemplate(w, "newcategory.html", sess.Data)
		if err != nil {
			LogPanic.Panicln("parse template views/newcategory.html failed", err)
		}
	}

	if r.Method == "POST" {
		r.ParseForm()
		// 因为是只有登录了才能新建Category，所以多余的表单信息验证也不需要了。
		LogDebug.Println(r.Form)
		Name := r.Form.Get("Name")
		EnglishName := r.Form.Get("EnglishName")
		Description := r.Form.Get("Description")
		NewCategory := category.New(Name, EnglishName, Description)
		NewCategory.Add()
		showJumpMessage(MessagePage{Message: "添加分类成功", URL: "/category"}, w)
	}
}

// CategoryHandle 处理分类页的路由
func CategoryHandle(w http.ResponseWriter, r *http.Request) {
	sess := session.New()
	sess = sess.SessionStart(w, r)
	if r.Method == "GET" {
		categoryCotainer := category.Category{}
		res, finderr := categoryCotainer.FindAllCategory()
		if finderr != nil {
			LogPanic.Panicln("get all categoryinfo err", finderr)
		}
		sess.Data["CategoryList"] = res
		temp, err := template.ParseFiles("views/category.html", "views/components/footer.html", "views/components/header.html", "views/components/navbar.html")
		if err != nil {
			LogPanic.Panicln("parse all templates failed", err)
		}
		err = temp.ExecuteTemplate(w, "category.html", sess.Data)
		if err != nil {
			LogPanic.Panicln("parse template views/category.html failed", err)
		}
	}
}

func checkIDhexLen(idhex string) bool {
	if len(strings.TrimSpace(idhex)) == 24 {
		return true
	}
	return false
}

// CategoryDeleteHandle 执行删除的路由操作
func CategoryDeleteHandle(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		//检查url参数和登录状况,如果出错获取不到id，会panic没有做其他处理
		urlValue := r.URL.Query()
		idhex := urlValue.Get("id")
		if !checkIDhexLen(idhex) {
			return
		}
		sess := session.New()
		sess = sess.SessionStart(w, r)
		if checkSignin(sess.Data) != true {
			showJumpMessage(MessagePage{Message: "只有管理员用户才能进行该操作", URL: "/category"}, w)
			return
		}
		categoryManage := category.Category{}
		err := categoryManage.FindByIDhexAndDelete(idhex)
		if err != nil {
			showJumpMessage(MessagePage{Message: "删除失败，未知错误", URL: "/category"}, w)
			LogPanic.Panicln("delete category failed", err)
			return
		}
		showJumpMessage(MessagePage{Message: "删除成功", URL: "/category"}, w)
	}

}

// CategoryEditHandle 处理标签编辑操作的路由
func CategoryEditHandle(w http.ResponseWriter, r *http.Request) {
	//检查url参数和登录状况,如果出错获取不到id，会panic没有做其他处理
	urlValue := r.URL.Query()
	idhex := urlValue.Get("id")
	if !checkIDhexLen(idhex) {
		fmt.Fprintf(w, "不合法的id")
		return
	}
	sess := session.New()
	sess = sess.SessionStart(w, r)
	if checkSignin(sess.Data) != true {
		showJumpMessage(MessagePage{Message: "只有管理员用户才能进行该操作", URL: "/category"}, w)
		return
	}
	if r.Method == "GET" {
		categoryManage := category.Category{}
		err := categoryManage.FindOneCategoryInfoByIDhex(idhex)
		// 如果找不到或者查找过程出错
		if err != nil {
			showJumpMessage(MessagePage{Message: "访问目标，未知错误", URL: "/category"}, w)
			return
		}
		sess.Data["CategoryInfo"] = categoryManage
		temp, err := template.ParseFiles("views/editcategory.html", "views/components/footer.html", "views/components/header.html", "views/components/navbar.html")
		if err != nil {
			LogPanic.Panicln("parse all templates failed", err)
		}
		err = temp.ExecuteTemplate(w, "editcategory.html", sess.Data)
		if err != nil {
			LogPanic.Panicln("parse template views/editcategory.html failed", err)
		}
		return
	}
	if r.Method == "POST" {
		r.ParseForm()
		Name := r.Form.Get("Name")
		EnglishName := r.Form.Get("EnglishName")
		Description := r.Form.Get("Description")
		categoryCotainer := category.New(Name, EnglishName, Description)
		err := categoryCotainer.UpdateByIDhex(idhex)
		if err != nil {
			showJumpMessage(MessagePage{Message: "访问目标，未知错误", URL: "/category"}, w)
		}
		showJumpMessage(MessagePage{Message: "更新分类信息成功", URL: "/category"}, w)
	}
}
func main() {
	// 开启静态文件服务，http.StripPrefix提供去掉前缀的静态路由，否则会把路由全部当做路径匹配
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	LogDebug.Println("FileServer start，root ./static")

	// 首页路由
	http.HandleFunc("/", IndexHandle)
	// 登录页路由
	http.HandleFunc("/signin", SigninHandle)
	http.HandleFunc("/signup", SignupHandle)
	http.HandleFunc("/signout", SignoutHandle)
	// 文章页路由
	// http.HandleFunc("/post", PostHandle)
	http.HandleFunc("/post/new", NewPostHandle)           //增加文章
	http.HandleFunc("/postid", PostIDHandle)              //获取显示指定文章
	http.HandleFunc("/postid/edit", PostIDEditHandle)     //编辑指定文章
	http.HandleFunc("/postid/delete", PostIDDeleteHandle) //删除指定文章

	// 文章标签页路由
	http.HandleFunc("/category/new", NewCategoryHandle)       //添加标签页面
	http.HandleFunc("/category", CategoryHandle)              //标签列表页面
	http.HandleFunc("/category/delete", CategoryDeleteHandle) //标签删除操作路由
	http.HandleFunc("/category/edit", CategoryEditHandle)     //标签编辑操作路由

	err := http.ListenAndServe(":3333", nil)
	if err != nil {
		LogFatal.Fatal("ListenAndServe error")
	}

}
