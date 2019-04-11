package user

import (
	"log"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// User 用户表的一行信息 包括用户名 密码
// 暂时先使用明文存储密码
type User struct {
	ID           bson.ObjectId `bson:"_id"`
	Username     string        `bson:"username"`
	Password     string        `bson:"password"`
	PasswordHash string        `bson:"passwordhash,omitempty"`
	Nickname     string        `bson:"nickname,omitempty"`
	Email        string        `bson:"email"`
}

var (
	session    *mgo.Session
	database   *mgo.Database
	collection *mgo.Collection
)

func init() {
	dialInfo := &mgo.DialInfo{
		Addrs:     []string{"127.0.0.1:27017"},
		Timeout:   time.Second * 1,
		PoolLimit: 4096,
	}
	// 创建一个维护套接字池的session
	session, err := mgo.DialWithInfo(dialInfo)
	if err != nil {
		log.Panicln("dial database failed", err)
	}
	database = session.DB("goblog")
	collection = database.C("user")
}

// CheckUserName 检查用户名是否存在
func CheckUserName(username string) bool {
	collection := GetCollection()
	res := User{}
	err := collection.Find(bson.M{"username": username}).One(&res)
	// log.Println(res)
	if err != mgo.ErrNotFound {
		return true
	}
	return false
}

// CheckLogin 检查对应用户名和密码在数据库中能否找到，如果能找到，返回true。
func CheckLogin(username string, password string) bool {
	res := User{}
	err := collection.Find(bson.M{"username": username, "password": password}).One(&res)
	// 没有mgo.ErrNotFound错误，说明找到了。其他错误说明没找到。
	if err != mgo.ErrNotFound {
		return true
	}
	return false
}

// InsertUser 添加新用户到数据库
func InsertUser(user *User) (err error) {
	collection := GetCollection()
	err = collection.Insert(user)
	return
}

// GetDatabase 获取数据库对象
func GetDatabase() *mgo.Database {
	return database
}

// GetCollection 获取集合对象
func GetCollection() *mgo.Collection {
	return collection
}

// Exist 判断用户是否存在
func Exist(user *User) bool {
	collection := GetCollection()
	var res interface{}
	err := collection.Find(bson.M{"username": user.Username}).One(&res)
	if err != mgo.ErrNotFound {
		return true
	}
	return false
}

/*func main() {
	// Dial可以设置多个服务器，这里就一个
	session, err := mgo.Dial("127.0.0.1:27017")
	if err != nil {
		panic(err)
	}
	defer session.Close()
	//设置连接模式，有三种连接模式，默认是strong，即只连接主服务器，一直使用一个连接，因此所有的读写操作会完全的一致
	// 另外两个模式是多个服务器情况下使用的
	// Monotonic 这个模式读取的不一定是最新的数据，首先向其他服务器发起连接，只要出现了一次写操作，session的连接就会切换到主服务器。
	// Eventual 这个模式的读操作会向任意其他服务器发起，多次读操作并不一定使用相同的连接，也就是读操作不一定有序。
	// session的写操作总是向主服务器发起，但是可能使用不同的连接，也就是写操作也不一定有序
	// session.SetMode(mgo.Monotonic)
	// 因为我们用默认的就行了，所以不用进行设置

	c := session.DB("goblog").C("user")
	err = c.Insert(&User{ID: bson.NewObjectId(), Username: "abc3", Password: "123456"})
	if err != nil {
		log.Fatal(err)
	}
	result := User{}
	err = c.Find(bson.M{"username": "abc3"}).One(&result)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(result)
}
*/
