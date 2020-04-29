package main

import (
	"os"
	"os/signal"
	"syscall"

	animation "github.com/jirikopecky/saturn-v-blastoff/blastoff-animation/engine"
	log "github.com/sirupsen/logrus"
)

const (
	brightness = 60
	ledCounts  = 60
)

func init() {
	log.SetLevel(log.TraceLevel)
}

func handleTermination(done chan bool, engine *animation.AnimationEngine) {
	signals := make(chan os.Signal, 1)

	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	// wait for signal
	sig := <-signals
	log.WithField("signal", sig).Debug("Signal received")

	engine.CleanAndDestroy()
	done <- true
}

func main() {
	engine, err := animation.InitAnimationEngine(brightness, ledCounts)
	if err != nil {
		log.Panic("Cannot initialize LED stripe driver!")
	}

	engine.Setup()

	done := make(chan bool, 1)
	go handleTermination(done, engine)

	go engine.StartAnimation()

	// start animation
	log.Info("Started")
	<-done
	log.Info("Exiting")
}
