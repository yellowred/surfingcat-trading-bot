package utils

import (
	"flag"
	"log"
	"os"
	"os/signal"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type Bot struct {
	Uuid        string   `json:"Uuid" bson:"Uuid"`
	Status      string   `json:"Status" bson:"Status"`
	Started     string   `json:"Started" bson:"Started"`
	Finished    string   `json:"Finished" bson:"Finished"`
	Strategy    string   `json:"Strategy" bson:"Strategy"`
	Market      string   `json:"Market" bson:"Market"`
	Config      string   `json:"Config" bson:"Config"`
	Performance string   `json:"Performance" bson:"Performance"`
	Actions     []string `json:"Actions" bson:"Actions"`
}

// UserJSON - json data expected for login/signup
type User struct {
	Uuid     string `json:"Uuid" bson:"Uuid"`
	Login    string `json:"Login" bson:"Login"`
	Password string `json:"Password" bson:"Password"`
}

var (
	mongoConn    = flag.String("mongo-host", "", "A comma-separated MongoDB connection string (e.g. `192.168.10.100:27017`)")
	sessionMongo *mgo.Session
)

func init() {
	flag.Parse()
	log.SetFlags(0)

	var err error
	sessionMongo, err = mgo.Dial(*mongoConn)
	if err != nil {
		log.Fatalln("Mongo Error", err)
	}
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, os.Kill)

	go func() {
		<-c
		sessionMongo.Close()
		log.Println("MongoDb session closed")
		os.Exit(1)
	}()
}

func Bots() []Bot {
	bots := []Bot{}
	sessionMongo.DB("sf-trading-bot").C("bot").Find(bson.M{}).All(&bots)
	return bots
}

func FindUser(login string) User {
	user := User{}
	sessionMongo.DB("sf-trading-bot").C("user").Find(bson.M{"Login": login}).One(&user)
	return user
}
