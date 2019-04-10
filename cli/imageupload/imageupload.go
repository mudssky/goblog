package main

import (
	"bytes"
	"crypto/md5"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	prefix   string //图片地址前缀，即图片文件名前面的部分地址
	input    string //处理文件的全名，包括路径，可以是相对路径和绝对路径
	output   string //输出目录
	target   string //上传目标的url地址
	password string // 密码，服务端要验证密码才能上传
	cookie   string // 另一种验证方式是验证cookie
	picPath  string //上传图片的路径
)

// 在初始化函数里面，注册并解析参数
func init() {
	flag.StringVar(&prefix, "prefix", "/static/assets/", "图片地址前缀，即md文件中图片文件名前面的部分地址")
	// flag.StringVar(&prefix, "prefix", "/static/assets/", "图片地址前缀，即md文件中图片文件名前面的部分地址")
	flag.StringVar(&input, "input", "", "处理文件的全名，包括路径，可以是相对路径和绝对路径")
	flag.StringVar(&output, "output", "./output", "处理后的文件和图片输出路径")
	flag.StringVar(&target, "target", "http://localhost:3333/upload", "上传目标的url地址")
	flag.StringVar(&password, "password", "imageupload", "密码，服务端要验证密码才能上传")
	flag.StringVar(&cookie, "cookie", "Pycharm-24d5082b=846db6bd-86be-4279-aa95-a3e4f79c9fe3; __utmz=111872281.1552395922.1.1.utmcsr=(direct)|utmccn=(direct)|utmcmd=(none); __utma=111872281.1528707869.1552395922.1552395922.1552398878.2; mod=mpaKwt9kQCGE2eS_zAHMo6IScWnZtoHrowtnSYVk1gvK-QiQUvUrP37EhrM8SPnc1wVIUQx8l25XzD6HrPYcAA%3D%3D", "另一种验证方式是验证cookie")
	flag.StringVar(&picPath, "picPath", "test.png", "只上传图片时图片的路径")
	flag.Parse()
}

// 转换浏览器上截取的字符串为cookie列表
func convertStrCookie(cookieStr string) (res []*http.Cookie) {
	cookieList := strings.Split(cookieStr, ";")
	for _, v := range cookieList {
		keyValue := strings.Split(v, "=")
		res = append(res, &http.Cookie{Name: keyValue[0], Value: keyValue[1]})
	}
	return
}

func uploadPicture(picbytes []byte, picHashFileName string, target string, cookie string) error {
	req, err := http.NewRequest("POST", target, bytes.NewReader(picbytes))
	cookieList := convertStrCookie(cookie)
	// 把列表中的cookie都加到请求上
	for _, v := range cookieList {
		req.AddCookie(v)
	}
	// http协议传参有3个地方，一个是url，一个是header，另一个是请求体。这里请求体被图片占用，url固定，所以姑且用header传参数
	req.Header.Add("picHashFileName", picHashFileName)
	// urlValue := req.URL.Query()
	// urlValue.Add("picHashName", picHashName)
	// log.Println(req.URL)
	client := &http.Client{}
	res, err := client.Do(req)
	log.Println("上传成功", res)
	return err
}

// Md5FromBytes 输入[]byte类型，求出base16的md5值并返回
func Md5FromBytes(picbytes []byte) string {
	md5bytes := md5.Sum(picbytes)
	// log.Println(hex.EncodeToString(md5bytes))
	// md5bytes需要用base16转码才能得到想要的结果
	picHashName := fmt.Sprintf("%x", md5bytes)
	return picHashName
}

/*func main() {
	picbytes, err := ioutil.ReadFile(picPath)
	log.Println("read pic complete:", picPath)
	if err != nil {
		log.Panicln("read pic error:", err)
	}
	picExt := filepath.Ext(picPath)
	picHashName := Md5FromBytes(picbytes)
	log.Println("get md5 succeed", picHashName+picExt)
	err = uploadPicture(picbytes, picHashName+picExt, target, cookie)
	if err != nil {
		log.Println(err)
	}
}*/

