// Package session 使用monggodb进行session存储
// 使用方法
// sess:=session.New() 新建一个session结构体
// sess=sess.StartSession()  从数据库中获取数据到session对象
// sess.Set()   sess.Get()
// sess.Save()   设置完在之后，使用Save方法持久化保存到数据库里
package session

import (
	"crypto/rand"
	"encoding/base64"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// Session Session的结构体
type Session struct {
	CookieName  string `bson:"cookiename"`
	SessionID   string `bson:"sessionid"`
	Maxlifetime int    `bson:"maxlifetime"`
	// 最大生存时间，时间为秒
	// Data        map[string]interface{} `bson:"Data"`
	// Data bson.M `bson:"data,omitempty"`
	Data map[string]interface{} `bson:"data"`
}

var (
	// 数据库会话，数据库，以及集合
	session    *mgo.Session
	database   *mgo.Database
	collection *mgo.Collection
)

// GenSessionID 生成一个SessionID
func (s *Session) GenSessionID() string {
	b := make([]byte, 64)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)
}

// SessionStart ,获取session，如果不能从cookie中获取SessionID就新创建一个session并加入mongodb同时设置cookie
func (s *Session) SessionStart(w http.ResponseWriter, r *http.Request) (session *Session) {
	cookie, err := r.Cookie(s.CookieName)
	log.Println("SessionStart cookie:", cookie.Value)
	// log.Println("SessionStart sid:", s.SessionID)
	// 如果cookie中找不到SessionID的值，说明客户端没有sessionID或者cookie被删除，新建一个session并加入到cookie中
	if err != nil || cookie.Value == "" {
		sid := s.SessionID
		// log.Println(1)
		// url.QueryEscape()是将字符串转换成url编码
		// 创建一个新的SessionID后，我们把它加入到mongodb中
		cookie := http.Cookie{Name: s.CookieName, Value: url.QueryEscape(sid), Path: "/", HttpOnly: true, MaxAge: s.Maxlifetime}
		s.insertNewSession(s)
		http.SetCookie(w, &cookie)
	} else {
		// log.Println(2)
		sid, _ := url.QueryUnescape(cookie.Value)
		err := s.GetAllData(sid)
		log.Println("SessionStart sid:", s.SessionID)
		session = s
		// 如果数据获取出错，说明本地的session数据被删除了，或者客户端伪造了sessionID
		// 此时我们需要重设session，因为登录会设置signin标志，所以不会影响到登录部分的安全性。伪造的情况一样重设逻辑，但是普通的session没有signin标记并没什么用。
		if err != nil {
			// log.Println(3)
			log.Println("get session Data error", err)
			sess := New()
			cookie := http.Cookie{Name: sess.CookieName, Value: url.QueryEscape(sess.SessionID), Path: "/", HttpOnly: true, MaxAge: sess.Maxlifetime}
			s.insertNewSession(sess)
			http.SetCookie(w, &cookie)
			session = sess
		}
	}
	return
}

// Init 初始化一个session
func (s *Session) Init() (session *Session) {
	sid := s.GenSessionID()
	data := make(map[string]interface{})
	session = &Session{CookieName: "mod", SessionID: sid, Maxlifetime: 60 * 60 * 24 * 7, Data: data}
	return
}

// New 返回一个初始化好的session
func New() *Session {
	return new(Session).Init()
}

func (s *Session) insertNewSession(session *Session) {
	c := GetCollection()
	err := c.Insert(session)
	if err != nil {
		log.Println("insert session error", err)
	}
}

// GetAllData 从monggodb中获取所有session信息
func (s *Session) GetAllData(SessionID string) (err error) {
	c := GetCollection()
	err = c.Find(bson.M{"sessionid": SessionID}).One(s)
	return
}

// GetAllData 从monggodb中获取所有session信息,返回errnotfound说明查找不到对应的session信息，可能没有创建也可能过期被删除了
func GetAllData(SessionID string) (session Session, err error) {
	c := GetCollection()
	err = c.Find(bson.M{"sessionid": SessionID}).One(&session)
	return
}

// Set 设置session键值
func (s *Session) Set(key string, value interface{}) {
	log.Println(s)
	log.Println("set", s.Data)
	s.Data[key] = value
}

// SetAndSave 设置session键值,并且持久化到数据库
func (s *Session) SetAndSave(key string, value interface{}) {
	s.Data[key] = value
	s.Save()
}

// Get 获取session某个键的值
func (s *Session) Get(key string) interface{} {
	return s.Data[key]
}

// Save 设置完session的Data值后要保存到数据库里
func (s *Session) Save() (err error) {
	c := GetCollection()
	err = c.Update(bson.M{"sessionid": s.SessionID}, bson.M{"$set": bson.M{"data": s.Data}})
	return
}

// Destroy 从数据库中删除session
func (s *Session) Destroy() {
	c := GetCollection()
	c.Remove(bson.M{"sessionid": s.SessionID})
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
		log.Panicln("dial Database failed", err)
	}
	database = session.DB("goblog")
	collection = database.C("session")
}

// GetDatabase 获取数据库对象
func GetDatabase() *mgo.Database {
	return database
}

// GetCollection 获取集合对象
func GetCollection() *mgo.Collection {
	return collection
}

/*func main() {

}
*/
