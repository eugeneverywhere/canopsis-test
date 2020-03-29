package main

import (
	"canopsis/config"
	"cmc/environment/rabbit"
	"context"
	"flag"
	"fmt"
	"github.com/lillilli/logger"
	"github.com/lillilli/vconf"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	configFile = flag.String("config", "", "set service config file")
)

func main() {
	flag.Parse()

	cfg := &config.Config{}

	if err := vconf.InitFromFile(*configFile, cfg); err != nil {
		fmt.Printf("unable to load config: %s\n", err)
		os.Exit(1)
	}

	logger.Init(cfg.Log)
	log := logger.NewLogger("alarm_monitoring service")

	inputSubscriber, err := initRabbit(cfg)
	if err != nil {
		log.Errorf("Rabbit subscribe to queue failed: %v", err)
		os.Exit(1)
	}

	db, err := initMongoDB(cfg)
	if err != nil {
		log.Errorf("MongoDB init failed: %v", err)
		os.Exit(1)
	}

	startProcessing(log, inputSubscriber, db)
}

func initMongoDB(cfg *config.Config) (*mongo.Client, error) {
	dbUri := fmt.Sprintf("mongodb://%s:%s", cfg.DB.Host, cfg.DB.Port)
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(dbUri))
	if err != nil {
		return nil, err
	}
	ctx, _ = context.WithTimeout(context.Background(), 2*time.Second)
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, err
	}
	return client, nil
}

func initRabbit(cfg *config.Config) (subscriber rabbit.QueueSubscriber, err error) {
	rabbitConnection, err := rabbit.NewConnection(cfg.Rabbit.Addr)
	if err != nil {
		return nil, err
	}

	inputChannel, err := rabbitConnection.DeclareQueue(cfg.Rabbit.InputChannel)
	if err != nil {
		return nil, err
	}

	subscriber, err = rabbit.NewQueueSubscriber(rabbitConnection, inputChannel)
	return subscriber, err
}

func startProcessing(log logger.Logger,
	subscriber rabbit.QueueSubscriber,
	client *mongo.Client) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	inputChannel, err := subscriber.SubscribeOnQueue()
	if err != nil {
		log.Errorf("Failed to subscribe on input queue: %v", err)
		os.Exit(1)
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Info("Listening stopped")
				return
			case data := <-inputChannel:
				fmt.Println(data)
				//go dispatcher.Dispatch(data.Body)
			}
		}
	}()

	<-signals
	close(signals)
	cancel()
}
