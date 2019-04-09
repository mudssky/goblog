package category

import (
	"log"
	"time"

	"github.com/globalsign/mgo/bson"

	"github.com/globalsign/mgo"
)

// Category  记录分类标签信息的结构体
type Category struct {
	ID          bson.ObjectId `bson:"_id"`
	IDhex       string        `bson:"idhex"`
	Name        string        `bson:"name"`
	EnglishName string        `bson:"englishname"`
	PostsCounts int           `bson:"postscounts"`
	Description string        `bson:"description"`
}

var (
	session    *mgo.Session
	database   *mgo.Database
	collection *mgo.Collection
)

// Init Category结构体的初始化方法，给定3个对应值进行初始化
func (c *Category) Init(Name string, EnglishName string, Description string) *Category {
	c.ID = bson.NewObjectId()
	c.IDhex = c.ID.Hex()
	c.Name = Name
	c.EnglishName = EnglishName
	c.PostsCounts = 0
	c.Description = Description
	return c
}

// New 给定初始值，返回一个初始化好的Category对象
func New(Name string, EnglishName string, Desp string) *Category {
	return new(Category).Init(Name, EnglishName, Desp)
}

// Add 添加一个category到数据库，如果插入过程出错，将返回一个error对象
func (c *Category) Add() error {
	collection := GetCollection()
	err := collection.Insert(c)
	return err
}

// FindAllCategory  从数据库中获取所有类别信息
func (c *Category) FindAllCategory() (res []interface{}, err error) {
	collection := GetCollection()
	err = collection.Find(bson.M{}).Select(bson.M{"name": 1, "englishname": 1, "idhex": 1, "description": 1}).All(&res)
	return
}

// AddPostsCountsByName 对对应目录的Count数进行+1操作，如果失败，返回err
func (c *Category) AddPostsCountsByName(Name string) error {
	collection = GetCollection()
	err := collection.Update(bson.M{"name": Name}, bson.M{"$inc": bson.M{"postscounts": 1}})
	return err
}

// FindAllCategoryName  从数据库中获取所有类别的Name值
func (c *Category) FindAllCategoryName() (res []interface{}, err error) {
	collection := GetCollection()
	err = collection.Find(bson.M{}).Select(bson.M{"name": 1}).All(&res)
	return
}

// FindByIDhexAndDelete 根据idhex查找，如果没有找到或者其他错误会返回err
func (c *Category) FindByIDhexAndDelete(idhex string) error {
	objid := bson.ObjectIdHex(idhex)
	err := c.DeleteByID(objid)
	// collection.RemoveID()
	return err
}

// DeleteByID 根据id删除数据
func (c *Category) DeleteByID(id bson.ObjectId) (err error) {
	collection := GetCollection()
	// log.Println(id)
	err = collection.RemoveId(id)
	return
}

// FindOneCategoryInfoByIDhex 传入idhex，查找数据库获得对应id的信息，如果出错或者找不到，返回err
func (c *Category) FindOneCategoryInfoByIDhex(idhex string) error {
	objid := bson.ObjectIdHex(idhex)
	collection := GetCollection()
	err := collection.FindId(objid).One(c)
	return err
}

// UpdateByIDhex 传入idhex，查找数据库并修改对应的值
func (c *Category) UpdateByIDhex(idhex string) error {
	objid := bson.ObjectIdHex(idhex)
	collection := GetCollection()
	err := collection.UpdateId(objid, bson.M{"$set": bson.M{"name": c.Name, "englishname": c.EnglishName, "description": c.Description}})
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
	collection = database.C("category")
}

// GetDatabase 获取数据库对象
func GetDatabase() *mgo.Database {
	return database
}

// GetCollection 获取集合对象
func GetCollection() *mgo.Collection {
	return collection
}
