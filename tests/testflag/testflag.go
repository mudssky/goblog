package main

import (
	"flag"
	"fmt"
)

func main() {
	/*-flag
	-flag=x
	-flag x  // 只有非bool类型的flag可以
	命令行有以上三种指明参数的方法
	一个 -  和两个 -- 连接符是一样的效果，对应的参数必须有连接符。
	*/

	namePtr := flag.String("name", "admin", "名字")
	agePtr := flag.Int("age", 18, "年龄")
	married := flag.Bool("married", false, "是否结婚了")

	// 可以事先定义变量，引用赋值
	var email string
	flag.StringVar(&email, "email", "none", "邮箱")

	flag.Parse()
	// Args方法用于获取 non-flag参数
	args := flag.Args()
	// NArgs 方法用于获取non-flag参数的个数
	nargs := flag.NArg()
	// NFlag 活气已经设置了的参数个数
	nflag := flag.NFlag()

	fmt.Println("name:", *namePtr)
	fmt.Println("age:", *agePtr)
	fmt.Println("married:", *married)
	fmt.Println("email:", email)
	fmt.Println("args:", args)
	fmt.Println("nargs:", nargs)
	fmt.Println("nflag:", nflag)
}

/*
 PrintDefaults() 打印所有已定义参数默认值
Flag结构体
 type Flag struct {
    Name     string // flag在命令行中的名字
    Usage    string // 帮助信息
    Value    Value  // 要设置的值
    DefValue string // 默认值（文本格式），用于使用信息
}

func (f *FlagSet) Lookup(name string) *Flag 返回已经注册的flag指针结构体

func (f *FlagSet) Visit(fn func(*Flag)) 遍历解析时设置了的标签
func (f *FlagSet) VisitAll(fn func(*Flag)) 遍历所有
*/
