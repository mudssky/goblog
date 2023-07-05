package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"goblog/models/category"
	"goblog/models/post"
	"goblog/models/session"
	"goblog/models/user"

	"github.com/globalsign/mgo/bson"
)

// 日志对象，分别用于输出不同级别的日志
var (
	LogDebug   *log.Logger
	LogWarning *log.Logger
	LogPanic   *log.Logger
	LogFatal   *log.Logger

	// 初始化过程中，一次性解析所有模板
	GlobalTemp *template.Template
)

// MessagePage 存储跳转页渲染信息的结构体
type MessagePage struct {
	Message string
	URL     string
}

// 加载dir 目录下所有结尾为.html的文件，并把相对路径的字符串加入结果的字符串切片。
func loadTemplateDir(dir string) (dirs []string) {
	fileList, err := ioutil.ReadDir(dir)
	if err != nil {
		LogFatal.Fatalln("read template dir error", err)
	}
	for _, v := range fileList {
		if !v.IsDir() {
			if strings.HasSuffix(v.Name(), ".html") {
				dirs = append(dirs, filepath.Join(dir, v.Name()))
			}
			// 如果不是文件是目录，那么递归遍历
		} else {
			newdirs := loadTemplateDir(filepath.Join(dir, v.Name()))
			dirs = append(dirs, newdirs...)
		}
	}
	return
}
func loadConfig() {
	confbytes, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Fatalln("read config file error", err)
	}
	newuser := &user.User{}
	err = json.Unmarshal(confbytes, newuser)
	if err != nil {
		log.Fatalln("json config file error", err)
	}
	log.Println(newuser)
	if user.Exist(newuser) {
		log.Println("user already exist")
	} else {
		newuser.ID = bson.NewObjectId()
		err = user.InsertUser(newuser)
		if err != nil {
			log.Fatalln("user insert error", err)
		}
	}
	log.Println("creat only user succeed")
}
func init() {
	// 创建输出日志文件
	logFile := os.Stdout
	errFile := os.Stderr
	LogDebug = log.New(logFile, "[Debug]:", log.Ldate|log.Ltime|log.Llongfile)
	LogWarning = log.New(logFile, "[Warning]:", log.Ldate|log.Ltime|log.Llongfile)
	LogPanic = log.New(errFile, "[Panic]:", log.Ldate|log.Ltime|log.Llongfile)
	LogFatal = log.New(errFile, "[Fatal]:", log.Ldate|log.Ltime|log.Llongfile)
	loadConfig()
	/*
		// init函数里面用:=赋值 相当于给局部变量复制，所以我们要用=赋值
		var alltemperr error
		AllTemplate, alltemperr = template.ParseFiles("views/index.html", "views/components/navbar.html", "views/components/footer.html","views/components/header.html", "views/signin.html", "views/signup.html")
		if alltemperr != nil {
			LogPanic.Panicln("parse template views/index.html failed")
		}
		LogDebug.Println("parse template succeed:", AllTemplate.DefinedTemplates())
	*/
	templates := loadTemplateDir("views")
	LogDebug.Printf("load templates succeed ,total template count is %v \n%v\n", len(templates), templates)
	var err error
	// GlobalTemp, err = template.ParseFiles("views/index.html", "views/components/navbar.html", "views/components/footer.html",
	// 	"views/components/header.html", "views/404.html", "views/category.html", "views/edit.html", "views/editcategory.html", "views/messagepage.html", "views/newcategory.html", "views/newpost.html",
	// 	"views/post.html", "views/signin.html", "views/signup.html")
	GlobalTemp, err = template.ParseFiles(templates...)
	if err != nil {
		LogFatal.Fatalln("parse all template failed", err)
	}
	GlobalTemp.Funcs(template.FuncMap{"unescaped": unescaped})

}

func unescaped(x string) interface{} { return template.HTML(x) }

