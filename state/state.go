package main

import (
	"bytes"
	"encoding/gob"
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/Shopify/sarama"
	"github.com/wvanbergen/kafka/consumergroup"
	kazoo "github.com/wvanbergen/kazoo-go"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var (
	kafkaConn     = flag.String("kafka-host", "", "Kafka connection string (e.g. `192.168.10.100:9092`)")
	mongoConn     = flag.String("mongo-host", "", "Mongo connection string (e.g. `192.168.10.100:27017`)")
	consumerGroup = flag.String("kafka-consumer-group", "group.testing", "The name of the consumer group, used for coordination and load balancing")
	zookeeper     = flag.String("kafka-zookeeper-host", "", "A comma-separated Zookeeper connection string (e.g. `zookeeper1.local:2181,zookeeper2.local:2181,zookeeper3.local:2181`)")

	zookeeperNodes []string
)

type Bot struct {
	Uuid        string
	Status      string
	Started     string
	Finished    string
	Strategy    string
	Market      string
	Config      string
	Performance string
	Actions     []string
}

func main() {
	sarama.Logger = log.New(os.Stdout, "[Sarama] ", log.LstdFlags)

	flag.Parse()
	log.SetFlags(0)

	config := consumergroup.NewConfig()
	config.Offsets.Initial = sarama.OffsetOldest
	config.Offsets.ProcessingTimeout = 10 * time.Second

	zookeeperNodes, config.Zookeeper.Chroot = kazoo.ParseConnectionString(*zookeeper)

	log.Println("zookeeper", *zookeeper)
	consumer, consumerErr := consumergroup.JoinConsumerGroup(*consumerGroup, []string{"platform", "bot", "market"}, zookeeperNodes, config)
	if consumerErr != nil {
		log.Fatalln(consumerErr)
	}

	sessionMongo, err := mgo.Dial(*mongoConn)
	if err != nil {
		log.Fatalln("Mongo Error", err)
	}
	defer sessionMongo.Close()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		if err := consumer.Close(); err != nil {
			sarama.Logger.Println("Error closing the consumer", err)
		}
	}()

	log.Println("Listening to kafka events has started...")
	err = nil
	for message := range consumer.Messages() {

		if message.Topic == "bot" {

			// decode
			dec := gob.NewDecoder(bytes.NewBuffer(message.Value))
			var data []string
			dec.Decode(&data)

			if data[1] == "start" {

				_, err = sessionMongo.DB("sf-trading-bot").C("bot").Upsert(
					bson.M{"Uuid": data[0]},
					bson.M{
						"Uuid":    data[0],
						"Status":  "started",
						"Actions": []string{},
						"Started": data[2],
						"Market":  data[3],
						"Config":  data[4],
					},
				)

			} else if data[1] == "stop" {

				_, err = sessionMongo.DB("sf-trading-bot").C("bot").Upsert(bson.M{"Uuid": data[0]}, bson.M{"$set": bson.M{"Uuid": data[0], "Status": "finished"}})

			} else if data[1] == "market_buy" {

				// bot := Bot{}
				// err = sessionMongo.DB("sf-trading-bot").C("bot").Find(bson.M{"Uuid": data[0]}).One(&bot)
				_, err = sessionMongo.DB("sf-trading-bot").C("bot").Upsert(
					bson.M{"Uuid": data[0]},
					bson.M{"$push": bson.M{"Actions": strings.Join(data[1:], ",")}},
				)

			} else if data[1] == "market_sell" {

				// bot := Bot{}
				// err = sessionMongo.DB("sf-trading-bot").C("bot").Find(bson.M{"Uuid": data[0]}).One(&bot)
				_, err = sessionMongo.DB("sf-trading-bot").C("bot").Upsert(
					bson.M{"Uuid": data[0]},
					bson.M{"$push": bson.M{"Actions": strings.Join(data[1:], ",")}},
				)

			}
			if err != nil {
				log.Println("Error:", err)
			}
		}
		// consumer.CommitUpto(message)
	}
}

/*
		result := Bot{}
		err := sessionMongo.DB("sf-trading-bot").C("bot").Upsert(bson.M{"Uuid": data[0]}, bson.M{"Uuid": data[0], "Status": "started"})
		err = c.Find(bson.M{"Uuid": data[0]}).One(&result)

		if err == mgo.ErrNotFound {
			err = c.Insert(&Bot{data[0], "started", ""})
			if err != nil {
				log.Fatal(err)
			}
		} else if err != nil {
			log.Println("Error:", err)
		} else {
			c.Update(bson.M{"Uuid": data[0]}, bson.M{"Status": "started"})
		}
	}
*/
