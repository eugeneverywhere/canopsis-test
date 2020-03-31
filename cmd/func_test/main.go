package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/eugeneverywhere/canopsis-test/config"
	"github.com/eugeneverywhere/canopsis-test/rabbit"
	"github.com/eugeneverywhere/canopsis-test/types"
	"github.com/lillilli/logger"
	"github.com/lillilli/vconf"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	configFile = flag.String("config", "", "set service config file")
	sender     rabbit.Sender
)

func main() {
	flag.Parse()

	cfg := &config.Config{}

	if err := vconf.InitFromFile(*configFile, cfg); err != nil {
		fmt.Printf("unable to load config: %s\n", err)
		os.Exit(1)
	}

	logger.Init(cfg.Log)
	log := logger.NewLogger("alarm service tester")

	rabbitConnection, err := rabbit.NewConnection(cfg.Rabbit.Addr)
	if err != nil {
		log.Errorf("Rabbit connecting failed: %v", err)
		os.Exit(1)
	}

	inputChannel, err := rabbitConnection.DeclareQueue(cfg.Rabbit.InputChannel)
	if err != nil {
		log.Errorf("Rabbit connecting to input queue failed: %v", err)
		os.Exit(1)
	}

	sender = rabbit.NewSender(rabbitConnection, inputChannel)
	log.Info("Sending messages...")

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Info("Listening stopped")
				return
			default:
				sendSomething()
			}
		}
	}()
	<-signals
	close(signals)
	cancel()
}

const delay = time.Millisecond * 2000

func sendSomething() {
	time.Sleep(delay)
	rand.Seed(time.Now().UnixNano())
	critLevel := rand.Intn(4)
	component := rand.Intn(3)
	msg := rand.Intn(3)

	event := &types.Event{
		Source:    "testSource",
		Component: fmt.Sprintf("component_%v", component),
		Resource:  "testResource",
		Critical:  critLevel,
		Message:   fmt.Sprintf("message_%v", msg),
		Timestamp: time.Now().Unix(),
	}
	fmt.Print(event)
	sender.Send(event)
}
