package utils

import (
	"flag"
	"github.com/Shopify/sarama"
	"github.com/wvanbergen/kafka/consumergroup" // consumer groups currently in a separate package
	"github.com/wvanbergen/kazoo-go"
    "log"
	"os"
	"os/signal"
	"time"
	// "encoding/json"
	"strings"
)

var Logger LoggerInterface


type KafkaMessage struct {
	Message string
	Time time.Time
}


type LoggerInterface interface {
	PlatformLogger(message []string)
	BotLogger(botId string, message []string)
	MarketLogger(message []string)
}

type KafkaLogger struct {
	producer sarama.AsyncProducer
}


var (
	kafkaConn      = flag.String("kafka-host", "", "A comma-separated Zookeeper connection string (e.g. `zookeeper1.local:2181,zookeeper2.local:2181,zookeeper3.local:2181`)")
	consumerGroup  = flag.String("kafka-consumer-group", "group.testing", "The name of the consumer group, used for coordination and load balancing")
	zookeeper      = flag.String("kafka-zookeeper-host", "", "A comma-separated Zookeeper connection string (e.g. `zookeeper1.local:2181,zookeeper2.local:2181,zookeeper3.local:2181`)")

	zookeeperNodes []string
)


func init() {
	sarama.Logger = log.New(os.Stdout, "[Sarama] ", log.LstdFlags)
	flag.Parse()

	if *kafkaConn == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	config := sarama.NewConfig()
	config.ClientID = "sf-trading-bot-server"
	config.Producer.RequiredAcks = sarama.WaitForLocal
	config.Producer.Compression = sarama.CompressionNone
	producer, err := sarama.NewAsyncProducer([]string{*kafkaConn}, config)

	if err != nil {
		panic(err)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, os.Kill)

	go func() {
		<-c
		if err := producer.Close(); err != nil {
			log.Fatal("Error closing async producer", err)
		}

		log.Println("Async Producer closed")
		os.Exit(1)
	}()

	go func() {
		for err := range producer.Errors() {
			log.Println("Failed to write message to topic:", err)
		}
	}()


	Logger = &KafkaLogger{producer}
}



func (l *KafkaLogger) PlatformLogger(message []string) {
	msg := &sarama.ProducerMessage{
		Topic: "platform",
		Timestamp: time.Now(),
		Value: sarama.ByteEncoder(strings.Join(message, ",")),
	}
	l.producer.Input() <- msg
}

func (l *KafkaLogger) BotLogger(botId string, message []string) {
	msg := &sarama.ProducerMessage{
		Topic: "bot",
		Key: sarama.StringEncoder(botId),
		Timestamp: time.Now(),
		Value: sarama.ByteEncoder(strings.Join(message, ",")),
	}
	l.producer.Input() <- msg
}

func (l *KafkaLogger) MarketLogger(message []string) {
	msg := &sarama.ProducerMessage{
		Topic: "market",
		Timestamp: time.Now(),
		Value: sarama.ByteEncoder(strings.Join(message, ",")),
	}
	l.producer.Input() <- msg
}


/*
func ConsumeMessages(topic string) (messages []sarama.ConsumerMessage) {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true

	// Create new consumer
	consumer, err := sarama.NewConsumer([]string{*kafkaConn}, config)

	if err != nil {
		panic(err)
	}

	defer func() {
		if err := consumer.Close(); err != nil {
			panic(err)
		}
	}()


	partitionList, _ := consumer.Partitions(topic) //get all partitions
	messagesChan := make(chan *sarama.ConsumerMessage, 256)
	initialOffset := sarama.OffsetOldest //offset to start reading message from
	for _, partition := range partitionList {  
		pc, _ := consumer.ConsumePartition(topic, partition, initialOffset)
		go func(pc sarama.PartitionConsumer) {
			for message := range pc.Messages() {
				messagesChan <- message
			}
		}(pc)
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	// Count how many message processed
	msgCount := 0

	// Get signnal for finish
	doneCh := make(chan struct{})
	go func() {
		for {
			select {
			case msg := <-messagesChan:
				msgCount++
				messages = append(messages, *msg)
			case <-signals:
				log.Println("Interrupt is detected")
				doneCh <- struct{}{}
			}
		}
	}()

	<-doneCh
	fmt.Println("Processed", msgCount, "messages")
}
*/

// @see https://github.com/wvanbergen/kafka/blob/master/examples/consumergroup/main.go
func ConsumeMessages() (chan sarama.ConsumerMessage) {
	config := consumergroup.NewConfig()
	config.Offsets.Initial = sarama.OffsetOldest
	config.Offsets.ProcessingTimeout = 10 * time.Second

	zookeeperNodes, config.Zookeeper.Chroot = kazoo.ParseConnectionString(*zookeeper)

	consumer, consumerErr := consumergroup.JoinConsumerGroup(*consumerGroup, []string{"platform", "bot", "market"}, zookeeperNodes, config)
	if consumerErr != nil {
		log.Fatalln(consumerErr)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		if err := consumer.Close(); err != nil {
			sarama.Logger.Println("Error closing the consumer", err)
		}
	}()


	log.Println("Consume...")
	messages := make(chan sarama.ConsumerMessage)
	
	go func() {
		for message := range consumer.Messages() {
			log.Println(message)
			messages <- *message
			consumer.CommitUpto(message)
		}
	}()
	return messages
}