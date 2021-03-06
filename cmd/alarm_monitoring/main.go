package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/eugeneverywhere/canopsis-test/config"
	"github.com/eugeneverywhere/canopsis-test/db"
	"github.com/eugeneverywhere/canopsis-test/handler"
	"github.com/eugeneverywhere/canopsis-test/rabbit"
	"github.com/lillilli/logger"
	"github.com/lillilli/vconf"
	"os"
	"os/signal"
	"syscall"
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
	defer db.Close()

	startProcessing(log, inputSubscriber, handler.New(db))
}

func initMongoDB(cfg *config.Config) (db.DB, error) {
	db := db.New(fmt.Sprintf("%s:%v", cfg.DB.Host, cfg.DB.Port), cfg.DB.Name)
	err := db.Connect()
	if err != nil {
		return nil, err
	}
	return db, nil
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
	handler handler.AlarmHandler) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	inputChannel, err := subscriber.SubscribeOnQueue()
	if err != nil {
		log.Errorf("Failed to subscribe on input queue: %v", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Info("Listening stopped")
				return
			case data := <-inputChannel:
				go handler.HandleMsg(data.Body)
			}
		}
	}()

	<-signals
	close(signals)
	cancel()
}
