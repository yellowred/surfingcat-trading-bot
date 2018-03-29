package mongo

import (
	"encoding/json"
	"flag"
	"io"
	"log"
	"os"
	"os/signal"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
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

type MongoStateStorage struct {
	sessionMongo *mgo.Session
}

func NewMongoStateStorage(mongoConn string, debug bool) *MongoStateStorage {
	flag.Parse()
	log.SetFlags(0)

	if debug {
		mgo.SetDebug(true)
		mgo.SetLogger(log.New(os.Stdout, "[MGO] ", log.LstdFlags))
	}

	sessionMongo, err := mgo.Dial(mongoConn)
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

	return &MongoStateStorage{sessionMongo}
}

func (s *MongoStateStorage) Bots() []Bot {
	bots := []Bot{}
	s.sessionMongo.DB("sf-trading-bot").C("bot").Find(bson.M{}).All(&bots)
	return bots
}

func (s *MongoStateStorage) FindUser(login string) User {
	user := User{}
	s.sessionMongo.DB("sf-trading-bot").C("user").Find(bson.M{"Login": login}).One(&user)
	return user
}

func (s *MongoStateStorage) NewUserFromJson(dataStream io.Reader) User {
	decoder := json.NewDecoder(dataStream)
	jsondata := User{}
	_ = decoder.Decode(&jsondata)
	return jsondata
}

var signingKey = []byte("x-sign-key")

// GetToken create a jwt token with user claims
func (s *MongoStateStorage) GetToken(user User) string {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["uuid"] = user.Uuid
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()
	signedToken, _ := token.SignedString(signingKey)
	return signedToken
}

// GetJSONToken create a JSON token string
func (s *MongoStateStorage) GetJSONToken(user User) string {
	token := s.GetToken(user)
	jsontoken := "{\"id_token\": \"" + token + "\"}"
	return jsontoken
}