// IndexHandle 处理首页的逻辑
func IndexHandle(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		pageNum := 1
		urlValue := r.URL.Query()
		pageStr := urlValue.Get("page")
		if pageStr != "" {
			pageNum, _ = strconv.Atoi(pageStr)
		}
		posts := post.Post{}
		pageNumCount := posts.PageNumCount()
		if pageNum > pageNumCount || pageNum < 1 {
			pageNum = 1
		}
		postindex, err := posts.GetPostsIndexPaged(pageNum)
		// LogDebug.Println(postindex)
		LogDebug.Println("path", "/")
		// 获取session中的数据
		sess := session.New()
		sess = sess.SessionStart(w, r)
		// SessionMap = sess.Data
		// LogDebug.Println(sess.Data["signin"])
		previousNum := pageNum - 1
		if pageNum <= 1 {
			previousNum = 0
		}
		nextNum := pageNum + 1
		if nextNum > pageNumCount {
			nextNum = 0
		}
		sess.Data["postindex"] = postindex
		sess.Data["pageNumList"] = []int{pageNum, pageNum + 1, pageNum + 2}
		sess.Data["pageNumCount"] = pageNumCount
		sess.Data["previousNum"] = previousNum
		sess.Data["nextNum"] = nextNum
		err = GlobalTemp.ExecuteTemplate(w, "index", sess.Data)
		if err != nil {
			LogPanic.Panicln("parse template views/index.html failed", err)
		}
		// http包默认的路由规则，按照最长前缀匹配
		// 所有路径都可以匹配到/,那样404页面就失去作用了，所以当/匹配失败的时候展示404页面
	} else {
		err := GlobalTemp.ExecuteTemplate(w, "404.html", nil)
		if err != nil {
			LogPanic.Panicln("parse template views/404.html failed", err)
		}
	}
}
func validateSignin(form url.Values) (errorMessage string) {
	username := form.Get("Username")
	password := form.Get("Password")
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

		err := GlobalTemp.ExecuteTemplate(w, "signin.html", nil)
		if err != nil {
			LogPanic.Panicln("parse template views/signin.html failed", err)
		}
		return
	}
	err := r.ParseForm()
	if err != nil {
		LogPanic.Panicln("parse form error siginhandle", err)
	}
	// POST请求，处理登录逻辑
	if r.Method == "POST" {
		var errorMessage string
		errorMessage = validateSignin(r.Form)
		// 如果登陆过程出错，仍显示登录页，并显示错误信息
		if errorMessage != "" {
			err := GlobalTemp.ExecuteTemplate(w, "signin.html", errorMessage)
			if err != nil {
				LogPanic.Panicln("parse template views/signin.html failed", err)
			}
			// 如果登陆成功，跳转到首页，并把登陆状态加入到session
		} else {
			sess := session.New()
			// sess.Set("dsa", "dsa")
			// sess.Save()
			// fmt.Println(sess)
			username := r.Form.Get("Username")
			sess = sess.SessionStart(w, r)
			// 登陆成功，自动跳转到首页,设置session记录登录状态
			sess.Set("signin", true)
			sess.Set("username", username)
			err = sess.Save()
			if err != nil {
				LogPanic.Panicln("save session failed signin", err)
			}
			succeedMessage := fmt.Sprintf("登录成功,%s", username)
			message := MessagePage{Message: succeedMessage, URL: "/"}
			err := GlobalTemp.ExecuteTemplate(w, "messagepage.html", message)
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
		err := GlobalTemp.ExecuteTemplate(w, "signup.html", nil)
		if err != nil {
			LogPanic.Panicln("parse template views/signin.html failed", err)
		}
		return
	}
	w.Write([]byte("注册功能暂未开放"))
	return
	/*
		if r.Method == "POST" {
			// 解析表单数据
			r.ParseForm()
			// var errorMessage string
			// 对表单数据进行验证
			errorMessage := validateSignup(r.Form)
			LogDebug.Println(errorMessage)
			// 如果errorMessage=""，说明验证都能正常通过，我们还要验证用户名是否重复，查询mongodb中的用户表
			if errorMessage == "" && user.CheckUserName(r.Form.Get("Username")) {
				errorMessage = "输入的用户名已经存在"
			}
			if errorMessage != "" {
				err := GlobalTemp.ExecuteTemplate(w, "signup.html", errorMessage)
				if err != nil {
					LogPanic.Panicln("parse template views/signup.html failed", err)
				}
			} else {
				message := MessagePage{Message: "注册成功", URL: "/signin"}
				err := user.InsertUser(&user.User{Username: r.Form.Get("Username"), Password: r.Form.Get("Password"), Email: r.Form.Get("Email"), ID: bson.NewObjectId()})
				if err != nil {
					log.Panicln("insert user failed", err)
				}

				err = GlobalTemp.ExecuteTemplate(w, "messagepage.html", message)
				if err != nil {
					LogPanic.Panicln("parse template views/messagepage.html failed", err)
				}
				// http.Redirect(w, r, "/signin", http.StatusFound)
			}
			// r.Form["Username"]
		}*/
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

func getSummary(content string) (summary string) {
	runeContent := []rune(content)
	if len(runeContent) >= 200 {
		summary = string(runeContent[:200])
	} else {
		summary = content
	}
	return
}

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
		err = GlobalTemp.ExecuteTemplate(w, "newpost.html", sess.Data)
		if err != nil {
			LogPanic.Panicln("parse template views/newpost.html failed", err)
		}
		return
	}
	if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			LogPanic.Panicln("parse Form error", err)
		}
		Author := r.Form.Get("Author")
		Title := r.Form.Get("Title")
		Content := r.Form.Get("Content")
		// 转义html标签，防止xss
		// Content = html.EscapeString(Content)
		Category := r.Form.Get("Category")
		CategoryList := strings.Split(Category, "&")
		Summary := getSummary(Content)
		newpost := post.New(Author, Title, Content, CategoryList, Summary)
		err = newpost.Add()
		var message MessagePage
		if err != nil {
			message = MessagePage{Message: "保存文章失败，请重试", URL: "/post/new"}
		}
		// 添加文章成功，增加对应标签的文章计数
		categoryCotainer := category.Category{}
		for _, v := range CategoryList {
			err := categoryCotainer.AddPostsCountsByName(v)
			if err != nil {
				LogWarning.Println("add posts counts failed", err)
			}
		}
		message = MessagePage{Message: "保存文章成功", URL: "/post/new"}
		err = GlobalTemp.ExecuteTemplate(w, "messagepage.html", message)
		if err != nil {
			LogPanic.Panicln("parse template views/messagepage.html failed", err)
		}

	}
}

