package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
)

// 日志对象，分别用于输出不同级别的日志
var (
	LogDebug   *log.Logger
	LogWarning *log.Logger
	LogPanic   *log.Logger
	LogFatal   *log.Logger
)

type Info struct {
	Name  string
	Email string
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

}
func indexHandle(w http.ResponseWriter, r *http.Request) {
	temp, err := template.ParseFiles("views/index.html", "views/components/navbar.html", "views/components/footer.html")

	if err != nil {
		LogPanic.Panicln("parse template views/index.html failed")
	}
	err = temp.ExecuteTemplate(w, "index", nil)
	if err != nil {
		LogPanic.Panicln("parse template views/index.html failed")
	}

}

func main() {
	// 开启静态文件服务，http.StripPrefix提供去掉前缀的静态路由，否则会把路由全部当做路径匹配
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	LogDebug.Println("FileServer start，root ./static")

	// 首页路由
	http.HandleFunc("/", indexHandle)
	err := http.ListenAndServe(":3333", nil)
	if err != nil {
		LogFatal.Fatal("ListenAndServe error")
	}

}
