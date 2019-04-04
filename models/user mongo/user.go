package main

import (
	"fmt"

	"go.mongodb.org/mongo-driver/bson"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// User 用户表的一行信息 包括用户名 密码
// 暂时先使用明文存储密码
type User struct {
	ID       bson.ObjectId `bson:"_id"`
	Username string        `bson:"username"`
	Password string        `bson:"password"`
}

func main() {
	fmt.Println("ok1")
	// 创建一个client对象
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		panic(err)
	}
	// 建立连接
	ctx,cancel:=context.WithTimeout(context.Background(),20*time.Second)
	defer cancel()
	err=client.Connect(ctx)
	if err!=nil{
		panic(err)
	}
	// 获取操作的集合
	collection := client.Database("goblog").Collection("user")
	fmt.Println("ok")
	ctx, _ = context.WithTimeout(context.Background(), 5*time.Second)
	res, err := collection.InsertOne(ctx, &User{ID: , Username: "abc2", Password: "123456"})
	id := res.InsertedID
	fmt.Println("insert succeed", id)

}
