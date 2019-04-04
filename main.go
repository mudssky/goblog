package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"regexp"

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
		err = temp.ExecuteTemplate(w, "index", sess.Data)
		if err != nil {
			LogPanic.Panicln("parse template views/index.html failed")
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
			LogPanic.Panicln("parse template views/404.html failed")
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
			LogPanic.Panicln("parse template views/signin.html failed")
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
				LogPanic.Panicln("parse template views/signin.html failed")
			}
			// 如果登陆成功，跳转到首页，并把登陆状态加入到session
		} else {
			sess := session.New()
			// sess.Set("dsa", "dsa")
			// sess.Save()
			// fmt.Println(sess)
			sess = sess.SessionStart(w, r)
			// 登陆成功，自动跳转到首页,设置session记录登录状态
			sess.Set("signin", true)
			sess.Save()
			message := MessagePage{Message: "登录成功", URL: "/"}
			temp, err := template.ParseFiles("views/messagepage.html", "views/components/footer.html")
			if err != nil {
				LogPanic.Panicln("parse all templates failed", err)
			}
			err = temp.ExecuteTemplate(w, "messagepage.html", message)
			if err != nil {
				LogPanic.Panicln("parse template views/messagepage.html failed")
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
			LogPanic.Panicln("parse template views/signin.html failed")
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
				LogPanic.Panicln("parse template views/signup.html failed")
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
				LogPanic.Panicln("parse template views/messagepage.html failed")
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
func PostHandle(w http.ResponseWriter, r *http.Request) {
	temp, err := template.ParseFiles("views/post.html", "views/components/navbar.html", "views/components/footer.html",
		"views/components/header.html")
	if err != nil {
		LogPanic.Panicln("parse template views/index.html failed", err)
	}
	err = temp.ExecuteTemplate(w, "post.html", nil)
	if err != nil {
		LogPanic.Panicln("parse template views/index.html failed")
	}
}

// NewPostHandle 新建一篇文章
func NewPostHandle(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		temp, err := template.ParseFiles("views/newpost.html", "views/components/navbar.html", "views/components/footer.html",
			"views/components/header.html")
		if err != nil {
			LogPanic.Panicln("parse template views/newpost.html failed", err)
		}
		err = temp.ExecuteTemplate(w, "newpost.html", nil)
		if err != nil {
			LogPanic.Panicln("parse template views/newposy.html failed")
		}
		return
	}
	if r.Method == "POST" {
		r.ParseForm()

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
	http.HandleFunc("/post", PostHandle)
	http.HandleFunc("/post/new", NewPostHandle)
	err := http.ListenAndServe(":3333", nil)
	if err != nil {
		LogFatal.Fatal("ListenAndServe error")
	}

}
