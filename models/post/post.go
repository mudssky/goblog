package post

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/qiniu/qmgo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Post 文章的结构体
type Post struct {
	ID      primitive.ObjectID `bson:"_id"`
	IDhex   string             `bson:"idhex"`
	Author  string             `bson:"author"`
	Title   string             `bson:"title"`
	Content string             `bson:"content"`
	CreatAt time.Time          `bson:"creatat"`
	// CreatAt       int64         `bson:"creatat"`
	// LastModified  int64         `bson:"lastmodified"`
	LastModified  time.Time `bson:"lastmodified"`
	CategoryList  []string  `bson:"categorylist"`
	ViewsCounts   int       `bson:"viewscounts"`
	CommentCounts int       `bson:"commentcounts"`
	Summary       string    `bson:"summary"`
}

const (
	pageCount int64 = 10 //分页默认的每页项目数
)

var (
	ctx        context.Context
	collection *qmgo.QmgoClient
)

func init() {

	ctx = context.Background()
	var err error
	collection, err = qmgo.Open(ctx, &qmgo.Config{Uri: "mongodb://localhost:27017", Database: "goblog", Coll: "post"})
	if err != nil {
		log.Fatalln("connect error", err)
	}
}

// Init 初始化一个Post结构体，根据提供的参数
func (p *Post) Init(Author string, Title string, Content string, CategoryList []string, Summary string) *Post {
	p.ID = primitive.NewObjectID()
	p.IDhex = p.ID.Hex()
	p.Author = Author
	p.Title = Title
	p.Content = Content
	p.CategoryList = CategoryList
	p.CreatAt = time.Now()
	// p.CreatAt = time.Now().UnixNano()
	p.LastModified = p.CreatAt
	p.ViewsCounts = 0
	p.CommentCounts = 0
	p.Summary = Summary
	return p
}

// New 返回一个初始化的POST结构体，根据提供的参数
func New(Author string, Title string, Content string, CategoryList []string, Summary string) *Post {
	return new(Post).Init(Author, Title, Content, CategoryList, Summary)
}

// Add 添加一篇文章的数据到数据库,如果插入过程中出错，会返回错误对象
func (p *Post) Add() error {
	c := GetCollection()
	_, err := c.InsertOne(ctx, p)
	return err
}

// FindPostByIDhex 通过ID来查找文章数据并返回
func (p *Post) FindPostByIDhex(objectidhex string) (err error) {
	// objid := primitive.NewObjectID(objectidhex).Hex()
	c := GetCollection()
	err = c.Find(ctx, qmgo.M{
		"idhex": objectidhex,
	}).One(p)
	return
}

// AddViewsCounts 增加对应文章的浏览数 +1
func (p *Post) AddViewsCounts(objectidhex string) (err error) {
	// objid := bson.ObjectIdHex(objectidhex)
	c := GetCollection()

	err = c.UpdateId(ctx, objectidhex, bson.M{"$inc": bson.M{"viewscounts": 1}})
	return
}

// GetPostsIndex 获取首页需要用到的文章信息
func (p *Post) GetPostsIndex() (res []Post, err error) {
	c := GetCollection()
	err = c.Find(ctx, qmgo.M{}).Select(bson.M{"idhex": 1, "title": 1, "lastmodified": 1, "author": 1, "viewscounts": 1, "summary": 1}).All(&res)
	return
}

// GetPostsIndexDesc 获取首页需要用到的文章信息,按时间倒序排列
func (p *Post) GetPostsIndexDesc() (res []interface{}, err error) {
	c := GetCollection()
	err = c.Find(ctx, qmgo.M{}).Select(bson.M{"idhex": 1, "title": 1, "lastmodified": 1, "author": 1, "viewscounts": 1, "summary": 1}).Sort("-lastmodified").All(&res)
	return
}

// PostCount  返回总文档数
func (p *Post) PostCount() (n int64, err error) {
	c := GetCollection()
	n, err = c.Collection.Find(ctx, qmgo.M{}).Count()
	return
}

// PageNumCount  返回总页数
func (p *Post) PageNumCount() int64 {
	n, err := p.PostCount()
	// 错误处理，如果出错，只返回一页
	if err != nil {
		n = 1
	}

	// 如果是整页数，那么页数就是总文档数/每页文档数，如果不是整页数，页数需要+1
	if n%pageCount == 0 {
		n /= pageCount
	} else {
		n /= pageCount
		n++
	}

	return n
}

// GetPostsIndexPaged 分页，很明显需要每页显示的项目数，作为一个api，只需要输入页数返回对应的内容即可。
// 我们这里设置固定每页的文章数目 pageCount为10。除此之外我们每次要先计算总页数，也要写一个函数
func (p *Post) GetPostsIndexPaged(pageNum int64) (res []Post, err error) {
	c := GetCollection()
	err = c.Find(ctx, qmgo.M{}).Select(qmgo.M{"idhex": 1, "title": 1, "lastmodified": 1, "author": 1, "viewscounts": 1, "summary": 1}).Sort("-lastmodified").Skip(pageNum - 1).Limit(pageCount).All(&res)
	fmt.Println("index post", res)
	return
}

// Update 更新数据库中的文章信息，需要修改文章结构体后再调用，根据文章ID进行更新
func (p *Post) Update() error {

	c := GetCollection()
	err := c.UpdateOne(ctx, bson.M{"_id": p.ID}, bson.M{"$set": bson.M{"lastmodified": p.LastModified, "author": p.Author, "categorylist": p.CategoryList, "content": p.Content, "title": p.Title, "summary": p.Summary}})
	return err
}

// DeleteByID 通过ID删除文章数据
func (p *Post) DeleteByID(id primitive.ObjectID) error {
	c := GetCollection()
	err := c.RemoveId(ctx, id)
	return err
}

// DeleteByIDhex 通过IDhex 字符串删除文章数据
func (p *Post) DeleteByIDhex(idhex string) (err error) {
	p.ID, err = primitive.ObjectIDFromHex(idhex)
	if err != nil {
		panic(err)
	}
	c := GetCollection()
	err = c.RemoveId(ctx, p.ID)
	return err
}

// GetCollection 获取集合对象
func GetCollection() *qmgo.QmgoClient {
	return collection
}
