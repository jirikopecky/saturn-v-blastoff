package main

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/introspect"
	log "github.com/sirupsen/logrus"
)

const (
	brightness = 60
	ledCounts  = 60
)

const dbusIntrospect = `
<node>
	<interface name="com.github.jirikopecky.SaturnV.BlastOff">
		<method name="Start" />
		<method name="Stop" />
	</interface>` + introspect.IntrospectDataString + `</node> `

type animationApp struct {
	engine *AnimationEngine
	mux    *sync.Mutex
}

func init() {
	log.SetLevel(log.TraceLevel)
}

func (app *animationApp) Start() *dbus.Error {
	app.mux.Lock()
	defer app.mux.Unlock()

	if app.engine != nil {
		log.Trace("Start called but animation already running.")
		return nil
	}

	engine, err := InitAnimationEngine(brightness, ledCounts)
	if err != nil {
		log.WithError(err).Error("Cannot initialize LED stripe driver!")
		return dbus.MakeFailedError(err)
	}

	err = engine.Setup()
	if err != nil {
		log.WithError(err).Error("Cannot initialize animation engine!")
		return dbus.MakeFailedError(err)
	}

	go engine.StartAnimation()

	app.engine = engine
	return nil
}

func (app *animationApp) Stop() *dbus.Error {
	app.mux.Lock()
	defer app.mux.Unlock()

	if app.engine == nil {
		log.Trace("Stop called but animation is not running.")
		return nil
	}

	err := app.engine.CleanAndDestroy()
	if err != nil {
		log.WithError(err).Error("Error when cleaning up the animation!")
		return dbus.MakeFailedError(err)
	}

	app.engine = nil
	return nil
}

func handleTermination(done chan bool) {
	signals := make(chan os.Signal, 1)

	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	// wait for signal
	sig := <-signals
	log.WithField("signal", sig).Debug("Signal received")

	done <- true
}

func main() {
	app := &animationApp{
		engine: nil,
		mux:    &sync.Mutex{},
	}
	defer app.Stop()

	conn, err := dbus.SystemBus()
	if err != nil {
		log.WithError(err).Panic("Cannot connect to DBus System bus")
	}
	defer conn.Close()

	done := make(chan bool, 1)
	go handleTermination(done)

	conn.Export(app, "/com/github/jirikopecky/SaturnV", "com.github.jirikopecky.SaturnV.BlastOff")
	conn.Export(introspect.Introspectable(dbusIntrospect),
		"/com/github/jirikopecky/SaturnV", "org.freedesktop.DBus.Introspectable")

	reply, err := conn.RequestName("com.github.jirikopecky.SaturnV", dbus.NameFlagDoNotQueue)
	if err != nil {
		log.WithError(err).Panic("Failed to call DBus RequestName method")
	}
	if reply != dbus.RequestNameReplyPrimaryOwner {
		log.Fatal("DBus name already taken")
	}

	log.Info("Started")
	<-done
	log.Info("Exiting")
}
