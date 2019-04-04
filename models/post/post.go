package main

import (
	"log"
	"time"

	"github.com/globalsign/mgo/bson"

	"github.com/globalsign/mgo"
)

// Post 文章的结构体
type Post struct {
	ID           bson.ObjectId       `bson:"_id"`
	Autuor       string              `bson:"author"`
	Title        string              `bson:"title"`
	Content      string              `bson:"content"`
	CreatAt      bson.MongoTimestamp `bson:"creatat"`
	LastModified bson.MongoTimestamp `bson:"lastmodified"`
	Category     string              `bson:"category"`
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
	collection = database.C("post")
}

// GetDatabase 获取数据库对象
func GetDatabase() *mgo.Database {
	return database
}

// GetCollection 获取集合对象
func GetCollection() *mgo.Collection {
	return collection
}
func main() {

}