func main() {
	// 图片的地址是这样的
	// ![1554601894370](assets/1554601894370.png)

	// 首先要读入markdown文件
	markdownPath := filepath.Clean(input)
	bytes, err := ioutil.ReadFile(markdownPath)
	if err != nil {
		log.Panicln("read markdown file error", err)
	}
	log.Println("read markdwon completed:", markdownPath)

	re := regexp.MustCompile(`!\[.*?\]\((.*?)\)`)
	//第二个参数，表示最大匹配的个数 n>=0会返回 最多前n个匹配结果，-1就是返回所有结果
	// strSlice := re.FindAllString(string(bytes), -1)
	// log.Println(strSlice)
	// indexSlice := re.FindAllIndex(bytes, -1)

	// 找到所有子匹配索引，返回n个长度为4的切片，表示匹配中分组的部分，和分组之外的部分
	indexSlice := re.FindAllSubmatchIndex(bytes, -1)
	log.Println("regexp match completed:", indexSlice)

	// 用一个新的byte数组存放处理的结果
	newbytes := []byte{}
	// 存放上次处理的匹配部分的末尾
	lastEnd := 0
	for _, v := range indexSlice {
		// bytes[v[0]:v[1]] 是字符串的范围
		// log.Println(string(bytes[v[0]:v[1]]))
		// 首先把不需要处理的部分数据加入newbytes
		// newbytes = append(newbytes, bytes[lastEnd:v[0]]...)
		// 匹配到的部分字符串需要进行处理
		// 这个是匹配到分组的部分
		// log.Println(string(bytes[v[2]:v[3]]))
		newbytes = append(newbytes, bytes[lastEnd:v[2]]...)
		// bytes[v[2]:v[3]] 是匹配到图片的相对地址
		picPath := filepath.Clean(string(bytes[v[2]:v[3]]))
		// log.Println(picPath, len(picPath))
		log.Println("read pic :", picPath)
		picbytes, err := ioutil.ReadFile(picPath)
		if err != nil {
			log.Println("read pic error", err)
		}
		// log.Println(len(picbytes))
		// fmt.Println(picbytes)

		picHashName := Md5FromBytes(picbytes)
		// 用这个名字替换原来图片的名字
		// 首先获取原来图片名的后缀
		picExt := filepath.Ext(picPath)
		picHashFileName := picHashName + picExt
		newPicPath := filepath.Join(output, "assets", picHashFileName)
		cleanOutput := filepath.Clean(output)
		// _, err = os.Stat(cleanOutput)
		// if os.IsNotExist(err) {}
		err = os.MkdirAll(filepath.Join(cleanOutput, "assets"), os.FileMode(0733))
		if err != nil {
			// log.Println("mkdir error", err)
			log.Panicln("mkdir error", err)
		}
		log.Println("make output directory complete.")
		// 图片拷贝到新目录
		err = ioutil.WriteFile(newPicPath, picbytes, os.FileMode(0733))
		if err != nil {
			log.Panicln("write pic error", err)
		}
		log.Println("new pic gen completed:", newPicPath)

		// 只有当目标网址的参数被解析的时候才进行上传

		uploadPicture(picbytes, picHashFileName, target, cookie)

		// 接下来，要把markdown里面的图片路径进行替换
		prePicPath := prefix + picHashName + picExt
		newbytes = append(newbytes, prePicPath...)
		lastEnd = v[3]
	}
	// 最后还有内容没加入newbytes
	newbytes = append(newbytes, bytes[lastEnd:]...)

	// 替换完成后newbytes已经是新的markdown文档，我们把他拷贝到输出目录
	markdownFileName := filepath.Base(markdownPath)
	err = ioutil.WriteFile(filepath.Join(output, markdownFileName), newbytes, os.FileMode(0777))
	if err != nil {
		log.Println("write markdown error", err)
	}
	log.Println("write new markdown completed.")
	log.Println("done")
}