// PostIDHandle 根据monggodb的ID字符串来查找文章，返回对应hexid的文章页
func PostIDHandle(w http.ResponseWriter, r *http.Request) {
	urlValue := r.URL.Query()
	idhex := urlValue.Get("id")
	if !checkIDhexLen(idhex) {
		fmt.Fprintf(w, "不合法的id")
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
	err = curpost.AddViewsCounts(idhex)
	if err != nil {
		LogWarning.Println("AddViewsCounts error", err)
	}
	sess.Data["post"] = curpost
	err = GlobalTemp.ExecuteTemplate(w, "post.html", sess.Data)
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

		err = GlobalTemp.ExecuteTemplate(w, "edit.html", sess.Data)
		if err != nil {
			LogPanic.Panicln("parse template views/edit.html failed", err)
		}
		return
	}
	if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			LogPanic.Panicln("parse form error postidhandle", err)
		}

		Author := r.Form.Get("Author")
		Title := r.Form.Get("Title")
		Content := r.Form.Get("Content")
		// 不进行转码了，go的template会自动对html字符实体进行转码，所以不需要重复劳动
		// Content = html.EscapeString(Content)
		Category := r.Form.Get("Category")
		CategoryList := strings.Split(Category, "&")
		Summary := getSummary(Content)
		newpost := post.New(Author, Title, Content, CategoryList, Summary)
		newpost.IDhex = idhex
		err = newpost.Update()
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

	err := GlobalTemp.ExecuteTemplate(w, "messagepage.html", message)
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

		err := GlobalTemp.ExecuteTemplate(w, "newcategory.html", sess.Data)
		if err != nil {
			LogPanic.Panicln("parse template views/newcategory.html failed", err)
		}
	}

	if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			LogPanic.Panicln("parse form error", err)
		}
		// 因为是只有登录了才能新建Category，所以多余的表单信息验证也不需要了。
		LogDebug.Println(r.Form)
		Name := r.Form.Get("Name")
		EnglishName := r.Form.Get("EnglishName")
		Description := r.Form.Get("Description")
		NewCategory := category.New(Name, EnglishName, Description)
		err = NewCategory.Add()
		if err != nil {
			LogPanic.Panicln("add category failed", err)
		}
		showJumpMessage(MessagePage{Message: "添加分类成功", URL: "/category"}, w)
	}
}

