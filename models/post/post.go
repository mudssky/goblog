package post

import (
	"log"
	"time"

	"github.com/globalsign/mgo/bson"

	"github.com/globalsign/mgo"
)

// Post 文章的结构体
type Post struct {
	ID      bson.ObjectId `bson:"_id"`
	IDhex   string        `bson:"idhex"`
	Author  string        `bson:"author"`
	Title   string        `bson:"title"`
	Content string        `bson:"content"`
	CreatAt int64         `bson:"creatat"`
	// CreatAt       int64         `bson:"creatat"`
	// LastModified  int64         `bson:"lastmodified"`
	LastModified  int64         `bson:"lastmodified"`
	CategoryList  []interface{} `bson:"categorylist"`
	ViewsCounts   int           `bson:"viewscounts"`
	CommentCounts int           `bson:"commentCounts"`
}

var (
	session    *mgo.Session
	database   *mgo.Database
	collection *mgo.Collection
)

// Init 初始化一个Post结构体，根据提供的参数
func (p *Post) Init(Author string, Title string, Content string, Category string) *Post {
	p.ID = bson.NewObjectId()
	p.IDhex = p.ID.Hex()
	p.Author = Author
	p.Title = Title
	p.Content = Content
	p.CategoryList = []interface{}{}
	p.CreatAt = time.Now().UnixNano()
	p.LastModified = p.CreatAt
	p.ViewsCounts = 0
	p.CommentCounts = 0
	return p
}

// New 返回一个初始化的POST结构体，根据提供的参数
func New(Author string, Title string, Content string, Category string) *Post {
	return new(Post).Init(Author, Title, Content, Category)
}

// Add 添加一篇文章的数据到数据库,如果插入过程中出错，会返回错误对象
func (p *Post) Add() error {
	c := GetCollection()
	err := c.Insert(p)
	return err
}

// FindPostByIDhex 通过ID来查找文章数据并返回
func (p *Post) FindPostByIDhex(objectidhex string) (err error) {
	objid := bson.ObjectIdHex(objectidhex)
	c := GetCollection()
	err = c.FindId(objid).One(p)
	return
}

// GetPostsIndex 获取首页需要用到的文章信息
func (p *Post) GetPostsIndex() (res []interface{}, err error) {
	c := GetCollection()
	err = c.Find(bson.M{}).Select(bson.M{"idhex": 1, "title": 1, "lastmodified": 1, "author": 1, "viewscounts": 1, "commentscounts": 1}).All(&res)
	return
}

// Update 更新数据库中的文章信息，需要修改文章结构体后再调用，根据文章ID进行更新
func (p *Post) Update() error {
	p.ID = bson.ObjectIdHex(p.IDhex)
	c := GetCollection()
	err := c.Update(bson.M{"_id": p.ID}, bson.M{"$set": bson.M{"lastmodified": p.LastModified, "author": p.Author, "categorylist": p.CategoryList, "content": p.Content, "title": p.Title}})
	return err
}

// DeleteByID 通过ID删除文章数据
func (p *Post) DeleteByID(id bson.ObjectId) error {
	c := GetCollection()
	err := c.RemoveId(id)
	return err
}

// DeleteByIDhex 通过IDhex 字符串删除文章数据
func (p *Post) DeleteByIDhex(idhex string) error {
	p.ID = bson.ObjectIdHex(idhex)
	c := GetCollection()
	err := c.RemoveId(p.ID)
	return err
}
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

/*
func main() {

}
*/
