package user

import (
	"context"
	"log"

	"github.com/qiniu/qmgo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User 用户表的一行信息 包括用户名 密码
// 暂时先使用明文存储密码
type User struct {
	ID           primitive.ObjectID `bson:"_id"`
	Username     string             `bson:"username"`
	Password     string             `bson:"password"`
	PasswordHash string             `bson:"passwordhash,omitempty"`
	Nickname     string             `bson:"nickname,omitempty"`
	Email        string             `bson:"email"`
}

var (
	ctx        context.Context
	collection *qmgo.QmgoClient
)

func init() {
	ctx = context.Background()
	var err error
	collection, err = qmgo.Open(ctx, &qmgo.Config{Uri: "mongodb://localhost:27017", Database: "goblog", Coll: "user"})
	if err != nil {
		log.Fatalln("connect error", err)
	}
}

// CheckUserName 检查用户名是否存在
func CheckUserName(username string) bool {
	collection := GetCollection()
	res := User{}

	err := collection.Find(ctx, qmgo.M{"username": username}).One(&res)
	if err != nil {
		if qmgo.IsErrNoDocuments(err) {
			return false
		} else {
			panic(err)
		}
	}
	return true
}

// CheckLogin 检查对应用户名和密码在数据库中能否找到，如果能找到，返回true。
func CheckLogin(username string, password string) bool {
	res := User{}
	err := collection.Find(ctx, bson.M{"username": username, "password": password}).One(&res)
	// 没有mgo.ErrNotFound错误，说明找到了。其他错误说明没找到。

	if err != nil {
		if qmgo.IsErrNoDocuments(err) {
			return false
		} else {
			panic(err)
		}
	}
	return true
}

// InsertUser 添加新用户到数据库
func InsertUser(user *User) (err error) {
	collection := GetCollection()
	_, err = collection.InsertOne(ctx, user)
	return
}

// GetCollection 获取集合对象
func GetCollection() *qmgo.QmgoClient {
	return collection
}

// Exist 判断用户是否存在
func Exist(user *User) bool {
	collection := GetCollection()
	var res interface{}
	err := collection.Find(ctx, bson.M{"username": user.Username}).One(&res)
	if err != nil {
		if qmgo.IsErrNoDocuments(err) {
			return false
		} else {
			panic(err)
		}
	}
	return true

}