// CategoryHandle 处理分类页的路由
func CategoryHandle(w http.ResponseWriter, r *http.Request) {
	sess := session.New()
	sess = sess.SessionStart(w, r)
	if r.Method == "GET" {
		categoryCotainer := category.Category{}
		res, err := categoryCotainer.FindAllCategory()
		if err != nil {
			LogPanic.Panicln("get all categoryinfo err", err)
		}
		sess.Data["CategoryList"] = res

		err = GlobalTemp.ExecuteTemplate(w, "category.html", sess.Data)
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

		err = GlobalTemp.ExecuteTemplate(w, "editcategory.html", sess.Data)
		if err != nil {
			LogPanic.Panicln("parse template views/editcategory.html failed", err)
		}
		return
	}
	if r.Method == "POST" {
		err := r.ParseForm()
		panicErr("parse form error", err)
		Name := r.Form.Get("Name")
		EnglishName := r.Form.Get("EnglishName")
		Description := r.Form.Get("Description")
		categoryCotainer := category.New(Name, EnglishName, Description)
		err = categoryCotainer.UpdateByIDhex(idhex)
		if err != nil {
			showJumpMessage(MessagePage{Message: "访问目标，未知错误", URL: "/category"}, w)
		}
		showJumpMessage(MessagePage{Message: "更新分类信息成功", URL: "/category"}, w)
	}
}
func panicErr(message string, err error) {
	LogPanic.Panicln(message, err)
}

// UploadHandle 处理图片的上传操作
func UploadHandle(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		return
	}
	sess := session.New()
	sess = sess.SessionStart(w, r)
	if checkSignin(sess.Data) != true {
		w.Write([]byte("只有管理员用户才能进行该操作"))
		return
	}
	// LogDebug.Println(r)
	// urlValue := r.URL.Query()
	picHashFileName := r.Header.Get("picHashFileName")

	picbytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		LogWarning.Println("read upload body failed")
	}
	filename := "static/assets/" + picHashFileName
	LogDebug.Println(filename)
	err = ioutil.WriteFile(filename, picbytes, os.FileMode(0733))
	if err != nil {
		LogWarning.Println("write file filed:", err)
		w.Write([]byte("upload failed"))
	}
	w.Write([]byte("upload succeed"))
}

// MyStripPrefixAndCheck 修改去前缀的方法，检查目录，不存在的文件返回404
func MyStripPrefixAndCheck(prefix string, h http.Handler) http.Handler {
	if prefix == "" {
		return h
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//log.Println("enter",r.URL.Path)
		f, err := os.Stat("." + r.URL.Path)
		//判断目录或者文件是否存在，如果存在，不存在，返回404。再判断是否是目录,是目录直接返回404
		if err != nil {
			log.Println(r.URL.Path)
			http.NotFound(w, r)
			return
		}
		if f.IsDir() {
			log.Println(r.URL.Path)
			http.NotFound(w, r)
			return
		}
		if p := strings.TrimPrefix(r.URL.Path, prefix); len(p) < len(r.URL.Path) {
			//log.Println("excuted",r.URL.Path)
			r2 := new(http.Request)
			*r2 = *r
			r2.URL = new(url.URL)
			*r2.URL = *r.URL
			r2.URL.Path = p
			h.ServeHTTP(w, r2)
		} else {
			http.NotFound(w, r)
		}
	})
}

// SearchHandle 处理搜索功能的路由
func SearchHandle(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("搜索功能暂未实装"))

}

//go:generate go env -w GO111MODULE=on
//go:generate go env -w GOPROXY=https://goproxy.cn,direct
//go:generate go mod tidy
//go:generate go mod download
func main() {
	// 开启静态文件服务，http.StripPrefix提供去掉前缀的静态路由，否则会把路由全部当做路径匹配
	http.Handle("/static/", MyStripPrefixAndCheck("/static/", http.FileServer(http.Dir("./static"))))
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
	// 搜索框路由
	http.HandleFunc("/search", SearchHandle)

	// 图片上传路由
	http.HandleFunc("/upload", UploadHandle)
	err := http.ListenAndServe(":3333", nil)
	if err != nil {
		LogFatal.Fatal("ListenAndServe error")
	}

}
