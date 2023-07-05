package conf

import (
	"encoding/json"
	"io/ioutil"
	"log"

	"goblog/models/user"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func Main() {

	confbytes, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Fatalln("read config file error", err)
	}
	newuser := &user.User{}
	err = json.Unmarshal(confbytes, newuser)
	if err != nil {
		log.Fatalln("json config file error", err)
	}
	log.Println(newuser)
	if user.Exist(newuser) {
		log.Println("user already exist")
	} else {
		newuser.ID = primitive.NewObjectID()
		err = user.InsertUser(newuser)
		if err != nil {
			log.Fatalln("user insert error", err)
		}
	}
	log.Println("add user succeed")

}
